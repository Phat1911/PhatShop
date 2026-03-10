package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/models"
	"phatshop-backend/internal/repository"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	products   *repository.ProductRepo
	categories *repository.CategoryRepo
	orders     *repository.OrderRepo
	users      *repository.UserRepo
	cfg        *config.Config
}

func NewAdminHandler(
	products *repository.ProductRepo,
	categories *repository.CategoryRepo,
	orders *repository.OrderRepo,
	users *repository.UserRepo,
	cfg *config.Config,
) *AdminHandler {
	return &AdminHandler{
		products:   products,
		categories: categories,
		orders:     orders,
		users:      users,
		cfg:        cfg,
	}
}

var slugRe = regexp.MustCompile(`[^a-z0-9-]`)

func toSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = slugRe.ReplaceAllString(s, "")
	return s
}

func safeName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(name, "")
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	sellerID, _ := c.Get("user_id")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadSize)

	title := c.PostForm("title")
	description := c.PostForm("description")
	productType := c.PostForm("product_type")
	priceStr := c.PostForm("price")
	categoryID := c.PostForm("category_id")
	tagsStr := c.PostForm("tags")

	price, _ := strconv.ParseInt(priceStr, 10, 64)
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			if tag := strings.TrimSpace(t); tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	productID := uuid.New().String()

	productFile, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product file is required"})
		return
	}
	defer productFile.Close()

	fileName := safeName(header.Filename)
	storageDir := filepath.Join(h.cfg.StorageDir, "products", productID)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create storage dir"})
		return
	}
	filePath := filepath.Join(storageDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	fileSize, _ := io.Copy(dst, productFile)
	dst.Close()

	var thumbnailURL string
	if thumbFile, thumbHeader, err := c.Request.FormFile("thumbnail"); err == nil {
		defer thumbFile.Close()
		thumbName := productID + filepath.Ext(thumbHeader.Filename)
		thumbPath := filepath.Join(h.cfg.UploadDir, "thumbnails", thumbName)
		if out, err := os.Create(thumbPath); err == nil {
			io.Copy(out, thumbFile)
			out.Close()
			thumbnailURL = "/uploads/thumbnails/" + thumbName
		}
	}

	var previewURLs []string
	if c.Request.MultipartForm != nil {
		for i, fh := range c.Request.MultipartForm.File["previews"] {
			if i >= 5 {
				break
			}
			previewFile, _ := fh.Open()
			previewName := fmt.Sprintf("%s_%d%s", productID, i, filepath.Ext(fh.Filename))
			previewPath := filepath.Join(h.cfg.UploadDir, "previews", previewName)
			if out, err := os.Create(previewPath); err == nil {
				io.Copy(out, previewFile)
				out.Close()
				previewURLs = append(previewURLs, "/uploads/previews/"+previewName)
			}
			previewFile.Close()
		}
	}

	var catIDPtr *string
	if categoryID != "" {
		catIDPtr = &categoryID
	}

	slug := toSlug(title) + "-" + productID[:8]

	// Trailer: uploaded video file
	var trailerURL string
	if trailerFile, trailerHeader, err := c.Request.FormFile("trailer"); err == nil {
		defer trailerFile.Close()
		trailerDir := filepath.Join(h.cfg.UploadDir, "trailers")
		if err := os.MkdirAll(trailerDir, 0755); err == nil {
			trailerName := productID + filepath.Ext(trailerHeader.Filename)
			trailerPath := filepath.Join(trailerDir, trailerName)
			if out, err := os.Create(trailerPath); err == nil {
				io.Copy(out, trailerFile)
				out.Close()
				trailerURL = "/uploads/trailers/" + trailerName
			}
		}
	}

	product := &models.Product{
		ID:           productID,
		SellerID:     sellerID.(string),
		CategoryID:   catIDPtr,
		Title:        title,
		Slug:         slug,
		Description:  description,
		ProductType:  productType,
		Price:        price,
		ThumbnailURL: thumbnailURL,
		PreviewURLs:  previewURLs,
		FilePath:     filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		Tags:         tags,
		TrailerURL:   trailerURL,
	}

	created, err := h.products.Create(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := repository.ProductFilter{
		ProductType: c.Query("type"),
		Search:      c.Query("search"),
		Page:        page,
		Limit:       limit,
		AdminView:   true,
	}
	products, total, err := h.products.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{"data": products, "total": total, "page": page, "total_pages": totalPages})
}

func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	if err := h.products.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AdminHandler) PublishProduct(c *gin.Context) {
	var body struct {
		IsPublished bool `json:"is_published"`
	}
	c.ShouldBindJSON(&body)
	if err := h.products.SetPublished(c.Request.Context(), c.Param("id"), body.IsPublished); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.categories.Create(c.Request.Context(), req.Name, req.Slug, req.ProductType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.categories.Update(c.Request.Context(), c.Param("id"), req.Name, req.Slug, req.ProductType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	if err := h.categories.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AdminHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit
	orders, total, err := h.orders.ListAll(c.Request.Context(), c.Query("status"), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{"data": orders, "total": total, "page": page, "total_pages": totalPages})
}

func (h *AdminHandler) GetOrder(c *gin.Context) {
	order, err := h.orders.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID := c.Param("id")
	var err error

	if body.Status == "paid" {
		// Use MarkPaidByAdmin so purchases table is populated and buyers can download
		err = h.orders.MarkPaidByAdmin(c.Request.Context(), orderID)
	} else {
		err = h.orders.UpdateStatus(c.Request.Context(), orderID, body.Status)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit
	users, total, err := h.users.ListAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{"data": users, "total": total, "page": page, "total_pages": totalPages})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	var body struct {
		Role string `json:"role" binding:"required,oneof=user admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.users.UpdateRole(c.Request.Context(), c.Param("id"), body.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.orders.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

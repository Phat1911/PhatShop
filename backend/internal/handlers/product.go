package handlers

import (
	"net/http"
	"phatshop-backend/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	products   *repository.ProductRepo
	categories *repository.CategoryRepo
}

func NewProductHandler(products *repository.ProductRepo, categories *repository.CategoryRepo) *ProductHandler {
	return &ProductHandler{products: products, categories: categories}
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := repository.ProductFilter{
		ProductType: c.Query("type"),
		CategoryID:  c.Query("category"),
		Search:      c.Query("search"),
		Sort:        c.Query("sort"),
		Page:        page,
		Limit:       limit,
	}

	products, total, err := h.products.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"data":        products,
		"total":       total,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := h.products.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	go h.products.IncrementView(c.Request.Context(), id)
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	cats, err := h.categories.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

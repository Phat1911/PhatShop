package handlers

import (
	"net/http"
	"phatshop-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cart     *repository.CartRepo
	products *repository.ProductRepo
}

func NewCartHandler(cart *repository.CartRepo, products *repository.ProductRepo) *CartHandler {
	return &CartHandler{cart: cart, products: products}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	items, total, err := h.cart.GetItems(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var body struct {
		ProductID string `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.products.GetByID(c.Request.Context(), body.ProductID)
	if err != nil || product == nil || !product.IsPublished {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if err := h.cart.AddItem(c.Request.Context(), userID.(string), body.ProductID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "added to cart"})
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID := c.Param("product_id")
	if err := h.cart.RemoveItem(c.Request.Context(), userID.(string), productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed from cart"})
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if err := h.cart.Clear(c.Request.Context(), userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}

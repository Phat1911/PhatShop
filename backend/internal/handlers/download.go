package handlers

import (
	"net/http"
	"os"
	"phatshop-backend/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type DownloadHandler struct {
	downloads *repository.DownloadRepo
}

func NewDownloadHandler(downloads *repository.DownloadRepo) *DownloadHandler {
	return &DownloadHandler{downloads: downloads}
}

func (h *DownloadHandler) RequestToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID := c.Param("product_id")

	has, err := h.downloads.HasPurchased(c.Request.Context(), userID.(string), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusForbidden, gin.H{"error": "you have not purchased this product"})
		return
	}

	token, err := h.downloads.CreateToken(c.Request.Context(), userID.(string), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create download token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token.Token, "expires_at": token.ExpiresAt})
}

func (h *DownloadHandler) CheckPurchase(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID := c.Param("product_id")

	has, err := h.downloads.HasPurchased(c.Request.Context(), userID.(string), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"purchased": has})
}

func (h *DownloadHandler) ServeFile(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	t, err := h.downloads.ValidateAndGetFile(c.Request.Context(), tokenStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if t == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid token"})
		return
	}
	if time.Now().After(t.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "download link expired"})
		return
	}
	if t.UsedCount >= t.MaxUses {
		c.JSON(http.StatusGone, gin.H{"error": "download limit reached"})
		return
	}

	if _, err := os.Stat(t.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file not found on server"})
		return
	}

	h.downloads.IncrementUsage(c.Request.Context(), t.TokenID)
	c.Header("Content-Disposition", `attachment; filename="`+t.FileName+`"`)
	c.File(t.FilePath)
}

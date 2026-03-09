package handlers

import (
	"net/http"
	"phatshop-backend/internal/models"
	"phatshop-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users *repository.UserRepo
}

func NewUserHandler(users *repository.UserRepo) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.users.GetByID(c.Request.Context(), userID.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Update(c.Request.Context(), userID.(string), req.DisplayName, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, user)
}

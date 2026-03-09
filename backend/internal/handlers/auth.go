package handlers

import (
	"net/http"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/models"
	"phatshop-backend/internal/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users *repository.UserRepo
	cfg   *config.Config
}

func NewAuthHandler(users *repository.UserRepo, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.users.Create(c.Request.Context(), req.Username, req.Email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.GetByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	duration, err := time.ParseDuration(h.cfg.JWTExpires)
	if err != nil {
		duration = 168 * time.Hour
	}
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

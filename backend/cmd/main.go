package main

import (
	"context"
	"log"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/db"
	"phatshop-backend/internal/handlers"
	"phatshop-backend/internal/middleware"
	"phatshop-backend/internal/repository"
	"phatshop-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Repositories
	userRepo     := repository.NewUserRepo(database.Pool)
	categoryRepo := repository.NewCategoryRepo(database.Pool)
	productRepo  := repository.NewProductRepo(database.Pool)
	cartRepo     := repository.NewCartRepo(database.Pool)
	orderRepo    := repository.NewOrderRepo(database.Pool)
	downloadRepo := repository.NewDownloadRepo(database.Pool)
	receiptRepo  := repository.NewReceiptRepo(database.Pool)

	// Services
	ocrService := services.NewOCRService(cfg.GeminiAPIKey)

	// Handlers
	authHandler     := handlers.NewAuthHandler(userRepo, cfg)
	userHandler     := handlers.NewUserHandler(userRepo)
	productHandler  := handlers.NewProductHandler(productRepo, categoryRepo)
	cartHandler     := handlers.NewCartHandler(cartRepo, productRepo)
	orderHandler    := handlers.NewOrderHandler(orderRepo, cartRepo)
	paymentHandler  := handlers.NewPaymentHandler(orderRepo, cfg)
	downloadHandler := handlers.NewDownloadHandler(downloadRepo)
	receiptHandler  := handlers.NewReceiptHandler(receiptRepo, orderRepo, ocrService, cfg)
	adminHandler    := handlers.NewAdminHandler(productRepo, categoryRepo, orderRepo, userRepo, cfg)

	r := gin.Default()
	r.Use(middleware.CORS())
	r.Static("/uploads", cfg.UploadDir)

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "app": "PhatShop"})
	})

	api := r.Group("/api/v1")

	// Auth
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Public
	api.GET("/categories", productHandler.ListCategories)
	api.GET("/products", productHandler.ListProducts)
	api.GET("/products/:id", productHandler.GetProduct)

	// VNPay IPN (no auth - called by VNPay server)
	api.GET("/payments/vnpay/ipn", paymentHandler.VNPayIPN)

	// Download file by token (no JWT auth)
	api.GET("/downloads/file", downloadHandler.ServeFile)

	// Protected
	protected := api.Group("")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.GET("/users/me", userHandler.GetMe)
		protected.PUT("/users/me", userHandler.UpdateMe)

		protected.GET("/cart", cartHandler.GetCart)
		protected.POST("/cart", cartHandler.AddToCart)
		protected.DELETE("/cart/:product_id", cartHandler.RemoveFromCart)
		protected.DELETE("/cart", cartHandler.ClearCart)

		protected.POST("/orders", orderHandler.CreateOrder)
		protected.GET("/orders", orderHandler.ListOrders)
		protected.GET("/orders/:id", orderHandler.GetOrder)
		protected.POST("/orders/:id/receipt", receiptHandler.UploadReceipt)

		protected.POST("/payments/vnpay/create", paymentHandler.CreatePaymentURL)

		protected.GET("/downloads/request/:product_id", downloadHandler.RequestToken)
		protected.GET("/downloads/check/:product_id", downloadHandler.CheckPurchase)
	}

	// Admin
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(cfg.JWTSecret), middleware.Admin())
	{
		admin.GET("/stats", adminHandler.GetStats)

		admin.GET("/products", adminHandler.ListProducts)
		admin.POST("/products", adminHandler.CreateProduct)
		admin.DELETE("/products/:id", adminHandler.DeleteProduct)
		admin.PATCH("/products/:id/publish", adminHandler.PublishProduct)

		admin.GET("/categories", productHandler.ListCategories)
		admin.POST("/categories", adminHandler.CreateCategory)
		admin.PUT("/categories/:id", adminHandler.UpdateCategory)
		admin.DELETE("/categories/:id", adminHandler.DeleteCategory)

		admin.GET("/orders", adminHandler.ListOrders)
		admin.GET("/orders/:id", adminHandler.GetOrder)
		admin.PATCH("/orders/:id/status", adminHandler.UpdateOrderStatus)

		admin.GET("/users", adminHandler.ListUsers)
		admin.PATCH("/users/:id/role", adminHandler.UpdateUserRole)
	}

	log.Printf("PhatShop backend running on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}

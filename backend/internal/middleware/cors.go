package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	origins := []string{
		"http://localhost:3000",
		"http://localhost:3001",
	}

	// Add any extra origins from FRONTEND_URL env var
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		for _, u := range strings.Split(frontendURL, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				origins = append(origins, u)
			}
		}
	}

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	}
	return cors.New(config)
}

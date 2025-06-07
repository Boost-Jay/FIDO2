package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	// Get allowed origins from environment variable
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string

	// If not set, use default restrictive setting
	if allowedOriginsStr == "" {
		allowedOrigins = []string{"http://localhost:3000"} // Adjust to your frontend URL
	} else {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			// If no specific origins defined, allow all
			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				return true
			}

			// Otherwise check if origin is allowed
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		},
	})
}
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" || apiKey != cfg.XAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}


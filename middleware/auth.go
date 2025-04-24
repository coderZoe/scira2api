package middleware

import (
	"scira2api/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := config.ApiKey
		if apiKey != "" {
			Key := c.GetHeader("Authorization")
			if Key != "" {
				Key = strings.TrimPrefix(Key, "Bearer ")
				if Key != apiKey {
					c.JSON(401, gin.H{
						"error": "Invalid API key",
					})
					c.Abort()
					return
				} else {
					c.Next()
					return
				}
			} else {
				c.JSON(401, gin.H{
					"error": "Missing or invalid Authorization header",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

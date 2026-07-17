package middleware

import (
	"net/http"

	"go-guestbook/config"

	"github.com/gin-gonic/gin"
)

// DevelopmentOnly allows the request to continue only when the app runs in development.
// cfg carries the loaded application configuration used to decide development mode.
// Non-development requests receive 404 so tools stay hidden outside local work.
func DevelopmentOnly(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.IsDevelopment() {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

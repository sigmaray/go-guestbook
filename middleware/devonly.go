package middleware

import (
	"net/http"

	"go-guestbook/config"

	"github.com/gin-gonic/gin"
)

// DevToolsOnly allows the request to continue only when development tools are enabled.
// cfg carries the loaded application configuration used to decide tools access.
// Requests with tools disabled receive 404 so the UI and endpoints stay hidden by default.
func DevToolsOnly(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.DevToolsEnabled {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	}
}

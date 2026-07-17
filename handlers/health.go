package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health reports application and database readiness for uptime checks.
func (h *Handler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		respondHealth(c, http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		respondHealth(c, http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	respondHealth(c, http.StatusOK, gin.H{"status": "ok"})
}

// respondHealth writes a JSON or HEAD response for health endpoints.
func respondHealth(c *gin.Context, status int, body gin.H) {
	if c.Request.Method == http.MethodHead {
		c.Status(status)
		return
	}
	c.JSON(status, body)
}

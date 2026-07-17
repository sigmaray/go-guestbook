package handlers

import (
	"go-guestbook/config"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Handler groups HTTP dependencies used by route handlers.
type Handler struct {
	DB                 *gorm.DB
	Validate           *validator.Validate
	Logger             zerolog.Logger
	TestAPI            bool
	DevToolsEnabled    bool
	MaxMessagesEnabled bool
	MaxMessages        int
}

// NewHandler constructs a Handler with validation and logging dependencies.
// db is the open GORM connection shared by handlers.
// logger is the structured logger used for operational events.
// cfg supplies feature flags such as the test API, development tools, and message capacity limits.
func NewHandler(db *gorm.DB, logger zerolog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		DB:                 db,
		Validate:           validator.New(),
		Logger:             logger,
		TestAPI:            cfg.TestAPIEnabled,
		DevToolsEnabled:    cfg.DevToolsEnabled,
		MaxMessagesEnabled: cfg.MaxMessagesEnabled,
		MaxMessages:        cfg.MaxMessages,
	}
}

// adminHTML renders an admin HTML template and injects shared layout values.
// c is the Gin request context; code is the HTTP status; name is the template name;
// data is the page-specific template data merged with DevToolsEnabled for the admin nav.
func (h *Handler) adminHTML(c *gin.Context, code int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["DevToolsEnabled"] = h.DevToolsEnabled
	c.HTML(code, name, data)
}

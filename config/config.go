package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds application settings loaded from environment variables.
type Config struct {
	Environment      string `envconfig:"GO_GUESTBOOK_ENVIRONMENT" default:"development"`
	HTTPPort         int    `envconfig:"GO_GUESTBOOK_HTTP_PORT" default:"8084"`
	GinMode          string `envconfig:"GIN_MODE" default:"release"`
	SessionSecret    string `envconfig:"GO_GUESTBOOK_SESSION_SECRET" required:"true"`
	SessionSecure    bool   `envconfig:"GO_GUESTBOOK_SESSION_SECURE" default:"false"`
	DatabaseHost     string `envconfig:"GO_GUESTBOOK_DATABASE_HOST" default:"shared-postgres"`
	DatabasePort     string `envconfig:"GO_GUESTBOOK_DATABASE_PORT" default:"5432"`
	DatabaseName     string `envconfig:"GO_GUESTBOOK_DATABASE_NAME" default:"goguestbook"`
	DatabaseUser     string `envconfig:"GO_GUESTBOOK_DATABASE_USER" default:"goguestbook"`
	DatabasePassword string `envconfig:"GO_GUESTBOOK_DATABASE_PASSWORD" required:"true"`
	TestAPIEnabled   bool   `envconfig:"GO_GUESTBOOK_TEST_API_ENABLED" default:"false"`

	// MaxMessagesEnabled turns on the guestbook capacity limit (default on).
	MaxMessagesEnabled bool `envconfig:"GO_GUESTBOOK_MAX_MESSAGES_ENABLED" default:"true"`
	// MaxMessages is the maximum number of stored messages when the limit is enabled.
	MaxMessages int `envconfig:"GO_GUESTBOOK_MAX_MESSAGES" default:"10000"`

	// LoginRateLimitEnabled turns on login POST rate limiting (default on).
	LoginRateLimitEnabled bool `envconfig:"GO_GUESTBOOK_LOGIN_RATE_LIMIT_ENABLED" default:"true"`
	// LoginRateLimitMaxAttempts is how many login attempts are allowed per IP inside the window.
	LoginRateLimitMaxAttempts int `envconfig:"GO_GUESTBOOK_LOGIN_RATE_LIMIT_MAX_ATTEMPTS" default:"10"`
	// LoginRateLimitWindow is the sliding window duration for login attempts.
	LoginRateLimitWindow time.Duration `envconfig:"GO_GUESTBOOK_LOGIN_RATE_LIMIT_WINDOW" default:"1m"`

	// MessageRateLimitEnabled turns on public message POST rate limiting (default on).
	MessageRateLimitEnabled bool `envconfig:"GO_GUESTBOOK_MESSAGE_RATE_LIMIT_ENABLED" default:"true"`
	// MessageRateLimitMaxAttempts is how many message posts are allowed per IP inside the window.
	MessageRateLimitMaxAttempts int `envconfig:"GO_GUESTBOOK_MESSAGE_RATE_LIMIT_MAX_ATTEMPTS" default:"10"`
	// MessageRateLimitWindow is the sliding window duration for message posts.
	MessageRateLimitWindow time.Duration `envconfig:"GO_GUESTBOOK_MESSAGE_RATE_LIMIT_WINDOW" default:"1m"`
}

// Load reads configuration from environment variables into a typed Config struct.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if len(cfg.SessionSecret) < 32 {
		return nil, fmt.Errorf("GO_GUESTBOOK_SESSION_SECRET must be at least 32 characters")
	}
	if cfg.MaxMessages < 1 {
		return nil, fmt.Errorf("GO_GUESTBOOK_MAX_MESSAGES must be at least 1")
	}
	if cfg.LoginRateLimitMaxAttempts < 1 {
		return nil, fmt.Errorf("GO_GUESTBOOK_LOGIN_RATE_LIMIT_MAX_ATTEMPTS must be at least 1")
	}
	if cfg.LoginRateLimitWindow <= 0 {
		return nil, fmt.Errorf("GO_GUESTBOOK_LOGIN_RATE_LIMIT_WINDOW must be greater than 0")
	}
	if cfg.MessageRateLimitMaxAttempts < 1 {
		return nil, fmt.Errorf("GO_GUESTBOOK_MESSAGE_RATE_LIMIT_MAX_ATTEMPTS must be at least 1")
	}
	if cfg.MessageRateLimitWindow <= 0 {
		return nil, fmt.Errorf("GO_GUESTBOOK_MESSAGE_RATE_LIMIT_WINDOW must be greater than 0")
	}
	return &cfg, nil
}

// DSN builds a PostgreSQL connection string from database settings.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseName,
	)
}

// IsDevelopment reports whether the application runs in development mode.
// Development mode enables the /admin/tools UI and related endpoints.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

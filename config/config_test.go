package config_test

import (
	"testing"
	"time"

	"go-guestbook/config"
)

func TestLoadRequiresSessionSecretLength(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "short")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected short session secret to fail")
	}
}

func TestLoadBuildsDSN(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")
	t.Setenv("GO_GUESTBOOK_DATABASE_HOST", "localhost")
	t.Setenv("GO_GUESTBOOK_DATABASE_PORT", "5433")
	t.Setenv("GO_GUESTBOOK_DATABASE_NAME", "guestbook_test")
	t.Setenv("GO_GUESTBOOK_DATABASE_USER", "guest")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "host=localhost port=5433 user=guest password=secret dbname=guestbook_test sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{name: "development", env: "development", want: true},
		{name: "production", env: "production", want: false},
		{name: "empty treated as non-dev when set", env: "staging", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
			t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")
			t.Setenv("GO_GUESTBOOK_ENVIRONMENT", tt.env)

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Fatalf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevToolsEnabledDefaultOff(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DevToolsEnabled {
		t.Fatal("DevToolsEnabled default should be false")
	}
}

func TestDevToolsEnabledWhenSet(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")
	t.Setenv("GO_GUESTBOOK_DEV_TOOLS_ENABLED", "1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.DevToolsEnabled {
		t.Fatal("DevToolsEnabled should be true when set")
	}
}

func TestLoadDefaultsForLimitsAndRateLimits(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.MaxMessagesEnabled {
		t.Fatal("MaxMessagesEnabled default should be true")
	}
	if cfg.MaxMessages != 10000 {
		t.Fatalf("MaxMessages = %d, want 10000", cfg.MaxMessages)
	}
	if !cfg.LoginRateLimitEnabled {
		t.Fatal("LoginRateLimitEnabled default should be true")
	}
	if cfg.LoginRateLimitMaxAttempts != 10 {
		t.Fatalf("LoginRateLimitMaxAttempts = %d, want 10", cfg.LoginRateLimitMaxAttempts)
	}
	if cfg.LoginRateLimitWindow != time.Minute {
		t.Fatalf("LoginRateLimitWindow = %v, want 1m", cfg.LoginRateLimitWindow)
	}
	if !cfg.MessageRateLimitEnabled {
		t.Fatal("MessageRateLimitEnabled default should be true")
	}
	if cfg.MessageRateLimitMaxAttempts != 10 {
		t.Fatalf("MessageRateLimitMaxAttempts = %d, want 10", cfg.MessageRateLimitMaxAttempts)
	}
	if cfg.MessageRateLimitWindow != time.Minute {
		t.Fatalf("MessageRateLimitWindow = %v, want 1m", cfg.MessageRateLimitWindow)
	}
}

func TestLoadOverridesLimitsAndRateLimits(t *testing.T) {
	t.Setenv("GO_GUESTBOOK_SESSION_SECRET", "test-session-secret-with-32-characters")
	t.Setenv("GO_GUESTBOOK_DATABASE_PASSWORD", "secret")
	t.Setenv("GO_GUESTBOOK_MAX_MESSAGES_ENABLED", "false")
	t.Setenv("GO_GUESTBOOK_MAX_MESSAGES", "42")
	t.Setenv("GO_GUESTBOOK_LOGIN_RATE_LIMIT_ENABLED", "0")
	t.Setenv("GO_GUESTBOOK_LOGIN_RATE_LIMIT_MAX_ATTEMPTS", "3")
	t.Setenv("GO_GUESTBOOK_LOGIN_RATE_LIMIT_WINDOW", "30s")
	t.Setenv("GO_GUESTBOOK_MESSAGE_RATE_LIMIT_ENABLED", "false")
	t.Setenv("GO_GUESTBOOK_MESSAGE_RATE_LIMIT_MAX_ATTEMPTS", "5")
	t.Setenv("GO_GUESTBOOK_MESSAGE_RATE_LIMIT_WINDOW", "2m")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxMessagesEnabled {
		t.Fatal("MaxMessagesEnabled should be false")
	}
	if cfg.MaxMessages != 42 {
		t.Fatalf("MaxMessages = %d, want 42", cfg.MaxMessages)
	}
	if cfg.LoginRateLimitEnabled {
		t.Fatal("LoginRateLimitEnabled should be false")
	}
	if cfg.LoginRateLimitMaxAttempts != 3 {
		t.Fatalf("LoginRateLimitMaxAttempts = %d, want 3", cfg.LoginRateLimitMaxAttempts)
	}
	if cfg.LoginRateLimitWindow != 30*time.Second {
		t.Fatalf("LoginRateLimitWindow = %v, want 30s", cfg.LoginRateLimitWindow)
	}
	if cfg.MessageRateLimitEnabled {
		t.Fatal("MessageRateLimitEnabled should be false")
	}
	if cfg.MessageRateLimitMaxAttempts != 5 {
		t.Fatalf("MessageRateLimitMaxAttempts = %d, want 5", cfg.MessageRateLimitMaxAttempts)
	}
	if cfg.MessageRateLimitWindow != 2*time.Minute {
		t.Fatalf("MessageRateLimitWindow = %v, want 2m", cfg.MessageRateLimitWindow)
	}
}

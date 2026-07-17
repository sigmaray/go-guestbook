package middleware

import (
	"sync"
	"time"
)

const (
	defaultLoginRateLimitMaxAttempts   = 10
	defaultLoginRateLimitWindow        = time.Minute
	defaultMessageRateLimitMaxAttempts = 10
	defaultMessageRateLimitWindow      = time.Minute
)

// RateLimitConfig configures in-memory IP rate limits for login and message POSTs.
type RateLimitConfig struct {
	// LoginEnabled turns login rate limiting on or off.
	LoginEnabled bool
	// LoginMax is the maximum login attempts allowed per IP inside LoginWindow.
	LoginMax int
	// LoginWindow is the sliding window duration for login attempts.
	LoginWindow time.Duration
	// MessageEnabled turns public message rate limiting on or off.
	MessageEnabled bool
	// MessageMax is the maximum message posts allowed per IP inside MessageWindow.
	MessageMax int
	// MessageWindow is the sliding window duration for message posts.
	MessageWindow time.Duration
}

// attemptTracker stores recent request timestamps keyed by client IP.
type attemptTracker struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

// newAttemptTracker builds an in-memory IP attempt tracker.
// max is how many attempts are allowed inside window before further requests are denied.
func newAttemptTracker(max int, window time.Duration) *attemptTracker {
	return &attemptTracker{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

// allow reports whether another attempt is permitted for clientIP at now.
// clientIP identifies the caller; now is the current time used to slide the window.
func (t *attemptTracker) allow(clientIP string, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := now.Add(-t.window)
	recent := t.attempts[clientIP][:0]
	for _, ts := range t.attempts[clientIP] {
		if ts.After(cutoff) {
			recent = append(recent, ts)
		}
	}

	if len(recent) >= t.max {
		t.attempts[clientIP] = recent
		return false
	}

	t.attempts[clientIP] = append(recent, now)
	return true
}

var (
	rateLimitMu    sync.RWMutex
	loginEnabled   = true
	messageEnabled = true
	loginTracker   = newAttemptTracker(defaultLoginRateLimitMaxAttempts, defaultLoginRateLimitWindow)
	messageTracker = newAttemptTracker(defaultMessageRateLimitMaxAttempts, defaultMessageRateLimitWindow)
)

// ConfigureRateLimits applies runtime rate-limit settings from application config.
// cfg carries enable flags, max attempts, and window durations for login and messages.
func ConfigureRateLimits(cfg RateLimitConfig) {
	loginMax := cfg.LoginMax
	if loginMax <= 0 {
		loginMax = defaultLoginRateLimitMaxAttempts
	}
	loginWindow := cfg.LoginWindow
	if loginWindow <= 0 {
		loginWindow = defaultLoginRateLimitWindow
	}
	messageMax := cfg.MessageMax
	if messageMax <= 0 {
		messageMax = defaultMessageRateLimitMaxAttempts
	}
	messageWindow := cfg.MessageWindow
	if messageWindow <= 0 {
		messageWindow = defaultMessageRateLimitWindow
	}

	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()
	loginEnabled = cfg.LoginEnabled
	messageEnabled = cfg.MessageEnabled
	loginTracker = newAttemptTracker(loginMax, loginWindow)
	messageTracker = newAttemptTracker(messageMax, messageWindow)
}

// AllowLoginAttempt reports whether another login POST is allowed for clientIP.
// clientIP is the remote address of the caller (typically from Gin ClientIP).
func AllowLoginAttempt(clientIP string) bool {
	rateLimitMu.RLock()
	enabled := loginEnabled
	tracker := loginTracker
	rateLimitMu.RUnlock()
	if !enabled {
		return true
	}
	return tracker.allow(clientIP, time.Now())
}

// AllowMessageAttempt reports whether another public message POST is allowed for clientIP.
// clientIP is the remote address of the caller (typically from Gin ClientIP).
func AllowMessageAttempt(clientIP string) bool {
	rateLimitMu.RLock()
	enabled := messageEnabled
	tracker := messageTracker
	rateLimitMu.RUnlock()
	if !enabled {
		return true
	}
	return tracker.allow(clientIP, time.Now())
}

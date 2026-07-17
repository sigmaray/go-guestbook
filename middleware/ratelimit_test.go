package middleware

import (
	"testing"
	"time"
)

func TestLoginAttemptTrackerAllow(t *testing.T) {
	tracker := newAttemptTracker(defaultLoginRateLimitMaxAttempts, defaultLoginRateLimitWindow)
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.10"

	for i := 0; i < defaultLoginRateLimitMaxAttempts; i++ {
		if !tracker.allow(ip, now) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	if tracker.allow(ip, now) {
		t.Fatal("expected rate limit to block further attempts in the same window")
	}

	later := now.Add(defaultLoginRateLimitWindow + time.Second)
	if !tracker.allow(ip, later) {
		t.Fatal("expected attempts to be allowed after the window expires")
	}
}

func TestMessageAttemptTrackerAllow(t *testing.T) {
	tracker := newAttemptTracker(defaultMessageRateLimitMaxAttempts, defaultMessageRateLimitWindow)
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	ip := "203.0.113.20"

	for i := 0; i < defaultMessageRateLimitMaxAttempts; i++ {
		if !tracker.allow(ip, now) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	if tracker.allow(ip, now) {
		t.Fatal("expected rate limit to block further message posts in the same window")
	}
}

func TestAttemptTrackerIsolatesIPs(t *testing.T) {
	tracker := newAttemptTracker(defaultLoginRateLimitMaxAttempts, defaultLoginRateLimitWindow)
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	for i := 0; i < defaultLoginRateLimitMaxAttempts; i++ {
		if !tracker.allow("203.0.113.1", now) {
			t.Fatalf("first IP attempt %d should be allowed", i+1)
		}
	}

	if !tracker.allow("203.0.113.2", now) {
		t.Fatal("expected separate IP to have its own attempt budget")
	}
}

func TestConfigureRateLimitsDisablesLogin(t *testing.T) {
	ConfigureRateLimits(RateLimitConfig{
		LoginEnabled:   false,
		LoginMax:       1,
		LoginWindow:    time.Minute,
		MessageEnabled: true,
		MessageMax:     10,
		MessageWindow:  time.Minute,
	})
	t.Cleanup(func() {
		ConfigureRateLimits(RateLimitConfig{
			LoginEnabled:   true,
			LoginMax:       defaultLoginRateLimitMaxAttempts,
			LoginWindow:    defaultLoginRateLimitWindow,
			MessageEnabled: true,
			MessageMax:     defaultMessageRateLimitMaxAttempts,
			MessageWindow:  defaultMessageRateLimitWindow,
		})
	})

	for i := 0; i < 20; i++ {
		if !AllowLoginAttempt("203.0.113.50") {
			t.Fatalf("disabled login rate limit blocked attempt %d", i+1)
		}
	}
}

func TestConfigureRateLimitsCustomMessageMax(t *testing.T) {
	ConfigureRateLimits(RateLimitConfig{
		LoginEnabled:   true,
		LoginMax:       10,
		LoginWindow:    time.Minute,
		MessageEnabled: true,
		MessageMax:     2,
		MessageWindow:  time.Minute,
	})
	t.Cleanup(func() {
		ConfigureRateLimits(RateLimitConfig{
			LoginEnabled:   true,
			LoginMax:       defaultLoginRateLimitMaxAttempts,
			LoginWindow:    defaultLoginRateLimitWindow,
			MessageEnabled: true,
			MessageMax:     defaultMessageRateLimitMaxAttempts,
			MessageWindow:  defaultMessageRateLimitWindow,
		})
	})

	ip := "203.0.113.60"
	if !AllowMessageAttempt(ip) {
		t.Fatal("first message attempt should be allowed")
	}
	if !AllowMessageAttempt(ip) {
		t.Fatal("second message attempt should be allowed")
	}
	if AllowMessageAttempt(ip) {
		t.Fatal("third message attempt should be blocked with max=2")
	}
}

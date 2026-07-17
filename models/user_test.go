package models_test

import (
	"testing"

	"go-guestbook/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := models.HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !models.CheckPassword(hash, "secret-password") {
		t.Fatal("CheckPassword() expected true for matching password")
	}

	if models.CheckPassword(hash, "wrong-password") {
		t.Fatal("CheckPassword() expected false for non-matching password")
	}
}

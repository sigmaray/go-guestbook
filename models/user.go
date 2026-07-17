package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User stores administrator credentials for the private admin area.
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
}

// HashPassword returns a bcrypt hash for the given plain-text password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether the plain-text password matches the stored hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CreateUser persists a new administrator with a hashed password.
func CreateUser(db *gorm.DB, username, password string) (*User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
		Username:     username,
		PasswordHash: hash,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// FindUserByUsername loads a user by username and returns nil when not found.
func FindUserByUsername(db *gorm.DB, username string) (*User, error) {
	var user User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

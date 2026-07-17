package models

import (
	"time"

	"gorm.io/gorm"
)

// Message represents a guestbook entry submitted by a visitor.
// Author may be empty when the visitor skips the optional name field.
// IPAddress, UserAgent, Referer, and AcceptLanguage capture request metadata for moderation.
type Message struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Author         string         `json:"author"`
	Email          string         `json:"email"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	IPAddress      string         `json:"ip_address"`
	UserAgent      string         `json:"user_agent"`
	Referer        string         `json:"referer"`
	AcceptLanguage string         `json:"accept_language"`
}

// DisplayAuthor returns the author name for templates, or "Anonymous" when blank.
func (m Message) DisplayAuthor() string {
	if m.Author == "" {
		return "Anonymous"
	}
	return m.Author
}

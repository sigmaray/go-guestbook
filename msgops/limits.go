package msgops

import (
	"errors"
	"fmt"

	"go-guestbook/models"

	"gorm.io/gorm"
)

// ErrMaxMessagesReached is returned when creating another message would exceed the configured cap.
var ErrMaxMessagesReached = errors.New("maximum number of messages reached")

// WouldExceedMax reports whether count plus additional would exceed max.
// count is the current number of messages; max is the configured capacity;
// additional is how many messages are about to be inserted (treated as at least 1).
func WouldExceedMax(count int64, max int, additional int) bool {
	if additional < 1 {
		additional = 1
	}
	if max < 1 {
		return true
	}
	return count+int64(additional) > int64(max)
}

// EnsureRoomForMessages checks whether additional messages may be inserted under the capacity limit.
// db is the GORM handle used to count existing (non-deleted) messages.
// enabled turns the limit on or off; when false the check always succeeds.
// max is the maximum allowed message count when enabled.
// additional is how many new messages are about to be inserted (treated as at least 1).
func EnsureRoomForMessages(db *gorm.DB, enabled bool, max int, additional int) error {
	if !enabled {
		return nil
	}

	var count int64
	if err := db.Model(&models.Message{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count messages: %w", err)
	}
	if WouldExceedMax(count, max, additional) {
		return ErrMaxMessagesReached
	}
	return nil
}

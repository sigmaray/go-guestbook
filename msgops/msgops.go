package msgops

import (
	"fmt"

	"go-guestbook/models"

	"gorm.io/gorm"
)

// Seed creates the requested number of sample guestbook messages.
func Seed(db *gorm.DB, count int) (int, error) {
	for i := 1; i <= count; i++ {
		message := models.Message{
			Author:  fmt.Sprintf("Guest %d", i),
			Email:   fmt.Sprintf("guest%d@example.com", i),
			Content: fmt.Sprintf("Sample message number %d.", i),
		}
		if err := db.Create(&message).Error; err != nil {
			return i - 1, err
		}
	}
	return count, nil
}

// Clear removes all guestbook messages from the database.
func Clear(db *gorm.DB) (int64, error) {
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Message{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

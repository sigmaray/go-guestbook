package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go-guestbook/models"
	"go-guestbook/msgops"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AdminMessagesRedirect sends /admin/ to the messages list at /admin/messages.
func (h *Handler) AdminMessagesRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/messages")
}

// AdminMessagesList lists guestbook messages for administrators with pagination.
// page is an optional 1-based query parameter selecting which page of results to show.
func (h *Handler) AdminMessagesList(c *gin.Context) {
	page := parseQueryPage(c.Query("page"))
	perPage := adminListPageSize

	var total int64
	if err := h.DB.Model(&models.Message{}).Count(&total).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to count admin messages")
		h.adminHTML(c, http.StatusInternalServerError, "admin/messages_index.html", gin.H{
			"Error": "Failed to load messages",
		})
		return
	}

	page = clampPage(page, total, perPage)

	var messages []models.Message
	if err := h.DB.Order("created_at desc").
		Offset(pageOffset(page, perPage)).
		Limit(perPage).
		Find(&messages).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to load admin messages")
		h.adminHTML(c, http.StatusInternalServerError, "admin/messages_index.html", gin.H{
			"Error": "Failed to load messages",
		})
		return
	}

	pagination := buildPaginationView(total, page, perPage, "Messages", func(p int) string {
		return buildAdminListURL("/admin/messages", p)
	})

	h.adminHTML(c, http.StatusOK, "admin/messages_index.html", gin.H{
		"Messages":   messages,
		"Pagination": pagination,
	})
}

// NewMessagePage renders the administrator message creation form.
func (h *Handler) NewMessagePage(c *gin.Context) {
	h.adminHTML(c, http.StatusOK, "admin/messages_new.html", gin.H{})
}

// AdminCreateMessageInput validates administrator message forms.
// Author is optional to match the public guestbook form.
type AdminCreateMessageInput struct {
	Author  string `form:"author" validate:"omitempty,min=2,max=100"`
	Email   string `form:"email" validate:"omitempty,email,max=255"`
	Content string `form:"content" validate:"required,min=1,max=5000"`
}

// CreateAdminMessage stores a message created from the admin panel.
func (h *Handler) CreateAdminMessage(c *gin.Context) {
	var input AdminCreateMessageInput
	if err := c.ShouldBind(&input); err != nil {
		h.renderAdminCreateError(c, "Invalid form data", input)
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		h.renderAdminCreateError(c, validationMessage(err), input)
		return
	}

	if err := msgops.EnsureRoomForMessages(h.DB, h.MaxMessagesEnabled, h.MaxMessages, 1); err != nil {
		if errors.Is(err, msgops.ErrMaxMessagesReached) {
			h.renderAdminCreateErrorStatus(c, http.StatusForbidden, "The guestbook has reached its maximum number of messages.", input)
			return
		}
		h.Logger.Error().Err(err).Msg("failed to check message capacity")
		h.renderAdminCreateError(c, "Failed to create message", input)
		return
	}

	message := models.Message{
		Author:  input.Author,
		Email:   input.Email,
		Content: input.Content,
	}

	if err := h.DB.Create(&message).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to create admin message")
		h.renderAdminCreateError(c, "Failed to create message", input)
		return
	}

	c.Redirect(http.StatusFound, "/admin/messages")
}

// ShowMessagePage renders a read-only administrator view of a single guestbook message.
// It is used when an admin opens /admin/messages/:id/ to inspect a message without editing.
func (h *Handler) ShowMessagePage(c *gin.Context) {
	message, ok := h.loadMessageOrRedirect(c)
	if !ok {
		return
	}

	h.adminHTML(c, http.StatusOK, "admin/messages_show.html", gin.H{
		"Message": message,
	})
}

// EditMessagePage renders the administrator message edit form.
func (h *Handler) EditMessagePage(c *gin.Context) {
	message, ok := h.loadMessageOrRedirect(c)
	if !ok {
		return
	}

	h.adminHTML(c, http.StatusOK, "admin/messages_edit.html", gin.H{
		"Message": message,
	})
}

// AdminUpdateMessageInput validates administrator message updates.
// Author is optional to match the public guestbook form.
type AdminUpdateMessageInput struct {
	Author  string `form:"author" validate:"omitempty,min=2,max=100"`
	Email   string `form:"email" validate:"omitempty,email,max=255"`
	Content string `form:"content" validate:"required,min=1,max=5000"`
}

// UpdateMessage updates an existing guestbook message from the admin panel.
func (h *Handler) UpdateMessage(c *gin.Context) {
	message, ok := h.loadMessageOrRedirect(c)
	if !ok {
		return
	}

	var input AdminUpdateMessageInput
	if err := c.ShouldBind(&input); err != nil {
		h.renderAdminEditError(c, message, "Invalid form data", input)
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		h.renderAdminEditError(c, message, validationMessage(err), input)
		return
	}

	message.Author = input.Author
	message.Email = input.Email
	message.Content = input.Content

	if err := h.DB.Save(&message).Error; err != nil {
		h.Logger.Error().Err(err).Uint("message_id", message.ID).Msg("failed to update message")
		h.renderAdminEditError(c, message, "Failed to update message", input)
		return
	}

	c.Redirect(http.StatusFound, "/admin/messages")
}

// DeleteMessage soft-deletes a guestbook message from the admin panel.
func (h *Handler) DeleteMessage(c *gin.Context) {
	message, ok := h.loadMessageOrRedirect(c)
	if !ok {
		return
	}

	if err := h.DB.Delete(&message).Error; err != nil {
		h.Logger.Error().Err(err).Uint("message_id", message.ID).Msg("failed to delete message")
	}

	c.Redirect(http.StatusFound, "/admin/messages")
}

// loadMessageOrRedirect parses the message ID and loads the record for admin actions.
func (h *Handler) loadMessageOrRedirect(c *gin.Context) (models.Message, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/messages")
		return models.Message{}, false
	}

	var message models.Message
	if err := h.DB.First(&message, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/messages")
		return models.Message{}, false
	}

	return message, true
}

// renderAdminCreateError re-renders the admin create form with a 400 validation error.
// message is the user-facing error text; input holds the submitted form values to redisplay.
func (h *Handler) renderAdminCreateError(c *gin.Context, message string, input AdminCreateMessageInput) {
	h.renderAdminCreateErrorStatus(c, http.StatusBadRequest, message, input)
}

// renderAdminCreateErrorStatus re-renders the admin create form with an error and preserved input.
// status is the HTTP status code; message is the user-facing error text; input holds form values to redisplay.
func (h *Handler) renderAdminCreateErrorStatus(c *gin.Context, status int, message string, input AdminCreateMessageInput) {
	h.adminHTML(c, status, "admin/messages_new.html", gin.H{
		"Error":   message,
		"Author":  input.Author,
		"Email":   input.Email,
		"Content": input.Content,
	})
}

// renderAdminEditError re-renders the admin edit form with validation feedback.
func (h *Handler) renderAdminEditError(c *gin.Context, message models.Message, errorMessage string, input AdminUpdateMessageInput) {
	h.adminHTML(c, http.StatusBadRequest, "admin/messages_edit.html", gin.H{
		"Message": message,
		"Error":   errorMessage,
		"Author":  input.Author,
		"Email":   input.Email,
		"Content": input.Content,
	})
}

// validationMessage converts validator errors into a user-facing message.
func validationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			switch fieldError.Field() {
			case "Author":
				return "Name must be between 2 and 100 characters when provided"
			case "Email":
				return "Email must be a valid address"
			case "Content":
				return "Message content is required"
			default:
				return fmt.Sprintf("%s is invalid", fieldError.Field())
			}
		}
	}
	return "Invalid input"
}

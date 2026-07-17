package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-guestbook/middleware"
	"go-guestbook/models"
	"go-guestbook/msgops"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LoginPage renders the administrator login form.
func (h *Handler) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user") != nil {
		c.Redirect(http.StatusFound, "/admin/messages")
		return
	}
	c.HTML(http.StatusOK, "admin/login.html", gin.H{})
}

// Login authenticates an administrator and stores the username in the session.
// username and password come from the login form POST body.
func (h *Handler) Login(c *gin.Context) {
	if !middleware.AllowLoginAttempt(c.ClientIP()) {
		c.HTML(http.StatusTooManyRequests, "admin/login.html", gin.H{
			"Error": "Too many login attempts. Please try again later.",
		})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := models.FindUserByUsername(h.DB, username)
	if err != nil {
		h.Logger.Error().Err(err).Str("username", username).Msg("login lookup failed")
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	if user == nil || !models.CheckPassword(user.PasswordHash, password) {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user", user.Username)
	if err := session.Save(); err != nil {
		h.Logger.Error().Err(err).Str("username", username).Msg("failed to save session")
		c.HTML(http.StatusInternalServerError, "admin/login.html", gin.H{
			"Error": "Failed to start session",
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/messages")
}

// Logout clears the administrator session and returns to the public homepage.
func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		h.Logger.Error().Err(err).Msg("failed to clear session")
	}
	c.Redirect(http.StatusFound, "/")
}

// Index renders the public guestbook with messages and a submission form.
// page is an optional 1-based query parameter for pagination.
func (h *Handler) Index(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit

	var messages []models.Message
	query := h.DB.Order("created_at desc")

	var total int64
	if err := query.Model(&models.Message{}).Count(&total).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to count messages")
		c.HTML(http.StatusInternalServerError, "public/index.html", gin.H{
			"Error":      "Failed to load messages",
			"IsLoggedIn": h.isLoggedIn(c),
		})
		return
	}

	if err := query.Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to list messages")
		c.HTML(http.StatusInternalServerError, "public/index.html", gin.H{
			"Error":      "Failed to load messages",
			"IsLoggedIn": h.isLoggedIn(c),
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	pages := make([]int, totalPages)
	for i := range pages {
		pages[i] = i + 1
	}

	c.HTML(http.StatusOK, "public/index.html", gin.H{
		"Messages":   messages,
		"Page":       page,
		"Pages":      pages,
		"TotalPages": totalPages,
		"HasNext":    int64(page*limit) < total,
		"IsLoggedIn": h.isLoggedIn(c),
	})
}

// CreateMessageInput validates public guestbook submissions.
// Author is optional; when provided it must be between 2 and 100 characters.
type CreateMessageInput struct {
	Author  string `form:"author" validate:"omitempty,min=2,max=100"`
	Email   string `form:"email" validate:"omitempty,email,max=255"`
	Content string `form:"content" validate:"required,min=1,max=5000"`
}

// CreateMessage stores a new guestbook entry from the public form.
// Form fields carry author, email, and content; request headers supply visitor metadata.
func (h *Handler) CreateMessage(c *gin.Context) {
	if !middleware.AllowMessageAttempt(c.ClientIP()) {
		h.renderIndexWithFormErrorStatus(c, http.StatusTooManyRequests, "Too many messages. Please try again later.", CreateMessageInput{})
		return
	}

	var input CreateMessageInput
	if err := c.ShouldBind(&input); err != nil {
		h.renderIndexWithFormError(c, "Invalid form data", input)
		return
	}

	input.Author = strings.TrimSpace(input.Author)
	input.Email = strings.TrimSpace(input.Email)
	input.Content = strings.TrimSpace(input.Content)

	if err := h.Validate.Struct(input); err != nil {
		h.renderIndexWithFormError(c, validationMessage(err), input)
		return
	}

	if err := msgops.EnsureRoomForMessages(h.DB, h.MaxMessagesEnabled, h.MaxMessages, 1); err != nil {
		if errors.Is(err, msgops.ErrMaxMessagesReached) {
			h.renderIndexWithFormErrorStatus(c, http.StatusForbidden, "The guestbook has reached its maximum number of messages.", input)
			return
		}
		h.Logger.Error().Err(err).Msg("failed to check message capacity")
		h.renderIndexWithFormError(c, "Failed to save message", input)
		return
	}

	message := models.Message{
		Author:         input.Author,
		Email:          input.Email,
		Content:        input.Content,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		Referer:        c.Request.Referer(),
		AcceptLanguage: c.GetHeader("Accept-Language"),
	}

	if err := h.DB.Create(&message).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to create message")
		h.renderIndexWithFormError(c, "Failed to save message", input)
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// renderIndexWithFormError re-renders the homepage with a 400 validation error and preserved input.
// message is the user-facing error text; input holds the submitted form values to redisplay.
func (h *Handler) renderIndexWithFormError(c *gin.Context, message string, input CreateMessageInput) {
	h.renderIndexWithFormErrorStatus(c, http.StatusBadRequest, message, input)
}

// renderIndexWithFormErrorStatus re-renders the homepage with an error and preserved input.
// status is the HTTP status code; message is the user-facing error text; input holds form values to redisplay.
func (h *Handler) renderIndexWithFormErrorStatus(c *gin.Context, status int, message string, input CreateMessageInput) {
	var messages []models.Message
	if err := h.DB.Order("created_at desc").Limit(10).Find(&messages).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to reload messages after validation error")
	}

	c.HTML(status, "public/index.html", gin.H{
		"Messages":      messages,
		"Error":         message,
		"Author":        input.Author,
		"Email":         input.Email,
		"Content":       input.Content,
		"Page":          1,
		"Pages":         []int{1},
		"TotalPages":    1,
		"HasNext":       false,
		"ShowFormError": true,
		"IsLoggedIn":    h.isLoggedIn(c),
	})
}

// isLoggedIn reports whether the current request has an authenticated admin session.
func (h *Handler) isLoggedIn(c *gin.Context) bool {
	session := sessions.Default(c)
	return session.Get("user") != nil
}

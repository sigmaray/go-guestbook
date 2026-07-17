package handlers

import (
	"net/http"
	"strconv"

	"go-guestbook/models"

	"github.com/gin-gonic/gin"
)

// UsersList renders the administrator user list with pagination.
// page is an optional 1-based query parameter selecting which page of results to show.
func (h *Handler) UsersList(c *gin.Context) {
	page := parseQueryPage(c.Query("page"))
	perPage := adminListPageSize

	var total int64
	if err := h.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to count admin users")
		h.adminHTML(c, http.StatusInternalServerError, "admin/users_list.html", gin.H{
			"Error": "Failed to load users",
		})
		return
	}

	page = clampPage(page, total, perPage)

	var users []models.User
	if err := h.DB.Order("created_at desc").
		Offset(pageOffset(page, perPage)).
		Limit(perPage).
		Find(&users).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to load admin users")
		h.adminHTML(c, http.StatusInternalServerError, "admin/users_list.html", gin.H{
			"Error": "Failed to load users",
		})
		return
	}

	pagination := buildPaginationView(total, page, perPage, "Users", func(p int) string {
		return buildAdminListURL("/admin/users", p)
	})

	h.adminHTML(c, http.StatusOK, "admin/users_list.html", gin.H{
		"Users":      users,
		"Pagination": pagination,
	})
}

// NewUserPage renders the administrator user creation form.
func (h *Handler) NewUserPage(c *gin.Context) {
	h.adminHTML(c, http.StatusOK, "admin/users_create.html", gin.H{})
}

// CreateUserInput validates administrator user creation forms.
type CreateUserInput struct {
	Username        string `form:"username" validate:"required,min=2,max=100"`
	Password        string `form:"password" validate:"required,min=4,max=128"`
	PasswordConfirm string `form:"password_confirm" validate:"required"`
}

// CreateUser stores a new administrator from the admin panel.
// username, password, and password_confirm come from the create-user form.
func (h *Handler) CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBind(&input); err != nil {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_create.html", gin.H{
			"Error":    "Invalid form data",
			"Username": input.Username,
		})
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_create.html", gin.H{
			"Error":    "Username and password are required",
			"Username": input.Username,
		})
		return
	}

	if input.Password != input.PasswordConfirm {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_create.html", gin.H{
			"Error":    "Passwords do not match",
			"Username": input.Username,
		})
		return
	}

	if _, err := models.CreateUser(h.DB, input.Username, input.Password); err != nil {
		h.Logger.Error().Err(err).Str("username", input.Username).Msg("failed to create user")
		h.adminHTML(c, http.StatusInternalServerError, "admin/users_create.html", gin.H{
			"Error":    "Failed to create user (username may already exist)",
			"Username": input.Username,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

// EditUserPage renders the administrator user edit form.
// id is the user primary key from the URL path.
func (h *Handler) EditUserPage(c *gin.Context) {
	user, ok := h.loadUserOrRedirect(c)
	if !ok {
		return
	}

	h.adminHTML(c, http.StatusOK, "admin/users_edit.html", gin.H{
		"User": user,
	})
}

// UpdateUserInput validates administrator user update forms.
// Password fields are optional so an edit can rename without resetting the password.
type UpdateUserInput struct {
	Username        string `form:"username" validate:"required,min=2,max=100"`
	Password        string `form:"password" validate:"omitempty,min=4,max=128"`
	PasswordConfirm string `form:"password_confirm"`
}

// UpdateUser updates an existing administrator from the admin panel.
// id is the user primary key from the URL path; form fields carry username and optional password.
func (h *Handler) UpdateUser(c *gin.Context) {
	user, ok := h.loadUserOrRedirect(c)
	if !ok {
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBind(&input); err != nil {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_edit.html", gin.H{
			"Error": "Invalid form data",
			"User":  user,
		})
		return
	}

	if err := h.Validate.Struct(input); err != nil {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_edit.html", gin.H{
			"Error": "Username is required",
			"User":  user,
		})
		return
	}

	if input.Password != "" && input.Password != input.PasswordConfirm {
		h.adminHTML(c, http.StatusBadRequest, "admin/users_edit.html", gin.H{
			"Error": "Passwords do not match",
			"User":  user,
		})
		return
	}

	user.Username = input.Username
	if input.Password != "" {
		hash, err := models.HashPassword(input.Password)
		if err != nil {
			h.Logger.Error().Err(err).Uint("user_id", user.ID).Msg("failed to hash password")
			h.adminHTML(c, http.StatusInternalServerError, "admin/users_edit.html", gin.H{
				"Error": "Failed to update password",
				"User":  user,
			})
			return
		}
		user.PasswordHash = hash
	}

	if err := h.DB.Save(&user).Error; err != nil {
		h.Logger.Error().Err(err).Uint("user_id", user.ID).Msg("failed to update user")
		h.adminHTML(c, http.StatusInternalServerError, "admin/users_edit.html", gin.H{
			"Error": "Failed to update user (username may already exist)",
			"User":  user,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

// DeleteUser permanently removes an administrator from the admin panel.
// id is the user primary key from the URL path.
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	if err := h.DB.Unscoped().Delete(&models.User{}, id).Error; err != nil {
		h.Logger.Error().Err(err).Uint64("user_id", id).Msg("failed to delete user")
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

// loadUserOrRedirect parses the user ID and loads the record for admin actions.
// Returns the user and true on success; redirects to the users list and returns false otherwise.
func (h *Handler) loadUserOrRedirect(c *gin.Context) (models.User, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return models.User{}, false
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return models.User{}, false
	}

	return user, true
}

package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"go-guestbook/models"

	"github.com/gin-gonic/gin"
)

var allowedTestTables = map[string]struct{}{
	"messages": {},
	"users":    {},
}

// TruncateTableRequest identifies a table that should be emptied in tests.
type TruncateTableRequest struct {
	Table string `json:"table" validate:"required"`
}

// CreateEntityRequest describes a test entity to insert through the test API.
type CreateEntityRequest struct {
	Table  string                 `json:"table" validate:"required"`
	Values map[string]interface{} `json:"values" validate:"required"`
}

// ExecuteSQLRequest carries a single SQL statement for test setup.
type ExecuteSQLRequest struct {
	Query string `json:"query" validate:"required"`
}

// TruncateTable removes all rows from an allowed table for isolated tests.
func (h *Handler) TruncateTable(c *gin.Context) {
	var req TruncateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table is required"})
		return
	}

	table := strings.ToLower(strings.TrimSpace(req.Table))
	if !isAllowedTestTable(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table not allowed"})
		return
	}

	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database unavailable"})
		return
	}

	stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
	if _, err := sqlDB.ExecContext(c.Request.Context(), stmt); err != nil {
		h.Logger.Error().Err(err).Str("table", table).Msg("truncate failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "truncate failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "table": table})
}

// CreateEntity inserts a test record into an allowed table.
func (h *Handler) CreateEntity(c *gin.Context) {
	var req CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table and values are required"})
		return
	}

	table := strings.ToLower(strings.TrimSpace(req.Table))
	if !isAllowedTestTable(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table not allowed"})
		return
	}

	switch table {
	case "messages":
		h.createTestMessage(c, req.Values)
	case "users":
		h.createTestUser(c, req.Values)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "table not supported"})
	}
}

// ExecuteSQL runs a single SQL statement and returns rows for SELECT queries.
func (h *Handler) ExecuteSQL(c *gin.Context) {
	var req ExecuteSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" || strings.Contains(query, ";") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only a single SQL statement is allowed"})
		return
	}

	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database unavailable"})
		return
	}

	rows, err := sqlDB.QueryContext(c.Request.Context(), query)
	if err != nil {
		h.Logger.Error().Err(err).Msg("test sql query failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	result, err := scanRows(rows)
	if err != nil {
		h.Logger.Error().Err(err).Msg("failed to scan test sql result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read query result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows": result})
}

// createTestMessage inserts a guestbook message using test API values.
func (h *Handler) createTestMessage(c *gin.Context, values map[string]interface{}) {
	author, _ := values["author"].(string)
	email, _ := values["email"].(string)
	content, _ := values["content"].(string)

	message := models.Message{
		Author:  author,
		Email:   email,
		Content: content,
	}

	if err := h.Validate.Struct(CreateMessageInput{
		Author:  author,
		Email:   email,
		Content: content,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationMessage(err)})
		return
	}

	if err := h.DB.Create(&message).Error; err != nil {
		h.Logger.Error().Err(err).Msg("failed to create test message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"entity": message})
}

// createTestUser inserts an administrator using a plain-text password from tests.
func (h *Handler) createTestUser(c *gin.Context, values map[string]interface{}) {
	username, _ := values["username"].(string)
	password, _ := values["password"].(string)

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	user, err := models.CreateUser(h.DB, username, password)
	if err != nil {
		h.Logger.Error().Err(err).Str("username", username).Msg("failed to create test user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"entity": user})
}

// isAllowedTestTable reports whether a table name is permitted for test helpers.
func isAllowedTestTable(table string) bool {
	_, ok := allowedTestTables[table]
	return ok
}

// scanRows converts SQL rows into JSON-friendly maps.
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanTargets := make([]interface{}, len(columns))
		for i := range values {
			scanTargets[i] = &values[i]
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			row[column] = normalizeSQLValue(values[i])
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result == nil {
		result = []map[string]interface{}{}
	}

	return result, nil
}

// normalizeSQLValue converts database driver values into JSON-safe types.
func normalizeSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

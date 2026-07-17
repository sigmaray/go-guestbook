package handlers

import (
	"context"
	"net/http"
	"time"

	"go-guestbook/database"
	"go-guestbook/msgops"

	"github.com/gin-gonic/gin"
)

const toolsSeedMessageCount = 10

// ToolsPage renders the development tools page for administrators.
// It lists public database tables so operators can clear data or run SQL locally.
func (h *Handler) ToolsPage(c *gin.Context) {
	tables, err := database.ListTables(h.DB)
	if err != nil {
		h.Logger.Error().Err(err).Msg("failed to list tables for tools page")
		tables = nil
	}

	h.adminHTML(c, http.StatusOK, "admin/tools/index.html", gin.H{
		"Tables": tables,
	})
}

// toolsClearTableRequest identifies which table the tools UI should truncate.
type toolsClearTableRequest struct {
	Table string `json:"table" validate:"required"`
}

// toolsSQLRequest carries a SQL statement from the development tools UI.
type toolsSQLRequest struct {
	Query string `json:"query" validate:"required"`
}

// ToolsClearTable truncates the selected table when development tools are enabled.
// The JSON body supplies table as the PostgreSQL table name to clear.
func (h *Handler) ToolsClearTable(c *gin.Context) {
	var req toolsClearTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table is required"})
		return
	}

	if err := database.ClearTable(h.DB, req.Table); err != nil {
		h.Logger.Error().Err(err).Str("table", req.Table).Msg("tools clear table failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ToolsExecuteSQL runs SQL with a one-minute timeout and returns rows or rows_affected.
// The JSON body supplies query as the SQL statement entered in the tools UI.
func (h *Handler) ToolsExecuteSQL(c *gin.Context) {
	var req toolsSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Minute)
	defer cancel()

	type result struct {
		columns  []string
		rows     [][]string
		affected int64
		err      error
	}
	ch := make(chan result, 1)
	go func() {
		cols, rows, affected, err := database.ExecuteSQL(h.DB.WithContext(ctx), req.Query)
		ch <- result{columns: cols, rows: rows, affected: affected, err: err}
	}()

	select {
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "query timeout exceeded"})
		return
	case res := <-ch:
		if res.err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.err.Error()})
			return
		}
		if res.columns != nil {
			c.JSON(http.StatusOK, gin.H{
				"columns": res.columns,
				"rows":    res.rows,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"rows_affected": res.affected,
		})
	}
}

// ToolsSeedMessages inserts sample guestbook messages for local development.
func (h *Handler) ToolsSeedMessages(c *gin.Context) {
	created, err := msgops.Seed(h.DB, toolsSeedMessageCount)
	if err != nil {
		h.Logger.Error().Err(err).Int("created", created).Msg("tools seed messages failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"created": created})
}

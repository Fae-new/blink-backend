package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"postman-runner/internal/config"
	"postman-runner/internal/models"
	"postman-runner/internal/validator"

	"github.com/gin-gonic/gin"
)

type CollectionHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewCollectionHandler(db *sql.DB, cfg *config.Config) *CollectionHandler {
	return &CollectionHandler{
		db:  db,
		cfg: cfg,
	}
}

// UploadCollection handles POST /collections/upload
func (h *CollectionHandler) UploadCollection(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, h.cfg.MaxRequestSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "read_error",
			Message: "Failed to read request body",
		})
		return
	}

	// Validate Postman collection
	collection, err := validator.ValidatePostmanCollection(body, h.cfg.MaxRequestSize, h.cfg.MaxHeaderCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Begin transaction
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback()

	// Insert collection
	var collectionID int
	err = tx.QueryRow(`
		INSERT INTO collections (name, description)
		VALUES ($1, $2)
		RETURNING id
	`, collection.Info.Name, collection.Info.Description).Scan(&collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create collection",
		})
		return
	}

	// Recursively import items
	if err := h.importItems(tx, collectionID, 0, collection.Item, 0); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "import_error",
			Message: fmt.Sprintf("Failed to import items: %v", err),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"collection_id": collectionID,
		"message":       "Collection imported successfully",
	})
}

func (h *CollectionHandler) importItems(tx *sql.Tx, collectionID int, parentID int, items []models.PostmanItem, startOrder int) error {
	for i, item := range items {
		sortOrder := startOrder + i

		if len(item.Item) > 0 {
			// It's a folder
			var folderID int
			err := tx.QueryRow(`
				INSERT INTO collection_items (collection_id, parent_id, name, item_type, sort_order)
				VALUES ($1, $2, $3, 'folder', $4)
				RETURNING id
			`, collectionID, nullInt(parentID), item.Name, sortOrder).Scan(&folderID)
			if err != nil {
				return fmt.Errorf("failed to insert folder: %w", err)
			}

			// Recursively import folder items
			if err := h.importItems(tx, collectionID, folderID, item.Item, 0); err != nil {
				return err
			}
		} else if item.Request != nil {
			// It's a request
			req := item.Request

			// Extract URL (can be empty for requests without URL)
			urlStr := extractURL(req.URL)

			// Convert headers to JSON
			headersJSON, err := json.Marshal(req.Header)
			if err != nil {
				return fmt.Errorf("failed to marshal headers: %w", err)
			}

			// Extract body
			body := ""
			if req.Body != nil && req.Body.Raw != "" {
				body = req.Body.Raw
			}

			_, err = tx.Exec(`
				INSERT INTO collection_items (collection_id, parent_id, name, item_type, sort_order, method, url, headers, body)
				VALUES ($1, $2, $3, 'request', $4, $5, $6, $7, $8)
			`, collectionID, nullInt(parentID), item.Name, sortOrder, req.Method, urlStr, string(headersJSON), body)
			if err != nil {
				return fmt.Errorf("failed to insert request: %w", err)
			}
		}
	}
	return nil
}

func extractURL(urlData interface{}) string {
	if urlData == nil {
		return ""
	}
	switch v := urlData.(type) {
	case string:
		return v
	case map[string]interface{}:
		if raw, ok := v["raw"].(string); ok {
			return raw
		}
		return ""
	default:
		return ""
	}
}

func nullInt(val int) interface{} {
	if val == 0 {
		return nil
	}
	return val
}

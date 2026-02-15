package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"postman-runner/internal/config"
	"postman-runner/internal/models"

	"github.com/gin-gonic/gin"
)

type UpdateItemRequest struct {
	Name            *string                  `json:"name,omitempty"`
	Method          *string                  `json:"method,omitempty"`
	URL             *string                  `json:"url,omitempty"`
	Headers         *[]models.PostmanHeader  `json:"headers,omitempty"`
	Body            *string                  `json:"body,omitempty"`
	ExtractionRules *[]models.ExtractionRule `json:"extraction_rules,omitempty"`
}

type CreateItemRequest struct {
	Name            string                  `json:"name" binding:"required"`
	ItemType        string                  `json:"item_type" binding:"required"` // "folder" or "request"
	ParentID        *int                    `json:"parent_id,omitempty"`
	Method          string                  `json:"method,omitempty"`
	URL             string                  `json:"url,omitempty"`
	Headers         []models.PostmanHeader  `json:"headers,omitempty"`
	Body            string                  `json:"body,omitempty"`
	ExtractionRules []models.ExtractionRule `json:"extraction_rules,omitempty"`
}

type ItemHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewItemHandler(db *sql.DB, cfg *config.Config) *ItemHandler {
	return &ItemHandler{
		db:  db,
		cfg: cfg,
	}
}

// CreateItem handles POST /collections/:id/items
func (h *ItemHandler) CreateItem(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Collection ID must be a valid integer",
		})
		return
	}

	// Verify collection exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1)", collectionID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to check collection existence",
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Collection not found",
		})
		return
	}

	var createReq CreateItemRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Name and item_type are required",
		})
		return
	}

	// Validate item_type
	if createReq.ItemType != "folder" && createReq.ItemType != "request" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_item_type",
			Message: "item_type must be 'folder' or 'request'",
		})
		return
	}

	// Validate request-specific fields
	if createReq.ItemType == "request" {
		if createReq.Method == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "missing_method",
				Message: "method is required for request items",
			})
			return
		}

		allowedMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		}
		if !allowedMethods[createReq.Method] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_method",
				Message: "Method must be one of: GET, POST, PUT, PATCH, DELETE",
			})
			return
		}
	}

	// Verify parent exists if parent_id is provided
	if createReq.ParentID != nil {
		var parentExists bool
		var parentCollectionID int
		err = h.db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM collection_items WHERE id = $1), collection_id 
			FROM collection_items WHERE id = $1
		`, *createReq.ParentID).Scan(&parentExists, &parentCollectionID)
		if err != nil || !parentExists {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_parent",
				Message: "Parent item not found",
			})
			return
		}
		if parentCollectionID != collectionID {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_parent",
				Message: "Parent item must belong to the same collection",
			})
			return
		}
	}

	// Calculate sort_order (max + 1 for the parent level)
	var maxSortOrder sql.NullInt64
	if createReq.ParentID != nil {
		err = h.db.QueryRow(`
			SELECT MAX(sort_order) FROM collection_items 
			WHERE collection_id = $1 AND parent_id = $2
		`, collectionID, *createReq.ParentID).Scan(&maxSortOrder)
	} else {
		err = h.db.QueryRow(`
			SELECT MAX(sort_order) FROM collection_items 
			WHERE collection_id = $1 AND parent_id IS NULL
		`, collectionID).Scan(&maxSortOrder)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to calculate sort order",
		})
		return
	}

	sortOrder := 0
	if maxSortOrder.Valid {
		sortOrder = int(maxSortOrder.Int64) + 1
	}

	// Insert item
	var newItem models.CollectionItem

	// Serialize extraction_rules
	extractionRulesJSON, err := json.Marshal(createReq.ExtractionRules)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_extraction_rules",
			Message: "Failed to serialize extraction_rules",
		})
		return
	}

	var extractionRulesBytes []byte
	if createReq.ItemType == "folder" {
		err = h.db.QueryRow(`
			INSERT INTO collection_items (collection_id, parent_id, name, item_type, sort_order, extraction_rules)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, collection_id, parent_id, name, item_type, sort_order, extraction_rules, created_at, updated_at
		`, collectionID, createReq.ParentID, createReq.Name, createReq.ItemType, sortOrder, extractionRulesJSON).Scan(
			&newItem.ID,
			&newItem.CollectionID,
			&newItem.ParentID,
			&newItem.Name,
			&newItem.ItemType,
			&newItem.SortOrder,
			&extractionRulesBytes,
			&newItem.CreatedAt,
			&newItem.UpdatedAt,
		)
	} else {
		// Request item
		headersJSON, err := json.Marshal(createReq.Headers)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_headers",
				Message: "Failed to serialize headers",
			})
			return
		}

		err = h.db.QueryRow(`
			INSERT INTO collection_items (collection_id, parent_id, name, item_type, sort_order, method, url, headers, body, extraction_rules)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, collection_id, parent_id, name, item_type, sort_order, method, url, headers, body, extraction_rules, created_at, updated_at
		`, collectionID, createReq.ParentID, createReq.Name, createReq.ItemType, sortOrder,
			createReq.Method, createReq.URL, string(headersJSON), createReq.Body, extractionRulesJSON).Scan(
			&newItem.ID,
			&newItem.CollectionID,
			&newItem.ParentID,
			&newItem.Name,
			&newItem.ItemType,
			&newItem.SortOrder,
			&newItem.Method,
			&newItem.URL,
			&newItem.Headers,
			&newItem.Body,
			&extractionRulesBytes,
			&newItem.CreatedAt,
			&newItem.UpdatedAt,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create item",
		})
		return
	}

	// Parse extraction_rules back
	if err := json.Unmarshal(extractionRulesBytes, &newItem.ExtractionRules); err != nil {
		newItem.ExtractionRules = []models.ExtractionRule{}
	}

	c.JSON(http.StatusCreated, newItem)
}

// UpdateItem handles PUT /items/:id (full update)
func (h *ItemHandler) UpdateItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Item ID must be a valid integer",
		})
		return
	}

	// Read request body
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, h.cfg.MaxRequestSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "read_error",
			Message: "Failed to read request body",
		})
		return
	}

	var updateReq UpdateItemRequest
	if err := json.Unmarshal(body, &updateReq); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_json",
			Message: "Invalid JSON format",
		})
		return
	}

	// Check if item exists and is a request (not a folder)
	var itemType string
	err = h.db.QueryRow("SELECT item_type FROM collection_items WHERE id = $1", itemID).Scan(&itemType)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Item not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch item",
		})
		return
	}

	if itemType != "request" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_item_type",
			Message: "Only items of type 'request' can be updated with this endpoint",
		})
		return
	}

	// Validate method if provided
	if updateReq.Method != nil {
		allowedMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		}
		if !allowedMethods[*updateReq.Method] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_method",
				Message: "Method must be one of: GET, POST, PUT, PATCH, DELETE",
			})
			return
		}
	}

	// Build update query dynamically based on provided fields
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if updateReq.Name != nil {
		updates = append(updates, "name = $"+strconv.Itoa(argCount))
		args = append(args, *updateReq.Name)
		argCount++
	}

	if updateReq.Method != nil {
		updates = append(updates, "method = $"+strconv.Itoa(argCount))
		args = append(args, *updateReq.Method)
		argCount++
	}

	if updateReq.URL != nil {
		updates = append(updates, "url = $"+strconv.Itoa(argCount))
		args = append(args, *updateReq.URL)
		argCount++
	}

	if updateReq.Headers != nil {
		headersJSON, err := json.Marshal(updateReq.Headers)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_headers",
				Message: "Failed to serialize headers",
			})
			return
		}
		updates = append(updates, "headers = $"+strconv.Itoa(argCount))
		args = append(args, string(headersJSON))
		argCount++
	}

	if updateReq.Body != nil {
		updates = append(updates, "body = $"+strconv.Itoa(argCount))
		args = append(args, *updateReq.Body)
		argCount++
	}

	if updateReq.ExtractionRules != nil {
		extractionRulesJSON, err := json.Marshal(updateReq.ExtractionRules)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_extraction_rules",
				Message: "Failed to serialize extraction_rules",
			})
			return
		}
		updates = append(updates, "extraction_rules = $"+strconv.Itoa(argCount))
		args = append(args, string(extractionRulesJSON))
		argCount++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "no_updates",
			Message: "No fields provided to update",
		})
		return
	}

	// Add updated_at timestamp
	updates = append(updates, "updated_at = NOW()")

	// Add item ID as final argument
	args = append(args, itemID)

	// Execute update
	query := "UPDATE collection_items SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated successfully",
	})
}

// DeleteItem handles DELETE /items/:id
func (h *ItemHandler) DeleteItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Item ID must be a valid integer",
		})
		return
	}

	// Check if item exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM collection_items WHERE id = $1)", itemID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to check item existence",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Item not found",
		})
		return
	}

	// Delete item (cascading will handle children)
	_, err = h.db.Exec("DELETE FROM collection_items WHERE id = $1", itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
	})
}

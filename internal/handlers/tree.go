package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"postman-runner/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCollectionTree handles GET /collections/:id/tree
func (h *CollectionHandler) GetCollectionTree(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Collection ID must be a valid integer",
		})
		return
	}

	// Get collection info
	var collection models.Collection
	err = h.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM collections
		WHERE id = $1
	`, collectionID).Scan(
		&collection.ID,
		&collection.Name,
		&collection.Description,
		&collection.CreatedAt,
		&collection.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Collection not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch collection",
		})
		return
	}

	// Get all items using recursive CTE
	rows, err := h.db.Query(`
		WITH RECURSIVE item_tree AS (
			-- Base case: root level items (no parent)
			SELECT 
				id, collection_id, parent_id, name, item_type, 
				sort_order, method, url, headers, body, extraction_rules,
				ARRAY[sort_order] as path
			FROM collection_items
			WHERE collection_id = $1 AND parent_id IS NULL
			
			UNION ALL
			
			-- Recursive case: child items
			SELECT 
				ci.id, ci.collection_id, ci.parent_id, ci.name, ci.item_type,
				ci.sort_order, ci.method, ci.url, ci.headers, ci.body, ci.extraction_rules,
				it.path || ci.sort_order
			FROM collection_items ci
			INNER JOIN item_tree it ON ci.parent_id = it.id
		)
		SELECT 
			id, parent_id, name, item_type, sort_order, 
			method, url, headers, body, extraction_rules
		FROM item_tree
		ORDER BY path
	`, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch collection items",
		})
		return
	}
	defer rows.Close()

	// Build flat list first
	var flatItems []models.CollectionItem
	for rows.Next() {
		var item models.CollectionItem
		var extractionRulesJSON []byte
		err := rows.Scan(
			&item.ID,
			&item.ParentID,
			&item.Name,
			&item.ItemType,
			&item.SortOrder,
			&item.Method,
			&item.URL,
			&item.Headers,
			&item.Body,
			&extractionRulesJSON,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to scan collection item",
			})
			return
		}

		// Parse extraction_rules
		if err := json.Unmarshal(extractionRulesJSON, &item.ExtractionRules); err != nil {
			item.ExtractionRules = []models.ExtractionRule{}
		}

		flatItems = append(flatItems, item)
	}

	// Convert to tree structure
	tree := buildTree(flatItems)

	response := models.CollectionTree{
		Collection: collection,
		Items:      tree,
	}

	c.JSON(http.StatusOK, response)
}

func buildTree(items []models.CollectionItem) []models.ItemTreeNode {
	// Create a map for quick lookup - store pointers
	nodeMap := make(map[int]*models.ItemTreeNode)
	childrenMap := make(map[int][]int) // parent_id -> list of child IDs
	var rootIDs []int

	// First pass: create all nodes and track parent-child relationships
	for _, item := range items {
		node := models.ItemTreeNode{
			ID:        item.ID,
			Name:      item.Name,
			ItemType:  item.ItemType,
			SortOrder: item.SortOrder,
			Children:  []models.ItemTreeNode{},
		}

		// Add request-specific fields
		if item.ItemType == "request" {
			if item.Method.Valid {
				node.Method = item.Method.String
			}
			if item.URL.Valid {
				node.URL = item.URL.String
			}
			if item.Headers.Valid {
				node.Headers = item.Headers.String
			}
			if item.Body.Valid {
				node.Body = item.Body.String
			}
		}

		// Add extraction rules
		node.ExtractionRules = item.ExtractionRules

		nodeMap[item.ID] = &node

		if !item.ParentID.Valid || item.ParentID.Int64 == 0 {
			rootIDs = append(rootIDs, item.ID)
		} else {
			parentID := int(item.ParentID.Int64)
			childrenMap[parentID] = append(childrenMap[parentID], item.ID)
		}
	}

	// Recursive function to build node with children
	var buildNode func(id int) models.ItemTreeNode
	buildNode = func(id int) models.ItemTreeNode {
		node := *nodeMap[id]
		node.Children = []models.ItemTreeNode{}

		if childIDs, ok := childrenMap[id]; ok {
			for _, childID := range childIDs {
				node.Children = append(node.Children, buildNode(childID))
			}
		}
		return node
	}

	// Build root nodes with all nested children
	var rootNodes []models.ItemTreeNode
	for _, id := range rootIDs {
		rootNodes = append(rootNodes, buildNode(id))
	}

	return rootNodes
}

// ListCollections handles GET /collections
func (h *CollectionHandler) ListCollections(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM collections
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch collections",
		})
		return
	}
	defer rows.Close()

	var collections []models.Collection
	for rows.Next() {
		var col models.Collection
		err := rows.Scan(&col.ID, &col.Name, &col.Description, &col.CreatedAt, &col.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to scan collection",
			})
			return
		}
		collections = append(collections, col)
	}

	if collections == nil {
		collections = []models.Collection{}
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": collections,
	})
}

// GetItem handles GET /items/:id
func (h *CollectionHandler) GetItem(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Item ID must be a valid integer",
		})
		return
	}

	var item models.CollectionItem
	var extractionRulesJSON []byte
	err = h.db.QueryRow(`
		SELECT 
			id, collection_id, parent_id, name, item_type, 
			sort_order, method, url, headers, body, extraction_rules, created_at, updated_at
		FROM collection_items
		WHERE id = $1
	`, itemID).Scan(
		&item.ID,
		&item.CollectionID,
		&item.ParentID,
		&item.Name,
		&item.ItemType,
		&item.SortOrder,
		&item.Method,
		&item.URL,
		&item.Headers,
		&item.Body,
		&extractionRulesJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

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

	// Parse headers if present
	var headers interface{}
	if item.Headers.Valid && item.Headers.String != "" {
		if err := json.Unmarshal([]byte(item.Headers.String), &headers); err == nil {
			// Successfully parsed
		}
	}

	// Parse extraction_rules
	if err := json.Unmarshal(extractionRulesJSON, &item.ExtractionRules); err != nil {
		item.ExtractionRules = []models.ExtractionRule{}
	}

	response := gin.H{
		"id":            item.ID,
		"collection_id": item.CollectionID,
		"name":          item.Name,
		"item_type":     item.ItemType,
		"sort_order":    item.SortOrder,
		"created_at":    item.CreatedAt,
		"updated_at":    item.UpdatedAt,
	}

	if item.ParentID.Valid {
		response["parent_id"] = item.ParentID.Int64
	}

	if item.ItemType == "request" {
		if item.Method.Valid {
			response["method"] = item.Method.String
		}
		if item.URL.Valid {
			response["url"] = item.URL.String
		}
		if headers != nil {
			response["headers"] = headers
		}
		if item.Body.Valid {
			response["body"] = item.Body.String
		}
		response["extraction_rules"] = item.ExtractionRules
	}

	c.JSON(http.StatusOK, response)
}

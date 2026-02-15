package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"postman-runner/internal/config"
	"postman-runner/internal/models"

	"github.com/gin-gonic/gin"
)

type EnvironmentHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewEnvironmentHandler(db *sql.DB, cfg *config.Config) *EnvironmentHandler {
	return &EnvironmentHandler{
		db:  db,
		cfg: cfg,
	}
}

// CreateEnvironmentRequest represents the request body for creating an environment
type CreateEnvironmentRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	CreatedBy   string            `json:"created_by"`
	Variables   map[string]string `json:"variables"`
}

// UpdateEnvironmentRequest represents the request body for updating an environment
type UpdateEnvironmentRequest struct {
	Name        *string            `json:"name"`
	Description *string            `json:"description"`
	CreatedBy   *string            `json:"created_by"`
	Variables   *map[string]string `json:"variables"`
}

// BatchUpdateVariablesRequest represents the request body for batch updating environment variables
type BatchUpdateVariablesRequest struct {
	Variables map[string]string `json:"variables" binding:"required"`
}

// CreateEnvironment handles POST /environments
func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
	var req CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Name is required",
		})
		return
	}

	// Default to empty variables if not provided
	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	variablesJSON, err := json.Marshal(req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "json_error",
			Message: "Failed to encode variables",
		})
		return
	}

	var env models.Environment
	err = h.db.QueryRow(`
		INSERT INTO environments (name, description, created_by, variables)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, created_by, variables, created_at, updated_at
	`, req.Name, req.Description, req.CreatedBy, variablesJSON).Scan(
		&env.ID,
		&env.Name,
		&env.Description,
		&env.CreatedBy,
		&variablesJSON,
		&env.CreatedAt,
		&env.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create environment",
		})
		return
	}

	// Parse variables back from JSON
	if err := json.Unmarshal(variablesJSON, &env.Variables); err != nil {
		env.Variables = make(map[string]string)
	}

	c.JSON(http.StatusCreated, env)
}

// ListEnvironments handles GET /environments
func (h *EnvironmentHandler) ListEnvironments(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, description, created_by, variables, created_at, updated_at
		FROM environments
		ORDER BY name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch environments",
		})
		return
	}
	defer rows.Close()

	environments := []models.Environment{}
	for rows.Next() {
		var env models.Environment
		var description, createdBy sql.NullString
		var variablesJSON []byte

		if err := rows.Scan(
			&env.ID,
			&env.Name,
			&description,
			&createdBy,
			&variablesJSON,
			&env.CreatedAt,
			&env.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to scan environment",
			})
			return
		}

		env.Description = description.String
		env.CreatedBy = createdBy.String

		// Parse variables from JSON
		if err := json.Unmarshal(variablesJSON, &env.Variables); err != nil {
			env.Variables = make(map[string]string)
		}

		environments = append(environments, env)
	}

	c.JSON(http.StatusOK, environments)
}

// GetEnvironment handles GET /environments/:id
func (h *EnvironmentHandler) GetEnvironment(c *gin.Context) {
	id := c.Param("id")

	var env models.Environment
	var description, createdBy sql.NullString
	var variablesJSON []byte

	err := h.db.QueryRow(`
		SELECT id, name, description, created_by, variables, created_at, updated_at
		FROM environments
		WHERE id = $1
	`, id).Scan(
		&env.ID,
		&env.Name,
		&description,
		&createdBy,
		&variablesJSON,
		&env.CreatedAt,
		&env.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Environment not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch environment",
		})
		return
	}

	env.Description = description.String
	env.CreatedBy = createdBy.String

	// Parse variables from JSON
	if err := json.Unmarshal(variablesJSON, &env.Variables); err != nil {
		env.Variables = make(map[string]string)
	}

	c.JSON(http.StatusOK, env)
}

// UpdateEnvironment handles PUT /environments/:id
func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
	id := c.Param("id")

	var req UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
		})
		return
	}

	// First, fetch the existing environment
	var env models.Environment
	var description, createdBy sql.NullString
	var variablesJSON []byte

	err := h.db.QueryRow(`
		SELECT id, name, description, created_by, variables, created_at, updated_at
		FROM environments
		WHERE id = $1
	`, id).Scan(
		&env.ID,
		&env.Name,
		&description,
		&createdBy,
		&variablesJSON,
		&env.CreatedAt,
		&env.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Environment not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch environment",
		})
		return
	}

	env.Description = description.String
	env.CreatedBy = createdBy.String

	// Parse existing variables
	if err := json.Unmarshal(variablesJSON, &env.Variables); err != nil {
		env.Variables = make(map[string]string)
	}

	// Apply updates
	if req.Name != nil {
		env.Name = *req.Name
	}
	if req.Description != nil {
		env.Description = *req.Description
	}
	if req.CreatedBy != nil {
		env.CreatedBy = *req.CreatedBy
	}
	if req.Variables != nil {
		env.Variables = *req.Variables
	}

	// Marshal variables back to JSON
	updatedVariablesJSON, err := json.Marshal(env.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "json_error",
			Message: "Failed to encode variables",
		})
		return
	}

	// Update in database
	err = h.db.QueryRow(`
		UPDATE environments
		SET name = $1, description = $2, created_by = $3, variables = $4, updated_at = $5
		WHERE id = $6
		RETURNING updated_at
	`, env.Name, env.Description, env.CreatedBy, updatedVariablesJSON, time.Now(), id).Scan(&env.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update environment",
		})
		return
	}

	c.JSON(http.StatusOK, env)
}

// DeleteEnvironment handles DELETE /environments/:id
func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	id := c.Param("id")

	result, err := h.db.Exec(`DELETE FROM environments WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete environment",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Environment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment deleted successfully"})
}

// BatchUpdateEnvironmentVariables handles PATCH /environments/:id/variables
func (h *EnvironmentHandler) BatchUpdateEnvironmentVariables(c *gin.Context) {
	id := c.Param("id")

	var req BatchUpdateVariablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Variables are required",
		})
		return
	}

	// Check if environment exists and get current variables
	var currentVarsJSON []byte
	err := h.db.QueryRow("SELECT variables FROM environments WHERE id = $1", id).Scan(&currentVarsJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Environment not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch environment",
		})
		return
	}

	// Parse current variables
	var currentVars map[string]string
	if err := json.Unmarshal(currentVarsJSON, &currentVars); err != nil {
		currentVars = make(map[string]string)
	}

	// Merge variables (update existing, add new)
	for key, value := range req.Variables {
		currentVars[key] = value
	}

	// Update in database
	updatedVarsJSON, err := json.Marshal(currentVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "json_error",
			Message: "Failed to encode variables",
		})
		return
	}

	_, err = h.db.Exec(`
		UPDATE environments 
		SET variables = $1, updated_at = NOW() 
		WHERE id = $2
	`, updatedVarsJSON, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update variables",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Variables updated successfully",
		"variables": currentVars,
	})
}

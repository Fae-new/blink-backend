package models

import (
	"database/sql"
	"time"
)

type Collection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CollectionItem struct {
	ID              int              `json:"id"`
	CollectionID    int              `json:"collection_id"`
	ParentID        sql.NullInt64    `json:"parent_id,omitempty"`
	Name            string           `json:"name"`
	ItemType        string           `json:"item_type"` // "folder" or "request"
	SortOrder       int              `json:"sort_order"`
	Method          sql.NullString   `json:"method,omitempty"`
	URL             sql.NullString   `json:"url,omitempty"`
	Headers         sql.NullString   `json:"headers,omitempty"` // JSONB as string
	Body            sql.NullString   `json:"body,omitempty"`
	ExtractionRules []ExtractionRule `json:"extraction_rules,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// Postman Collection Schema (simplified)
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

type PostmanInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

type PostmanItem struct {
	Name    string          `json:"name"`
	Item    []PostmanItem   `json:"item,omitempty"`    // For folders
	Request *PostmanRequest `json:"request,omitempty"` // For requests
}

type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header,omitempty"`
	Body   *PostmanBody    `json:"body,omitempty"`
	URL    interface{}     `json:"url"` // Can be string or object
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw,omitempty"`
}

// Response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ExecutionResponse struct {
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	DurationMs int64             `json:"duration_ms"`
}

// Tree structure for collection retrieval
type CollectionTree struct {
	Collection Collection     `json:"collection"`
	Items      []ItemTreeNode `json:"items"`
}

type ItemTreeNode struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	ItemType        string           `json:"item_type"`
	SortOrder       int              `json:"sort_order"`
	Method          string           `json:"method,omitempty"`
	URL             string           `json:"url,omitempty"`
	Headers         string           `json:"headers,omitempty"`
	Body            string           `json:"body,omitempty"`
	ExtractionRules []ExtractionRule `json:"extraction_rules,omitempty"`
	Children        []ItemTreeNode   `json:"children,omitempty"`
}

// Environment represents a set of variables for request execution
type Environment struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	CreatedBy   string            `json:"created_by,omitempty"`
	Variables   map[string]string `json:"variables"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// EnvironmentVariable represents a single key-value pair in an environment
type EnvironmentVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExtractionRule represents a rule for extracting values from API responses
type ExtractionRule struct {
	Enabled      bool   `json:"enabled"`
	JSONPath     string `json:"json_path"`
	VariableName string `json:"variable_name"`
}

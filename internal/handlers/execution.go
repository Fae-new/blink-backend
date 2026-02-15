package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"postman-runner/internal/config"
	"postman-runner/internal/models"
	"postman-runner/internal/validator"

	"github.com/gin-gonic/gin"
)

type ExecutionHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewExecutionHandler(db *sql.DB, cfg *config.Config) *ExecutionHandler {
	return &ExecutionHandler{
		db:  db,
		cfg: cfg,
	}
}

// ExecutionRequest represents the optional request body for execution
type ExecutionRequest struct {
	URL     *string            `json:"url,omitempty"`
	Headers *map[string]string `json:"headers,omitempty"`
	Body    *string            `json:"body,omitempty"`
}

// ExecuteRequest handles POST /items/:id/execute
func (h *ExecutionHandler) ExecuteRequest(c *gin.Context) {
	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Item ID must be a valid integer",
		})
		return
	}

	// Fetch item from database
	var item models.CollectionItem
	err = h.db.QueryRow(`
		SELECT id, name, item_type, method, url, headers, body
		FROM collection_items
		WHERE id = $1
	`, itemID).Scan(
		&item.ID,
		&item.Name,
		&item.ItemType,
		&item.Method,
		&item.URL,
		&item.Headers,
		&item.Body,
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

	// Validate item type
	if item.ItemType != "request" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_item_type",
			Message: "Only items of type 'request' can be executed",
		})
		return
	}

	// Parse optional request body for overrides (URL with variables replaced, etc.)
	var execReq ExecutionRequest
	if c.Request.ContentLength > 0 {
		// Use a more lenient JSON decoder that ignores unknown fields
		var rawBody map[string]interface{}
		if err := c.ShouldBindJSON(&rawBody); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("Invalid request body: %v", err),
			})
			return
		}

		// Extract only the fields we care about
		if urlVal, ok := rawBody["url"].(string); ok {
			execReq.URL = &urlVal
		}
		if bodyVal, ok := rawBody["body"].(string); ok {
			execReq.Body = &bodyVal
		}
		if headersVal, ok := rawBody["headers"].(map[string]interface{}); ok {
			headers := make(map[string]string)
			for k, v := range headersVal {
				if strVal, ok := v.(string); ok {
					headers[k] = strVal
				}
			}
			if len(headers) > 0 {
				execReq.Headers = &headers
			}
		}
	}

	// Extract request details with overrides
	if !item.Method.Valid {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Request is missing method",
		})
		return
	}

	method := item.Method.String

	// Use overridden URL if provided, otherwise use stored URL
	urlStr := ""
	if execReq.URL != nil {
		urlStr = *execReq.URL
	} else if item.URL.Valid {
		urlStr = item.URL.String
	}

	if urlStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Request URL is missing. Provide URL in request body or ensure item has a URL.",
		})
		return
	}

	// Perform SSRF validation
	if err := validator.ValidateExecutionURL(urlStr, h.cfg.AllowLocalhost, h.cfg.AllowPrivateIPs); err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "ssrf_protection",
			Message: fmt.Sprintf("URL blocked by SSRF protection: %v", err),
		})
		return
	}

	// Parse headers with optional overrides
	headers := make(map[string]string)
	if execReq.Headers != nil {
		// Use overridden headers from request body
		headers = *execReq.Headers
	} else if item.Headers.Valid && item.Headers.String != "" {
		// Use stored headers
		var postmanHeaders []models.PostmanHeader
		if err := json.Unmarshal([]byte(item.Headers.String), &postmanHeaders); err == nil {
			for _, h := range postmanHeaders {
				headers[h.Key] = h.Value
			}
		}
	}

	// Get body with optional override
	body := ""
	if execReq.Body != nil {
		body = *execReq.Body
	} else if item.Body.Valid {
		body = item.Body.String
	}

	// Execute request
	startTime := time.Now()
	response, err := h.executeHTTPRequest(method, urlStr, headers, body)
	duration := time.Since(startTime)

	if err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse{
			Error:   "execution_error",
			Message: fmt.Sprintf("Failed to execute request: %v", err),
		})
		return
	}

	response.DurationMs = duration.Milliseconds()
	c.JSON(http.StatusOK, response)
}

func (h *ExecutionHandler) executeHTTPRequest(method, urlStr string, headers map[string]string, bodyStr string) (*models.ExecutionResponse, error) {
	// Create HTTP client with timeout and redirect limit
	client := &http.Client{
		Timeout: h.cfg.RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= h.cfg.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", h.cfg.MaxRedirects)
			}
			// Validate redirect URL for SSRF
			if err := validator.ValidateExecutionURL(req.URL.String(), h.cfg.AllowLocalhost, h.cfg.AllowPrivateIPs); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}

	// Prepare request body (only for methods that support bodies)
	var bodyReader io.Reader
	methodUpper := strings.ToUpper(method)
	supportsBody := methodUpper == "POST" || methodUpper == "PUT" || methodUpper == "PATCH" || methodUpper == "DELETE"

	if bodyStr != "" && supportsBody {
		bodyReader = bytes.NewReader([]byte(bodyStr))
	}

	// Create request
	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, h.cfg.MaxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Extract response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return &models.ExecutionResponse{
		Status:  resp.StatusCode,
		Headers: responseHeaders,
		Body:    string(responseBody),
	}, nil
}

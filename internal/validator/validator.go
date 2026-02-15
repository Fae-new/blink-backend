package validator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"postman-runner/internal/models"
)

// ValidatePostmanCollection validates the entire Postman collection
func ValidatePostmanCollection(data []byte, maxRequestSize int64, maxHeaderCount int) (*models.PostmanCollection, error) {
	if int64(len(data)) > maxRequestSize {
		return nil, fmt.Errorf("collection JSON exceeds maximum size of %d bytes", maxRequestSize)
	}

	var collection models.PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	// Validate schema version
	if !isValidSchema(collection.Info.Schema) {
		return nil, fmt.Errorf("unsupported Postman schema version: %s (only v2.0 and v2.1 are supported)", collection.Info.Schema)
	}

	// Validate collection has items
	if len(collection.Item) == 0 {
		return nil, fmt.Errorf("collection must contain at least one item")
	}

	// Recursively validate all items
	if err := validateItems(collection.Item, maxHeaderCount); err != nil {
		return nil, err
	}

	return &collection, nil
}

func isValidSchema(schema string) bool {
	validSchemas := []string{
		"https://schema.getpostman.com/json/collection/v2.0.0/collection.json",
		"https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		"https://schema.getpostman.com/json/collection/v2.0",
		"https://schema.getpostman.com/json/collection/v2.1",
	}
	for _, valid := range validSchemas {
		if schema == valid {
			return true
		}
	}
	return false
}

func validateItems(items []models.PostmanItem, maxHeaderCount int) error {
	for _, item := range items {
		// Check if it's a folder
		if len(item.Item) > 0 {
			// Recursively validate folder items
			if err := validateItems(item.Item, maxHeaderCount); err != nil {
				return err
			}
		} else if item.Request != nil {
			// Validate request
			if err := validateRequest(item.Request, maxHeaderCount); err != nil {
				return fmt.Errorf("invalid request '%s': %w", item.Name, err)
			}
		}
		// Allow items without requests (e.g., placeholder items or items with only saved responses)
	}
	return nil
}

func validateRequest(req *models.PostmanRequest, maxHeaderCount int) error {
	// Validate HTTP method
	allowedMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}

	if !allowedMethods[strings.ToUpper(req.Method)] {
		return fmt.Errorf("unsupported HTTP method: %s (allowed: GET, POST, PUT, PATCH, DELETE)", req.Method)
	}

	// Validate URL (if present)
	urlStr, err := extractURL(req.URL)
	if err != nil {
		// Allow requests with missing or invalid URL format - they can be filled in later
		// or use environment variables
		urlStr = ""
	}

	// Only validate URL scheme if URL is present and not using template variables
	if urlStr != "" {
		if err := validateURL(urlStr); err != nil {
			return err
		}
	}

	// Validate header count
	if len(req.Header) > maxHeaderCount {
		return fmt.Errorf("request has %d headers, exceeding limit of %d", len(req.Header), maxHeaderCount)
	}

	return nil
}

func extractURL(urlData interface{}) (string, error) {
	switch v := urlData.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		if raw, ok := v["raw"].(string); ok {
			return raw, nil
		}
		return "", fmt.Errorf("URL object missing 'raw' field")
	default:
		return "", fmt.Errorf("invalid URL type")
	}
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Skip scheme validation if URL contains template variables (e.g., {{base_url}})
	// The scheme will be included in the environment variable value
	if strings.Contains(urlStr, "{{") && strings.Contains(urlStr, "}}") {
		return nil
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow http and https
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s (only http and https are allowed)", parsed.Scheme)
	}

	return nil
}

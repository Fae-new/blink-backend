package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultTimeout      = 30 * time.Second
	MaxResponseSize     = 50 * 1024 * 1024 // 50MB
	MaxRedirects        = 5
)

func executeHandler(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Execute the HTTP request
	response, err := executeHTTPRequest(req)
	if err != nil {
		c.JSON(http.StatusOK, ExecuteResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func executeHTTPRequest(execReq ExecuteRequest) (*ExecuteResponse, error) {
	startTime := time.Now()

	// Create HTTP client with timeout and redirect limit
	client := &http.Client{
		Timeout: DefaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", MaxRedirects)
			}
			return nil
		},
	}

	// Prepare request body
	var bodyReader io.Reader
	methodUpper := strings.ToUpper(execReq.Method)
	supportsBody := methodUpper == "POST" || methodUpper == "PUT" || methodUpper == "PATCH" || methodUpper == "DELETE"

	if execReq.Body != "" && supportsBody {
		bodyReader = bytes.NewReader([]byte(execReq.Body))
	}

	// Create request
	req, err := http.NewRequest(execReq.Method, execReq.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range execReq.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Extract response headers
	responseHeaders := make(map[string][]string)
	for key, values := range resp.Header {
		responseHeaders[key] = values
	}

	return &ExecuteResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    responseHeaders,
		Body:       string(responseBody),
		Timing: TimingInfo{
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  duration.Milliseconds(),
		},
	}, nil
}

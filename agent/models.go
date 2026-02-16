package main

import (
	"time"
)

// ExecuteRequest represents the request payload from the frontend
type ExecuteRequest struct {
	Method  string            `json:"method" binding:"required"`
	URL     string            `json:"url" binding:"required"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// ExecuteResponse represents the response sent back to the frontend
type ExecuteResponse struct {
	Status     int                 `json:"status"`
	StatusText string              `json:"statusText"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Timing     TimingInfo          `json:"timing"`
	Error      string              `json:"error,omitempty"`
}

// TimingInfo contains timing information about the request
type TimingInfo struct {
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Duration     int64     `json:"duration"` // milliseconds
	DNSLookup    int64     `json:"dnsLookup,omitempty"`
	TCPConnection int64    `json:"tcpConnection,omitempty"`
	TLSHandshake int64     `json:"tlsHandshake,omitempty"`
}

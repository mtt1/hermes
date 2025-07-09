// Package ai - error definitions
package ai

import (
	"fmt"
)

// No client-side errors - we validate API key presence before creating clients

// APIError represents an error from the AI API
// We pass through the actual error message from the API
type APIError struct {
	Provider   string // AI provider name (e.g., "gemini")
	StatusCode int    // HTTP status code
	Message    string // Raw error message from API
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s API error: %s", e.Provider, e.Message)
}

// NetworkError represents a network-related error
type NetworkError struct {
	Provider string // AI provider name
	Err      error  // Underlying network error
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("%s network error: %v", e.Provider, e.Err)
}

func (e NetworkError) Unwrap() error {
	return e.Err
}
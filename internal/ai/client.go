// Package ai provides AI client interface and implementations for hermes
package ai

import (
	"context"
	"fmt"
	"hermes/internal/safety"
)

// GenerateRequest represents a request for command generation
type GenerateRequest struct {
	Query string // Natural language query from user
}

// GenerateResponse represents the response from AI command generation
type GenerateResponse struct {
	Command     string              // Generated shell command
	SafetyLevel safety.SafetyLevel  // AI's assessment of command safety
	Reasoning   string              // Optional explanation of the generated command (for --explain-generation flag)
}

// ExplainRequest represents a request for command explanation
type ExplainRequest struct {
	Command string // Shell command to explain
}

// ExplainResponse represents the response from AI command explanation
type ExplainResponse struct {
	Explanation string // Human-readable explanation of the command
}

// Client interface defines the contract for AI providers
type Client interface {
	// GenerateCommand generates a shell command from natural language
	GenerateCommand(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	
	// ExplainCommand explains what a shell command does
	ExplainCommand(ctx context.Context, req ExplainRequest) (*ExplainResponse, error)
	
	// Close cleans up any resources used by the client
	Close() error
}

// Config holds configuration for AI clients
type Config struct {
	APIKey string // API key for the AI provider
	Model  string // Model name to use (optional)
	Debug  bool   // Enable debug logging
}

// NewClient creates a new AI client based on the provider type
func NewClient(provider string, config Config) (Client, error) {
	switch provider {
	case "gemini":
		return NewGeminiClient(config)
	case "mock":
		return NewMockClient(config)
	default:
		// This should never happen since we control the provider parameter
		return nil, fmt.Errorf("internal error: unknown provider %s", provider)
	}
}
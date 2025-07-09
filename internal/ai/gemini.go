// Package ai - Gemini API client
package ai

import (
	"context"
	"hermes/internal/safety"
)

// GeminiClient implements the Client interface for Google's Gemini API
type GeminiClient struct {
	config Config
	// TODO: Add HTTP client and API endpoint fields
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(config Config) (*GeminiClient, error) {
	// API key presence is validated before creating the client
	return &GeminiClient{
		config: config,
	}, nil
}

// GenerateCommand generates a shell command from natural language
func (g *GeminiClient) GenerateCommand(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// TODO: Task 12 - Implement actual Gemini API call
	// For now, return a placeholder
	return &GenerateResponse{
		Command:     "echo 'Gemini integration not yet implemented'",
		SafetyLevel: safety.Safe,
		Reasoning:   "Placeholder response - Gemini API integration pending",
	}, nil
}

// ExplainCommand explains what a shell command does
func (g *GeminiClient) ExplainCommand(ctx context.Context, req ExplainRequest) (*ExplainResponse, error) {
	// TODO: Task 12 - Implement actual Gemini API call
	// For now, return a placeholder
	return &ExplainResponse{
		Explanation: "Gemini explanation not yet implemented",
	}, nil
}

// Close cleans up any resources used by the client
func (g *GeminiClient) Close() error {
	// TODO: Task 12 - Clean up HTTP client if needed
	return nil
}
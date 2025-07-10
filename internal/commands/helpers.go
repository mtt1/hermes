// Package commands contains shared helper functions for CLI command definitions
package commands

import (
	"hermes/internal/ai"
	"hermes/internal/config"
	"hermes/internal/exit"
)

// createAIClient is a factory function that creates an AI client based on app config.
// It abstracts away the logic of choosing between the real Gemini client and the mock client.
func createAIClient(cfg *config.Config) (ai.Client, error) {
	var provider string
	var apiKey string

	// Determine the provider and API key based on the configuration.
	// The mock client is used for testing and development.
	if cfg.MockResponse != "" {
		provider = "mock"
		apiKey = "mock-key" // The mock client doesn't require a real key.
	} else {
		provider = "gemini"
		apiKey = cfg.GeminiAPIKey
	}

	// Create the new AI client using the determined provider.
	client, err := ai.NewClient(provider, ai.Config{
		APIKey: apiKey,
		Debug:  cfg.Debug,
	})

	// If client creation fails, return a structured error.
	if err != nil {
		return nil, exit.NewError(exit.CodeError, "Failed to create AI client: %v", err)
	}

	return client, nil
}

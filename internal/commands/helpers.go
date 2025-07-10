// Package commands contains shared helper functions for CLI command definitions
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"hermes/internal/ai"
	"hermes/internal/config"
	"hermes/internal/exit"
)

// createAIClient is a factory function that creates an AI client based on app config.
// It abstracts away the logic of choosing between the real Gemini client and the mock client.
// It also handles API key validation and debug logging in one place.
func createAIClient(cfg *config.Config) (ai.Client, error) {
	// Validate API key is available (unless using mock)
	if cfg.GeminiAPIKey == "" && cfg.MockResponse == "" {
		return nil, exit.NewError(exit.CodeConfig, "Gemini API key is required. Set it via (in priority order):\n"+
			"  - CLI flag: --gemini-api-key\n"+
			"  - Environment variable: GEMINI_API_KEY\n"+
			"  - Config file: ~/.config/hermes/config.toml")
	}

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

	// Debug logging for API key (centralized)
	if cfg.Debug {
		if apiKey == "mock-key" {
			fmt.Printf("DEBUG: Using mock AI client\n")
		} else if len(apiKey) > 4 {
			fmt.Printf("DEBUG: Using API key ending in ...%s\n", apiKey[len(apiKey)-4:])
		} else {
			fmt.Printf("DEBUG: Using API key (too short to truncate)\n")
		}
	}

	// Create the new AI client using the determined provider.
	client, err := ai.NewClient(provider, ai.Config{
		APIKey:       apiKey,
		Debug:        cfg.Debug,
		MockResponse: cfg.MockResponse,
	})

	// If client creation fails, return a structured error.
	if err != nil {
		return nil, exit.NewError(exit.CodeError, "Failed to create AI client: %v", err)
	}

	return client, nil
}

// checkShellIntegration detects if hermes shell integration is active and warns if not
func checkShellIntegration() {
	// Check if we're running from the hermes shell function
	// The shell function sets HERMES_SHELL_INTEGRATION=1 when calling hermes
	if os.Getenv("HERMES_SHELL_INTEGRATION") == "1" {
		return // Shell integration is active
	}
	
	// Check if user wants to suppress the tip
	if os.Getenv("HERMES_SUPPRESS_INTEGRATION_TIP") == "1" {
		return // User has chosen to suppress the tip
	}
	
	// Check if we're in an interactive shell that could benefit from integration
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return // No shell detected, probably running in a script
	}
	
	// Only show tip for shells we actually support
	shellName := filepath.Base(shellPath)
	switch shellName {
	case "zsh":
		// Show integration hint for supported shell
		fmt.Fprintf(os.Stderr, "\n   TIP: Enable shell integration for the best experience!\n")
		fmt.Fprintf(os.Stderr, "   Run: hermes init zsh >> ~/.zshrc && source ~/.zshrc\n")
		fmt.Fprintf(os.Stderr, "   This allows hermes to put commands directly in your shell buffer.\n")
		fmt.Fprintf(os.Stderr, "   To suppress this tip: export HERMES_SUPPRESS_INTEGRATION_TIP=1\n\n")
	default:
		// For unsupported shells, show no tip (future expansion point)
		return
	}
}

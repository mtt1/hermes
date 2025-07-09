// Package config handles configuration management for hermes
package config

import (
	"github.com/knadh/koanf/v2"
)

// Global Koanf instance
var K = koanf.New(".")

// Config holds all configuration for the application
type Config struct {
	GeminiAPIKey  string `koanf:"gemini_api_key" mapstructure:"gemini_api_key"`
	Debug         bool   `koanf:"debug" mapstructure:"debug"`
	MockResponse  string `koanf:"mock_response" mapstructure:"mock_response"`
	MockExitCode  int    `koanf:"mock_exit_code" mapstructure:"mock_exit_code"`
}

// Default returns a new Config with default values
func Default() Config {
	return Config{
		GeminiAPIKey: "", // No default API key
		Debug:        false,
		MockResponse: "", // No default mock response
		MockExitCode: 0,  // Default to safe exit code
	}
}
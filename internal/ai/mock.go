// Package ai - mock client for testing
package ai

import (
	"context"
	"fmt"
	"hermes/internal/safety"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	config         Config
	staticCommand  string            // Static command from MockResponse flag
	responseMap    map[string]string // Query -> Command mapping
	explanationMap map[string]string // Command -> Explanation mapping
}

// NewMockClient creates a new mock AI client
func NewMockClient(config Config) (*MockClient, error) {
	return &MockClient{
		config:        config,
		staticCommand: config.MockResponse, // Use MockResponse as the static command
		responseMap: map[string]string{
			"list files":        "ls -la",
			"list all files":    "ls -la",
			"delete everything": "rm -rf /",
			"install vim":       "sudo apt install vim",
			"check disk usage":  "df -h",
			"show processes":    "ps aux",
			"find python files": "find . -name '*.py'",
		},
		explanationMap: map[string]string{
			"ls -la":              "List all files and directories in long format, including hidden files",
			"rm -rf /":            "DANGEROUS: Recursively remove all files starting from root directory",
			"sudo apt install vim": "Install vim text editor using apt package manager with sudo privileges",
			"df -h":               "Display filesystem disk usage in human-readable format",
			"ps aux":              "Show all running processes with detailed information",
			"find . -name '*.py'": "Find all Python files in current directory and subdirectories",
		},
	}, nil
}

// GenerateCommand generates a shell command from natural language
func (m *MockClient) GenerateCommand(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if m.config.Debug {
		fmt.Printf("DEBUG: Mock AI generating command for: %s\n", req.Query)
	}
	
	// Prioritize static command from --mock-response flag
	if m.staticCommand != "" {
		// Determine safety level based on command content
		safetyLevel := safety.Safe
		if containsDangerousPatterns(m.staticCommand) {
			safetyLevel = safety.Attention
		}
		
		return &GenerateResponse{
			Command:     m.staticCommand,
			SafetyLevel: safetyLevel,
			Reasoning:   fmt.Sprintf("Mock static response for: %s", req.Query),
		}, nil
	}
	
	// Check if we have a predefined response
	if command, exists := m.responseMap[req.Query]; exists {
		// Determine safety level based on command content
		safetyLevel := safety.Safe
		if containsDangerousPatterns(command) {
			safetyLevel = safety.Attention
		}
		
		return &GenerateResponse{
			Command:     command,
			SafetyLevel: safetyLevel,
			Reasoning:   fmt.Sprintf("Mock reasoning for: %s", req.Query),
		}, nil
	}
	
	// Default response for unknown queries
	return &GenerateResponse{
		Command:     fmt.Sprintf("echo 'Mock command for: %s'", req.Query),
		SafetyLevel: safety.Safe,
		Reasoning:   "Mock default response",
	}, nil
}

// ExplainCommand explains what a shell command does
func (m *MockClient) ExplainCommand(ctx context.Context, req ExplainRequest) (*ExplainResponse, error) {
	if m.config.Debug {
		fmt.Printf("DEBUG: Mock AI explaining command: %s\n", req.Command)
	}

	// Prioritize static response from --mock-response flag
	if m.staticCommand != "" {
		return &ExplainResponse{
			Explanation: m.staticCommand,
		}, nil
	}

	// Check if we have a predefined explanation
	if explanation, exists := m.explanationMap[req.Command]; exists {
		return &ExplainResponse{
			Explanation: explanation,
		}, nil
	}

	// Default explanation for unknown commands
	return &ExplainResponse{
		Explanation: fmt.Sprintf("Mock explanation for command: %s", req.Command),
	}, nil
}

// Close cleans up any resources used by the client
func (m *MockClient) Close() error {
	// Mock client has no resources to clean up
	return nil
}

// containsDangerousPatterns checks if a command contains patterns that need attention
func containsDangerousPatterns(command string) bool {
	dangerousPatterns := []string{
		"rm -rf",
		"sudo",
		"dd",
		"mkfs",
		"fdisk",
		"systemctl start",
		"systemctl stop",
		"apt install",
		"yum install",
		"pacman -S",
	}
	
	for _, pattern := range dangerousPatterns {
		if contains(command, pattern) {
			return true
		}
	}
	
	return false
}

// contains checks if a string contains a substring (simple helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		   len(s) > len(substr) && contains(s[1:], substr)
}

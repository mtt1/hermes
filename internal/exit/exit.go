// Package exit provides custom error types for CLI exit codes
package exit

import "fmt"

// Error represents a CLI error with a specific exit code.
type Error struct {
	Code int
	Err  error
}

func (e Error) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// Success is a special error to signal a clean exit with code 0.
// Useful for commands like --explain or --version that should stop execution.
func Success() Error {
	return Error{Code: 0, Err: nil}
}

// NewError creates a new error with a specific code.
func NewError(code int, format string, a ...interface{}) Error {
	return Error{Code: code, Err: fmt.Errorf(format, a...)}
}

// Exit code constants for hermes
const (
	CodeSuccess   = 0  // Safe command
	CodeError     = 1  // Generic error
	CodeInvalid   = 2  // Invalid input
	CodeNetwork   = 3  // Network error
	CodeAPI       = 4  // API error
	CodeDangerous = 10 // Dangerous command
	CodeSudo      = 11 // Requires sudo
)
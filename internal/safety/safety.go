// Package safety provides binary command safety analysis for hermes
package safety

import (
	"context"
	"regexp"
	"hermes/internal/exit"
)

// SafetyLevel represents the safety level of a command
type SafetyLevel int

const (
	Safe SafetyLevel = iota
	Attention
)

// String returns the string representation of the safety level
func (s SafetyLevel) String() string {
	switch s {
	case Safe:
		return "safe"
	case Attention:
		return "attention"
	default:
		return "unknown"
	}
}

// ExitCode returns the exit code for the safety level
func (s SafetyLevel) ExitCode() int {
	switch s {
	case Safe:
		return exit.CodeSuccess
	case Attention:
		return exit.CodeDangerous // Using exit code 10 for all "attention" cases
	default:
		return exit.CodeError
	}
}

// Result represents the result of safety analysis
type Result struct {
	Level  SafetyLevel
	Reason string
	Layer  string // Which layer made the decision
}

// Analyzer provides binary command safety analysis
type Analyzer struct {
	// Pre-compiled regex patterns for performance
	attentionPatterns []*regexp.Regexp
	safePatterns      []*regexp.Regexp
	
	// AI client will be injected here in Phase 2
	// For now, this is a placeholder for the interface
}

// NewAnalyzer creates a new binary safety analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		// Patterns that require user attention (dangerous, sudo, etc.)
		attentionPatterns: []*regexp.Regexp{
			// Sudo commands (always need attention)
			regexp.MustCompile(`\bsudo\b`),
			
			// Dangerous operations
			regexp.MustCompile(`\brm\s+.*(-[rf]+|--recursive|--force).*\s+/\s*$`), // rm -rf /
			regexp.MustCompile(`\bdd\s+.*of=/dev/sd`),                              // dd to disk
			regexp.MustCompile(`\bmkfs\b`),                                         // format filesystem
			regexp.MustCompile(`\bfdisk\b`),                                        // disk partitioning
			regexp.MustCompile(`\bshred\b`),                                        // secure delete
			regexp.MustCompile(`\bwipe\b`),                                         // secure delete
			regexp.MustCompile(`\bchmod\s+777`),                                    // dangerous permissions
			regexp.MustCompile(`>\s*/dev/sd`),                                      // redirect to disk
			regexp.MustCompile(`\bcurl\s+.*\|\s*sh`),                              // pipe to shell
			regexp.MustCompile(`\bwget\s+.*\|\s*sh`),                              // pipe to shell
			
			// Commands that typically need sudo (even without sudo keyword)
			regexp.MustCompile(`\bsystemctl\s+(start|stop|restart|enable|disable)\b`), // service management
			regexp.MustCompile(`\bapt\s+(install|remove|update|upgrade)\b`),            // package management
			regexp.MustCompile(`\byum\s+(install|remove|update)\b`),                   // package management
			regexp.MustCompile(`\bpacman\s+-S\b`),                                     // package management
			regexp.MustCompile(`\bmodprobe\b`),                                        // kernel modules
			regexp.MustCompile(`\bmount\b`),                                           // mounting
			regexp.MustCompile(`\bumount\b`),                                          // unmounting
			regexp.MustCompile(`\biptables\b`),                                        // firewall
		},
		
		// High-confidence safe patterns (can execute directly)
		safePatterns: []*regexp.Regexp{
			regexp.MustCompile(`^ls\b`),                    // ls commands
			regexp.MustCompile(`^cd\b`),                    // cd commands  
			regexp.MustCompile(`^pwd\b`),                   // pwd command
			regexp.MustCompile(`^echo\b`),                  // echo command
			regexp.MustCompile(`^cat\b`),                   // cat command
			regexp.MustCompile(`^head\b`),                  // head command
			regexp.MustCompile(`^tail\b`),                  // tail command
			regexp.MustCompile(`^grep\b`),                  // grep command
			regexp.MustCompile(`^find\b`),                  // find command
			regexp.MustCompile(`^git\s+(status|log|diff|branch|show)\b`), // safe git commands
			regexp.MustCompile(`^ps\b`),                    // process list
			regexp.MustCompile(`^which\b`),                 // which command
			regexp.MustCompile(`^whereis\b`),               // whereis command
			regexp.MustCompile(`^man\b`),                   // man pages
			regexp.MustCompile(`^help\b`),                  // help command
			regexp.MustCompile(`^systemctl\s+status\b`),    // safe systemctl usage
		},
	}
}

// AnalyzeCommand performs binary safety analysis of a command
func (a *Analyzer) AnalyzeCommand(ctx context.Context, command string) (Result, error) {
	// Layer 1: Check for attention patterns first (dangerous, sudo, etc.)
	for _, pattern := range a.attentionPatterns {
		if pattern.MatchString(command) {
			return Result{
				Level:  Attention,
				Reason: "Command requires user attention",
				Layer:  "attention-patterns",
			}, nil
		}
	}
	
	// Layer 2: Check for safe patterns
	for _, pattern := range a.safePatterns {
		if pattern.MatchString(command) {
			return Result{
				Level:  Safe,
				Reason: "Command is known to be safe",
				Layer:  "safe-patterns",
			}, nil
		}
	}
	
	// Layer 3: AI Analysis (For Ambiguous Cases)
	// TODO: Phase 2 - Implement AI-based safety analysis
	// For now, default to safe for ambiguous cases
	return Result{
		Level:  Safe,
		Reason: "Command passed basic safety checks (AI analysis not yet implemented)",
		Layer:  "default-safe",
	}, nil
}

// MockAnalyzeCommand provides mock safety analysis for testing
// This will be controlled by --mock-exit-code flag
func (a *Analyzer) MockAnalyzeCommand(command string, mockExitCode int) Result {
	switch mockExitCode {
	case exit.CodeSuccess:
		return Result{
			Level:  Safe,
			Reason: "Mock: safe command",
			Layer:  "mock",
		}
	case exit.CodeDangerous:
		return Result{
			Level:  Attention,
			Reason: "Mock: requires attention",
			Layer:  "mock",
		}
	default:
		return Result{
			Level:  Safe,
			Reason: "Mock: default safe",
			Layer:  "mock",
		}
	}
}
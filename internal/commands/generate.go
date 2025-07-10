// Package commands - generate subcommand
package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"hermes/internal/ai"
	"hermes/internal/exit"
	"hermes/internal/safety"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate [natural language query]",
	Aliases: []string{"gen"},
	Short:   "Generate shell commands from natural language",
	Long: `Generate shell commands from natural language descriptions.

This is the primary function of Hermes. Describe what you want to do in natural
language, and Hermes will generate the appropriate shell command.

Usage:
  hermes gen list all files                    # Natural language expressions
  hermes gen "init git repo"                   # Use quotes to enclose expressions for disambiguation
  hermes gen -- init git repo                  # Use delimiter to separate expressions for disambiguation

Examples:
  hermes gen list all files                    # Generate command to list files
  hermes generate delete old log files         # Generate command to delete old logs
  hermes gen find all python files             # Generate command to find Python files
  hermes generate compress this directory      # Generate command to compress directory

Tip: Set up an alias for faster access:
  alias h='hermes gen'
  
Then you can use: h list all files`,

	Args: cobra.MinimumNArgs(1), // Require at least one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate API key is available (unless using mock)
		if appCtx.Config.GeminiAPIKey == "" && appCtx.Config.MockResponse == "" {
			return fmt.Errorf("Gemini API key is required. Set it via:\n" +
				"  - Environment variable: GEMINI_API_KEY\n" +
				"  - CLI flag: --gemini-api-key\n" +
				"  - Config file: ~/.config/hermes/config.toml")
		}

		query := strings.Join(args, " ")
		
		// Show immediate feedback about what we're processing (to stderr)
		fmt.Fprintf(os.Stderr, "└─ Generating command for: '%s'\n", query)
		
		if appCtx.Config.Debug {
			apiKey := appCtx.Config.GeminiAPIKey
			if apiKey == "" {
				fmt.Printf("DEBUG: No API key (using mock)\n")
			} else if len(apiKey) > 4 {
				fmt.Printf("DEBUG: Using API key ending in ...%s\n", apiKey[len(apiKey)-4:])
			} else {
				fmt.Printf("DEBUG: Using API key (too short to truncate)\n")
			}
		}
		
		// Create AI client
		aiClient, err := createAIClient(&appCtx.Config)
		if err != nil {
			return err
		}
		defer aiClient.Close()
		
		// Generate command using AI
		ctx := context.Background()
		response, err := aiClient.GenerateCommand(ctx, ai.GenerateRequest{
			Query: query,
		})
		
		if err != nil {
			return exit.NewError(exit.CodeAPI, "AI command generation failed: %v", err)
		}
		
		generatedCommand := response.Command
		aiSafetyLevel := response.SafetyLevel
		
		// Analyze safety of generated command (hybrid approach)
		analyzer := safety.NewAnalyzer()
		var safetyResult safety.Result
		
		if appCtx.Config.MockExitCode != 0 {
			// Use mock exit code for testing
			safetyResult = analyzer.MockAnalyzeCommand(generatedCommand, appCtx.Config.MockExitCode)
		} else {
			// Use hybrid safety analysis (AI assessment + pattern matching)
			ctx := context.Background()
			result, err := analyzer.AnalyzeCommand(ctx, generatedCommand)
			if err != nil {
				return exit.NewError(exit.CodeError, "Safety analysis failed: %v", err)
			}
			
			// Apply upgrade-only logic: if patterns detected something requiring attention,
			// upgrade the AI's assessment
			if result.Level == safety.Attention {
				safetyResult = result
			} else {
				// AI detected attention but patterns say safe - use AI's assessment
				if aiSafetyLevel == safety.Attention {
					safetyResult = safety.Result{
						Level:  safety.Attention,
						Reason: "AI flagged as requiring attention",
						Layer:  "ai-assessment",
					}
				} else {
					safetyResult = result
				}
			}
		}
		
		// Output only the command (for shell buffer)
		fmt.Printf("%s\n", generatedCommand)
		
		if appCtx.Config.Debug {
			fmt.Printf("DEBUG: Generated command: %s\n", generatedCommand)
			fmt.Printf("DEBUG: Safety level: %s\n", safetyResult.Level)
			fmt.Printf("DEBUG: Safety analysis: %s (reason: %s, layer: %s)\n", 
				safetyResult.Level, safetyResult.Reason, safetyResult.Layer)
		}
		
		// Check for shell integration and warn if not active
		checkShellIntegration()
		
		// Return appropriate exit code
		if safetyResult.Level.ExitCode() != exit.CodeSuccess {
			// Exit with the safety level code - let shell integration handle user messaging
			os.Exit(safetyResult.Level.ExitCode())
		}
		
		return nil
	},
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
		fmt.Printf("\n💡 TIP: Enable shell integration for the best experience!\n")
		fmt.Printf("   Run: hermes init zsh >> ~/.zshrc && source ~/.zshrc\n")
		fmt.Printf("   This allows hermes to put commands directly in your shell buffer.\n")
		fmt.Printf("   To suppress this tip: export HERMES_SUPPRESS_INTEGRATION_TIP=1\n\n")
	default:
		// For unsupported shells, show no tip (future expansion point)
		return
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

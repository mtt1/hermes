// Package commands - generate subcommand
package commands

import (
	"fmt"
	"os"
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
		query := strings.Join(args, " ")
		
		// Show immediate feedback about what we're processing (to stderr)
		fmt.Fprintf(os.Stderr, "└─ Generating command for: '%s'\n", query)
		
		// Create AI client (handles validation and debug logging)
		aiClient, err := createAIClient(&appCtx.Config)
		if err != nil {
			return err
		}
		defer aiClient.Close()
		
		// Generate command using AI
		ctx := cmd.Context()
		response, err := aiClient.GenerateCommand(ctx, ai.GenerateRequest{
			Query: query,
		})
		
		if err != nil {
			return exit.NewError(exit.CodeError, "AI command generation failed: %v", err)
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
			ctx := cmd.Context()
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
		
		// Handle exit code
		if safetyResult.Level.ExitCode() != exit.CodeSuccess {
			// Return clean error for shell integration - no error message, just exit code
			return exit.NewError(safetyResult.Level.ExitCode(), "")
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

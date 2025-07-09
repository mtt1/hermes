// Package commands - generate subcommand
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
		fmt.Printf("Generating command for: '%s'\n", query)
		
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
		
		// TODO: Phase 2 - Implement AI command generation
		// For now, simulate generated command
		var generatedCommand string
		if appCtx.Config.MockResponse != "" {
			generatedCommand = appCtx.Config.MockResponse
		} else {
			generatedCommand = "echo 'AI generation not yet implemented'"
		}
		
		// Analyze safety of generated command
		analyzer := safety.NewAnalyzer()
		var safetyResult safety.Result
		
		if appCtx.Config.MockExitCode != 0 {
			// Use mock exit code for testing
			safetyResult = analyzer.MockAnalyzeCommand(generatedCommand, appCtx.Config.MockExitCode)
		} else {
			// Use real safety analysis
			ctx := context.Background()
			result, err := analyzer.AnalyzeCommand(ctx, generatedCommand)
			if err != nil {
				return exit.NewError(exit.CodeError, "Safety analysis failed: %v", err)
			}
			safetyResult = result
		}
		
		if appCtx.Config.Debug {
			fmt.Printf("DEBUG: Generated command: %s\n", generatedCommand)
			fmt.Printf("DEBUG: Safety analysis: %s (reason: %s, layer: %s)\n", 
				safetyResult.Level, safetyResult.Reason, safetyResult.Layer)
		}
		
		// Output the generated command
		fmt.Printf("Generated command: %s\n", generatedCommand)
		fmt.Printf("Safety level: %s\n", safetyResult.Level)
		
		// Return appropriate exit code
		if safetyResult.Level.ExitCode() != exit.CodeSuccess {
			return exit.NewError(safetyResult.Level.ExitCode(), "Command safety level: %s", safetyResult.Level)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

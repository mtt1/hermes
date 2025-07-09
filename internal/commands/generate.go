// Package commands - generate subcommand
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate [natural language query]",
	Aliases: []string{"gen"},
	Short:   "Generate shell commands from natural language",
	Long: `Generate shell commands from natural language descriptions.

This is the primary function of Hermes. Describe what you want to do in natural
language, and Hermes will generate the appropriate shell command.

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
		// Validate API key is available
		if appCtx.Config.GeminiAPIKey == "" {
			return fmt.Errorf("Gemini API key is required. Set it via:\n" +
				"  - Environment variable: GEMINI_API_KEY\n" +
				"  - CLI flag: --gemini-api-key\n" +
				"  - Config file: ~/.config/hermes/config.toml")
		}

		query := strings.Join(args, " ")
		fmt.Printf("Generating command for: '%s'\n", query)
		
		if appCtx.Config.Debug {
			fmt.Printf("DEBUG: Using API key: %s\n", appCtx.Config.GeminiAPIKey)
		}
		
		// TODO: Implement core command generation logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
// Package commands - explain subcommand
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:     "explain [command]",
	Aliases: []string{"exp"},
	Short:   "Explain what a shell command does",
	Long: `Explain what a shell command does in plain English.

This command takes a shell command and explains what it does, including
breaking down the individual arguments and flags.

Usage:
  hermes exp ls                                # Simple command
  hermes exp "find . -name '*.go'"             # Use quotes for complex commands
  hermes exp -- ls --debug                     # Use delimiter for complex commands

Examples:
  hermes exp ls -la                            # Explain 'ls -la'
  hermes explain "find . -name '*.go'"         # Explain a find command
  hermes exp grep -r "TODO" --include="*.py"   # Explain a complex grep
  hermes explain tar -czf archive.tar.gz dir/  # Explain a tar command

Note: You can use quotes around the command or the delimiter (--)
if the commands contains special characters or flags or you want to be
explicit about the command boundaries.`,

	// Allow unknown flags to be passed through as arguments
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Args:               cobra.MinimumNArgs(1), // Require at least one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate API key is available
		if appCtx.Config.GeminiAPIKey == "" {
			return fmt.Errorf("Gemini API key is required. Set it via:\n" +
				"  - Environment variable: GEMINI_API_KEY\n" +
				"  - CLI flag: --gemini-api-key\n" +
				"  - Config file: ~/.config/hermes/config.toml")
		}

		command := strings.Join(args, " ")
		fmt.Printf("Explaining command: '%s'\n", command)
		
		if appCtx.Config.Debug {
            apiKey := appCtx.Config.GeminiAPIKey
            if len(apiKey) > 4 {
                fmt.Printf("DEBUG: Using API key ending in ...%s\n", apiKey[len(apiKey)-4:])
            } else {
                fmt.Printf("DEBUG: Using API key (too short to truncate).\n")
            }
        }
		
		// TODO: Implement explanation logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}

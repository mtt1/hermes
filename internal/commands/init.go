// Package commands - init subcommand
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"hermes/internal/exit"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Generate shell integration script",
	Long: `Generate shell integration script for the specified shell.

Supported shells:
  - zsh
  - bash
  - fish

Examples:
  hermes init zsh     # Generate zsh integration script
  hermes init bash    # Generate bash integration script
  hermes init fish    # Generate fish integration script`,
	
	Args: cobra.ExactArgs(1), // Require exactly one argument (shell name)
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		
		// Validate shell name
		switch shell {
		case "zsh", "bash", "fish":
			fmt.Printf("Generating init script for %s...\n", shell)
			// TODO: Implement shell script generation
			return nil
		default:
			return exit.NewError(exit.CodeInvalid, "unsupported shell: %s (supported: zsh, bash, fish)", shell)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
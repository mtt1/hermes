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

This command outputs shell-specific integration code that you can evaluate
in your shell to enable Hermes functionality. The integration includes:
  - The hermes function that handles command generation and safety warnings
  - Proper exit code handling for different command types
  - Shell-specific buffer manipulation

Supported shells:
  - zsh
  - bash  
  - fish

Examples:
  hermes init zsh                              # Generate zsh integration script
  hermes init bash                             # Generate bash integration script
  hermes init fish                             # Generate fish integration script

Installation:
  Add to your shell config file:
  
  # For zsh (~/.zshrc):
  eval "$(hermes init zsh)"
  
  # For bash (~/.bashrc):
  eval "$(hermes init bash)"
  
  # For fish (~/.config/fish/config.fish):
  hermes init fish | source`,
	
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
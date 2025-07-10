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
  - zsh (currently available)
  - bash (coming soon)
  - fish (coming soon)

Examples:
  hermes init zsh                              # Generate zsh integration script

Installation:
  Add to your zsh config file (~/.zshrc):
  
  eval "$(hermes init zsh)"
  
  Then restart your shell or run: source ~/.zshrc`,
	
	Args: cobra.ExactArgs(1), // Require exactly one argument (shell name)
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		
		// Generate shell-specific integration script
		switch shell {
		case "zsh":
			fmt.Print(generateZshScript())
			return nil
		// TODO: Future shell support
		// case "bash":
		// 	fmt.Print(generateBashScript())
		// 	return nil
		// case "fish":
		// 	fmt.Print(generateFishScript())
		// 	return nil
		default:
			return exit.NewError(exit.CodeError, "unsupported shell: %s (currently only 'zsh' is supported)", shell)
		}
	},
}

// generateZshScript returns the zsh integration script
func generateZshScript() string {
	return `# Hermes zsh integration
# This function provides natural language command generation with safety warnings

hermes() {
    # If no arguments provided, show help
    if [[ $# -eq 0 ]]; then
        command hermes --help
        return
    fi
    
    # Check if this is a generation request (needs buffer placement)
    # Look for 'gen' or 'generate' subcommand in arguments
    local is_generation=false
    for arg in "$@"; do
        case "$arg" in
            gen|generate)
                is_generation=true
                break
                ;;
        esac
    done
    
    # If it's NOT a generation command, pass through directly
    if [[ "$is_generation" = false ]]; then
        HERMES_SHELL_INTEGRATION=1 command hermes "$@"
        return $?
    fi
    
    # Otherwise, it's a generation command - capture output for buffer
    local output exit_code
    
    # Capture both stdout and exit code
    # Set HERMES_SHELL_INTEGRATION=1 to indicate we're running from shell integration
    # Note: stderr goes directly to terminal for immediate feedback
    output=$(HERMES_SHELL_INTEGRATION=1 command hermes "$@")
    exit_code=$?
    
    case $exit_code in
        0)
            # Safe command - place directly in buffer
            print -z "$output"
            ;;
        10)
            # Requires attention - show warning above prompt
            echo ""
            echo "REQUIRES ATTENTION - Potentially destructive action ahead, review before execution"
            echo ""
            print -z "$output"
            ;;
        *)
            # Error condition - show error message
            HERMES_SHELL_INTEGRATION=1 command hermes "$@"
            return $exit_code
            ;;
    esac
}

# Optional: Set up alias for faster access
# Uncomment the line below if you want 'h' as a shortcut
# alias h='hermes'
`
}

// TODO: Future shell support functions
// func generateBashScript() string { ... }
// func generateFishScript() string { ... }

func init() {
	rootCmd.AddCommand(initCmd)
}

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
			return exit.NewError(exit.CodeInvalid, "unsupported shell: %s (currently only 'zsh' is supported)", shell)
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
    
    # If it's a subcommand (gen, exp, init), pass through directly
    case "$1" in
        gen|generate|exp|explain|init|--help|-h|--version|-v)
            command hermes "$@"
            return $?
            ;;
    esac
    
    # Otherwise, treat it as natural language for generation
    local output exit_code
    
    # Capture both stdout and exit code
    output=$(command hermes gen "$@" 2>/dev/null)
    exit_code=$?
    
    case $exit_code in
        0)
            # Safe command - place directly in buffer
            print -z "$output"
            ;;
        10)
            # Requires attention - show warning above prompt
            echo ""
            echo "⚠️  REQUIRES ATTENTION - Review before execution"
            echo "Generated command: $output"
            echo ""
            print -z "$output"
            ;;
        *)
            # Error condition - show error message
            command hermes gen "$@"
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
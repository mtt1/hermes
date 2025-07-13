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
  hermes init fish                             # Generate fish function

Installation:
  For zsh - Add to ~/.zshrc:
    eval "$(hermes init zsh)"
    
  For bash - Add to ~/.bashrc:
    eval "$(hermes init bash)"
    
  For fish - Save function to functions directory:
    mkdir -p ~/.config/fish/functions
    hermes init fish > ~/.config/fish/functions/hermes.fish
    
  Then restart your shell or reload config.`,
	
	Args: cobra.ExactArgs(1), // Require exactly one argument (shell name)
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		
		// Generate shell-specific integration script
		switch shell {
		case "zsh":
			fmt.Print(generateZshScript())
			return nil
		case "bash":
			fmt.Print(generateBashScript())
			return nil
		case "fish":
			fmt.Print(generateFishScript())
			return nil
		default:
			return exit.NewError(exit.CodeError, "unsupported shell: %s (supported: zsh, bash, fish)", shell)
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

// generateBashScript returns the bash integration script
func generateBashScript() string {
	return `# Hermes bash integration
# This function provides natural language command generation with safety warnings

hermes() {
    # If no arguments provided, show help
    if [ "$#" -eq 0 ]; then
        command hermes --help
        return
    fi
    
    # Check if this is a generation request (needs buffer placement)
    # Look for 'gen' or 'generate' subcommand in arguments
    local is_generation=0
    for arg in "$@"; do
        if [[ "$arg" == "gen" || "$arg" == "generate" ]]; then
            is_generation=1
            break
        fi
    done
    
    # If it's NOT a generation command, pass through directly
    if [ "$is_generation" -eq 0 ]; then
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
            read -e -i "$output"
            ;;
        10)
            # Requires attention - show warning above prompt
            echo ""
            echo "REQUIRES ATTENTION - Potentially destructive action ahead, review before execution"
            echo ""
            read -e -i "$output"
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

// generateFishScript returns the fish function (pure function, no installation comments)
func generateFishScript() string {
	return `function hermes
    # If no arguments provided, show help
    if test (count $argv) -eq 0
        command hermes --help
        return
    end
    
    # Check if this is a generation request (needs buffer placement)
    # Look for 'gen' or 'generate' subcommand in arguments
    set -l is_generation 0
    if contains -- "gen" $argv; or contains -- "generate" $argv
        set is_generation 1
    end
    
    # If it's NOT a generation command, pass through directly
    if test $is_generation -eq 0
        HERMES_SHELL_INTEGRATION=1 command hermes $argv
        return
    end
    
    # Otherwise, it's a generation command - capture output for buffer
    set -l output (HERMES_SHELL_INTEGRATION=1 command hermes $argv)
    set -l exit_code $status
    
    switch $exit_code
        case 0
            # Safe command - place directly in buffer
            commandline $output
        case 10
            # Requires attention - show warning above prompt
            echo ""
            echo "REQUIRES ATTENTION - Potentially destructive action ahead, review before execution"
            echo ""
            commandline $output
        case '*'
            # Error condition - show error message
            HERMES_SHELL_INTEGRATION=1 command hermes $argv
            return 1
    end
end
`
}

func init() {
	rootCmd.AddCommand(initCmd)
}

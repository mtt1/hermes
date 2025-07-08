// Package commands contains all CLI command definitions
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"hermes/internal/exit"
)

// Global flags
var (
	explainFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hermes [natural language query]",
	Short: "Hermes is a smart CLI assistant that translates natural language to shell commands",
	Long: `Hermes is a terminal AI helper that translates natural language to shell commands.
	
Examples:
  hermes list files                    # Generate command to list files
  hermes delete old logs               # Generate command to delete old logs
  hermes --explain ls -la              # Explain what a command does
  hermes init zsh                      # Generate zsh integration script`,
	
	// This runs when no subcommand is specified (default action)
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// If no args given, show help
			return cmd.Help()
		}

		if explainFlag {
			// Explain mode: join all args as the command to explain
			command := strings.Join(args, " ")
			fmt.Printf("Explaining command: '%s'\n", command)
			// TODO: Implement explanation logic
			return exit.Success()
		}

		// Default mode: generate command from natural language
		query := strings.Join(args, " ")
		fmt.Printf("Generating command for: '%s'\n", query)
		// TODO: Implement core command generation logic
		return nil
	},
}

// Execute is the main entry point for the CLI
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().BoolVar(&explainFlag, "explain", false, "Explain the given command instead of generating a new one")
	
	// Set version - can be injected at build time
	rootCmd.Version = "0.1.0"
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)
}
// Package commands contains all CLI command definitions
package commands

import (
	"github.com/spf13/cobra"
)

// No global flags needed with explicit subcommands

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hermes",
	Short: "Hermes is a smart CLI assistant that translates natural language to shell commands",
	Long: `Hermes is a terminal AI helper that translates natural language to shell commands.

Commands:
  hermes gen/generate [natural language]    # Generate shell commands from natural language
  hermes exp/explain [command]              # Explain what a shell command does
  hermes init [shell]                       # Generate shell integration script

Examples:
  hermes gen list all files                 # Generate command to list files
  hermes generate delete old logs           # Generate command to delete old logs
  hermes exp ls -la                         # Explain what 'ls -la' does
  hermes explain "find . -name '*.go'"      # Explain a complex command
  hermes init zsh                           # Generate zsh integration script

Quick Start:
  Add this alias to your shell config for faster access:
  alias h='hermes gen'
  
  Then you can use: h list all files`,
	
	// Show help when no subcommand is provided
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute is the main entry point for the CLI
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set version - can be injected at build time
	rootCmd.Version = "0.1.0"
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)
}

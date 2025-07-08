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

Examples:
  hermes exp ls -la                            # Explain 'ls -la'
  hermes explain "find . -name '*.go'"         # Explain a find command
  hermes exp grep -r "TODO" --include="*.py"   # Explain a complex grep
  hermes explain tar -czf archive.tar.gz dir/  # Explain a tar command

Note: You can use quotes around the command if it contains special characters
or you want to be explicit about the command boundaries.`,

	Args: cobra.MinimumNArgs(1), // Require at least one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		command := strings.Join(args, " ")
		fmt.Printf("Explaining command: '%s'\n", command)
		// TODO: Implement explanation logic
		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}
// Package commands contains all CLI command definitions
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/cobra"
	"hermes/internal/config"
)

// AppContext holds dependencies for the application
type AppContext struct {
	Config config.Config
}

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
  
  Then you can use: h list all files

Configuration:
  Set your Gemini API key via:
  - Environment variable: GEMINI_API_KEY
  - CLI flag: --gemini-api-key
  - Config file: ~/.config/hermes/config.toml`,
	
	// Load configuration before any command runs
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return loadConfig(cmd)
	},
	
	// Show help when no subcommand is provided
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Global app context
var appCtx *AppContext

// Execute is the main entry point for the CLI
func Execute() error {
	return rootCmd.Execute()
}

func loadConfig(cmd *cobra.Command) error {
	// Initialize app context
	appCtx = &AppContext{
		Config: config.Default(),
	}

	// 1. Load config file (lowest priority)
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		configPath := filepath.Join(userConfigDir, "hermes", "config.toml")
		if err := config.K.Load(file.Provider(configPath), toml.Parser()); err != nil {
			// It's okay if the file doesn't exist
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "warning: failed to load config file: %v\n", err)
			}
		}
	}

	// 2. Load environment variables (higher priority) 
	// Check for GEMINI_API_KEY and map it to gemini_api_key
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		config.K.Set("gemini_api_key", geminiKey)
	}

	// 3. Load CLI flags (highest priority)
	if err := config.K.Load(posflag.Provider(cmd.Flags(), ".", config.K), nil); err != nil {
		return fmt.Errorf("failed to load flags: %w", err)
	}
	
	// Map CLI flag to config key (--gemini-api-key -> gemini_api_key)
	if flagValue, _ := cmd.Flags().GetString("gemini-api-key"); flagValue != "" {
		config.K.Set("gemini_api_key", flagValue)
	}

	// 4. Unmarshal all configuration into the Config struct
	if err := config.K.Unmarshal("", &appCtx.Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func init() {
	// Set version - can be injected at build time
	rootCmd.Version = "0.1.0"
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Add global flags
	rootCmd.PersistentFlags().String("gemini-api-key", "", "Gemini API key for AI command generation and explanation")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
}

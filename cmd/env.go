package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/devup/internal/config"
)

var (
	envFormat string
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print environment variables",
	Long: `Print environment variables from the .env file.

This command reads the .env file and outputs the variables in a format
that can be sourced by your shell.

Examples:
  # Print environment variables
  devup env

  # Export to current shell (bash/zsh)
  eval $(devup env)

  # Source the .env file directly
  source .env`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEnv()
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.Flags().StringVar(&envFormat, "format", "export", "output format: export, json, or raw")
}

func runEnv() error {
	// Load configuration
	loader := config.NewLoaderWithLocal(cfgFile, local)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get app
	app, err := getApp(cfg)
	if err != nil {
		return err
	}

	// Determine .env file path
	envFile := ".env"
	if app.Setup.EnvFile != "" {
		envFile = app.Setup.EnvFile
	}
	envPath := filepath.Join(app.WorkDir, envFile)

	// Check if .env file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf(".env file not found at %s. Run 'devup setup' first", envPath)
	}

	// Read .env file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Parse and output
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if envFormat == "export" {
			// Split on first = only
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				// Remove existing quotes if present
				value = strings.Trim(value, "\"'")
				// Use %q to properly escape special characters for shell
				fmt.Printf("export %s=%q\n", key, value)
			}
		} else if envFormat == "raw" {
			fmt.Println(line)
		} else if envFormat == "json" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				fmt.Printf("  \"%s\": \"%s\",\n", parts[0], strings.Trim(parts[1], "\"'"))
			}
		}
	}

	return nil
}

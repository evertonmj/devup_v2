package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage the default project used by devup",
	Long: `Configure a default project directory so you can run devup commands
from anywhere without passing -c or setting environment variables.

Priority order (highest to lowest):
  1) -c / --config flag
  2) -l / --local flag (forces local directory)
  3) DEVUP_DEFAULT_PROJECT environment variable
  4) Saved default project (this setting)
  5) Standard search locations (./devup.yaml, etc.)`,
}

var projectSetCmd = &cobra.Command{
	Use:   "set <path-to-project>",
	Short: "Set the default project directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Expand ~ if present
		if len(input) > 0 && input[0] == '~' {
			if home, err := os.UserHomeDir(); err == nil {
				input = filepath.Join(home, input[1:])
			}
		}

		absPath, err := filepath.Abs(input)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Validate directory exists
		stat, err := os.Stat(absPath)
		if err != nil || !stat.IsDir() {
			return fmt.Errorf("path is not a directory or does not exist: %s", absPath)
		}

		// Optional: warn if devup.yaml not present
		cfg := filepath.Join(absPath, "devup.yaml")
		if _, err := os.Stat(cfg); err != nil {
			fmt.Printf("⚠️  Warning: devup.yaml not found in %s\n", absPath)
		}

		// Ensure config dir exists
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "devup")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}

		// Write default project file
		filePath := filepath.Join(configDir, "default_project")
		if err := os.WriteFile(filePath, []byte(absPath), 0644); err != nil {
			return fmt.Errorf("failed to save default project: %w", err)
		}

		fmt.Printf("✅ Default project set to: %s\n", absPath)
		fmt.Println("You can now run 'devup start' from anywhere.")
		return nil
	},
}

var projectShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current default project directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show environment variable if set
		if env := os.Getenv("DEVUP_DEFAULT_PROJECT"); env != "" {
			fmt.Printf("ENV DEVUP_DEFAULT_PROJECT: %s (takes precedence)\n", env)
		}

		configDir := filepath.Join(os.Getenv("HOME"), ".config", "devup")
		filePath := filepath.Join(configDir, "default_project")
		if data, err := os.ReadFile(filePath); err == nil {
			fmt.Printf("Saved default project: %s\n", string(data))
		} else {
			fmt.Println("No saved default project configured.")
		}
		return nil
	},
}

var projectClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the saved default project",
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "devup")
		filePath := filepath.Join(configDir, "default_project")
		if err := os.Remove(filePath); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No saved default project to clear.")
				return nil
			}
			return fmt.Errorf("failed to clear default project: %w", err)
		}
		fmt.Println("✅ Saved default project cleared.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectClearCmd)
}

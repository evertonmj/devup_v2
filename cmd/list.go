package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"devup/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available applications",
	Long: `Display all applications defined in the configuration file.

Examples:
  # List all apps
  devup list

  # List apps from specific config
  devup list -c /path/to/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listApplications()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listApplications() error {
	// Load configuration
	loader := config.NewLoaderWithLocal(cfgFile, local)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Available Applications\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if len(cfg.Apps) == 0 {
		fmt.Println("No applications found in configuration.")
		return nil
	}

	for name, app := range cfg.Apps {
		fmt.Printf("📦 %s\n", name)
		if app.Description != "" {
			fmt.Printf("   %s\n", app.Description)
		}
		fmt.Printf("   Services: %d\n", len(app.Services))
		fmt.Printf("   Modes:    %d\n", len(app.Modes))
		fmt.Printf("   WorkDir:  %s\n", app.WorkDir)
		fmt.Println()
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Total: %d application(s)\n\n", len(cfg.Apps))

	return nil
}

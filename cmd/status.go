package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/devup/internal/config"
	"github.com/yourusername/devup/internal/service"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of services",
	Long: `Display the current status of all services for the specified application.

Examples:
  # Check status of default app
  devup status

  # Check status of specific app
  devup status -a myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showStatus() error {
	// Load configuration
	loader := config.NewLoaderWithLocal(cfgFile, local)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which app to check
	app, err := getApp(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Status: %s\n", app.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Create service manager
	mgr := service.NewManager(app, "")

	// Get status
	displayServiceStatus(mgr)

	return nil
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"devup/internal/config"
	"devup/internal/log"
	"devup/internal/service"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the application services",
	Long: `Stop all running services for the specified application.

Examples:
  # Stop default app
  devup stop

  # Stop specific app
  devup stop -a myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopApplication()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stopApplication() error {
	// Load configuration
	loader := config.NewLoaderWithLocal(cfgFile, local)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which app to stop
	app, err := getApp(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Stopping: %s\n", app.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Create service manager with empty mode (not needed for stop)
	mgr := service.NewManager(app, "")

	// Stop services
	ctx := context.Background()
	if err := mgr.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	// Clean ports after stopping to ensure everything is killed
	fmt.Println("🧹 Cleaning ports...")
	ports := collectPorts(app)
	if len(ports) > 0 {
		for _, port := range ports {
			if err := service.CleanPort(port); err != nil {
				log.Errorf("failed to clean port %d: %v", port, err)
			} else {
				fmt.Printf("   ✓ Cleaned port %d\n", port)
			}
		}
		fmt.Println()
	}

	fmt.Printf("✅ All services stopped successfully!\n\n")

	return nil
}


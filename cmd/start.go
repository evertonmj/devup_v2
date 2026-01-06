package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"devup/internal/config"
	"devup/internal/service"
)

var (
	mode       string
	followLogs bool
	showLogs   bool
	quietMode  bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application services",
	Long: `Start all services for the specified application in the given mode.

Examples:
  # Start default app in default mode
  devup start

  # Start specific app in mock mode
  devup start -a myapp --mode mock

  # Start in Python mode with debug
  devup start --mode py --debug`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startApplication()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&mode, "mode", "m", "default", "running mode (e.g., mock, py, debug)")
	startCmd.Flags().BoolVar(&followLogs, "follow", true, "follow and display service logs continuously (default: true)")
	startCmd.Flags().BoolVar(&showLogs, "logs", false, "display startup logs for 5 seconds only")
	startCmd.Flags().BoolVarP(&quietMode, "quiet", "q", false, "quiet mode - suppress log output")
}

func startApplication() error {
	// Load configuration
	loader := config.NewLoaderWithLocal(cfgFile, local)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which app to start
	app, err := getApp(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Starting: %s\n", app.Name)
	fmt.Printf("Mode:     %s\n", mode)
	fmt.Printf("WorkDir:  %s\n", app.WorkDir)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Clean ports before starting
	fmt.Println("🧹 Cleaning ports...")
	ports := collectPorts(app)
	if len(ports) > 0 {
		for _, port := range ports {
			if err := service.CleanPort(port); err != nil {
				fmt.Printf("   ⚠️  Warning: failed to clean port %d: %v\n", port, err)
			} else {
				fmt.Printf("   ✓ Cleaned port %d\n", port)
			}
		}
		fmt.Println()
	}

	// Create service manager
	mgr := service.NewManager(app, mode)

	// Start services
	ctx := context.Background()
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	fmt.Printf("\n✅ All services started successfully!\n\n")

	// Display service status
	displayServiceStatus(mgr)

	// Show logs by default unless quiet mode is enabled
	if !quietMode && (showLogs || followLogs) {
		fmt.Printf("\n📋 Service Logs:\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		if showLogs && !followLogs {
			// Show logs for 5 seconds only
			tailLogs(mgr, 5*time.Second)
			fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		} else {
			// Follow logs continuously until interrupted (default behavior)
			fmt.Println("Following logs... Press Ctrl+C to stop")

			// Setup signal handler
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Create a done channel
			done := make(chan bool)

			// Start tailing logs in background
			go func() {
				tailLogsForever(mgr)
				done <- true
			}()

			// Wait for interrupt
			select {
			case <-sigChan:
				fmt.Println("\n\nStopping log tail...")
			case <-done:
				// Logs ended naturally (shouldn't happen)
			}

			fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		}
	}

	return nil
}

func displayServiceStatus(mgr *service.Manager) {
	status := mgr.Status()

	fmt.Printf("Service Status:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for name, svcStatus := range status {
		statusIcon := "❌"
		if svcStatus.Running {
			statusIcon = "✅"
		}

		fmt.Printf("%s %s\n", statusIcon, name)
		if svcStatus.Running {
			fmt.Printf("   PID:  %d\n", svcStatus.PID)
			if svcStatus.Port > 0 {
				fmt.Printf("   Port: %d\n", svcStatus.Port)
			}
			fmt.Printf("   Started: %s\n", svcStatus.StartTime.Format("15:04:05"))
		}
		fmt.Println()
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func getApp(cfg *config.AppConfig) (*config.AppSpec, error) {
	// If app name specified, use it
	if appName != "" {
		app, exists := cfg.Apps[appName]
		if !exists {
			return nil, fmt.Errorf("app '%s' not found in configuration", appName)
		}
		return &app, nil
	}

	// If only one app, use it
	if len(cfg.Apps) == 1 {
		for _, app := range cfg.Apps {
			return &app, nil
		}
	}

	// Multiple apps, need to specify
	fmt.Fprintf(os.Stderr, "Multiple apps found. Please specify with -a flag:\n\n")
	for name, app := range cfg.Apps {
		fmt.Fprintf(os.Stderr, "  • %s - %s\n", name, app.Description)
	}
	fmt.Fprintln(os.Stderr)

	return nil, fmt.Errorf("please specify an app with -a/--app flag")
}

// tailLogs displays logs from all services for a specified duration
func tailLogs(mgr *service.Manager, duration time.Duration) {
	status := mgr.Status()
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		for name, svcStatus := range status {
			if !svcStatus.Running {
				continue
			}

			// Get logs for this service
			logs, err := mgr.GetServiceLogs(name)
			if err != nil {
				continue
			}

			// Display new logs
			for _, line := range logs {
				fmt.Printf("[%s] %s\n", name, line)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// tailLogsForever continuously displays logs from all services
func tailLogsForever(mgr *service.Manager) {
	status := mgr.Status()

	// Track last read position for each service
	lastLines := make(map[string]int)

	for {
		for name, svcStatus := range status {
			if !svcStatus.Running {
				continue
			}

			// Get logs for this service
			logs, err := mgr.GetServiceLogs(name)
			if err != nil {
				continue
			}

			// Display only new logs
			lastRead := lastLines[name]
			if len(logs) > lastRead {
				newLogs := logs[lastRead:]
				for _, line := range newLogs {
					fmt.Printf("[%s] %s\n", name, line)
				}
				lastLines[name] = len(logs)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// collectPorts collects all ports used by services in the app
func collectPorts(app *config.AppSpec) []int {
	ports := []int{}
	seen := make(map[int]bool)

	for _, service := range app.Services {
		if service.Port > 0 && !seen[service.Port] {
			ports = append(ports, service.Port)
			seen[service.Port] = true
		}
	}

	return ports
}

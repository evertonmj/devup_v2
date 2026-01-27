package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"devup/internal/config"

	"github.com/spf13/cobra"
)

var (
	cleanDryRun bool
	cleanEnv    bool
	cleanLogs   bool
	cleanDirs   bool
	cleanAll    bool
	cleanForce  bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove generated files and directories",
	Long: `Clean removes generated files and directories created by devup setup.

This includes:
  • Environment files (.env)
  • Log directories and files
  • Setup-created directories
  • Temporary files

Use --dry-run to see what would be removed without actually deleting anything.`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "show what would be removed without deleting")
	cleanCmd.Flags().BoolVar(&cleanEnv, "env", false, "remove .env file only")
	cleanCmd.Flags().BoolVar(&cleanLogs, "logs", false, "remove log directories only")
	cleanCmd.Flags().BoolVar(&cleanDirs, "dirs", false, "remove setup directories only")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "remove everything (env, logs, dirs)")
	cleanCmd.Flags().BoolVar(&cleanForce, "force", false, "skip confirmation prompt")
}

func runClean(cmd *cobra.Command, args []string) error {
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

	// Default behavior: clean logs and setup directories, but DO NOT remove .env
	// Users must explicitly pass --env or --all to remove the .env file.
	if !cleanEnv && !cleanLogs && !cleanDirs && !cleanAll {
		cleanLogs = true
		cleanDirs = true
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Cleaning: %s\n", app.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if cleanDryRun {
		fmt.Println("\n🔍 DRY-RUN MODE: No files will be deleted")
	}

	// Confirm before cleaning (unless --force or --dry-run)
	if !cleanDryRun && !cleanForce {
		fmt.Println("\n⚠️  WARNING: This will permanently delete files and directories.")
		fmt.Print("Are you sure you want to continue? (yes/no): ")

		var response string
		_, _ = fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("\n❌ Clean cancelled")
			return nil
		}
		fmt.Println()
	}

	// Change to app workdir
	if app.WorkDir != "" {
		if err := os.Chdir(app.WorkDir); err != nil {
			return fmt.Errorf("failed to change to workdir: %w", err)
		}
	}

	itemsRemoved := 0
	itemsFailed := 0

	// Clean .env file
	if cleanAll || cleanEnv {
		envFile := ".env"
		if app.Setup.EnvFile != "" {
			envFile = app.Setup.EnvFile
		}

		removed, failed := cleanPath(envFile, "Environment file", cleanDryRun)
		itemsRemoved += removed
		itemsFailed += failed
	}

	// Clean log directories
	if cleanAll || cleanLogs {
		logDirs := []string{"logs"}

		// Also check for log files in services
		for _, service := range app.Services {
			if service.LogFile != "" {
				logDir := filepath.Dir(service.LogFile)
				if logDir != "." && logDir != "" {
					logDirs = appendUnique(logDirs, logDir)
				}
			}
		}

		for _, dir := range logDirs {
			removed, failed := cleanPath(dir, "Log directory", cleanDryRun)
			itemsRemoved += removed
			itemsFailed += failed
		}
	}

	// Clean setup directories
	if cleanAll || cleanDirs {
		for _, dir := range app.Setup.Directories {
			removed, failed := cleanPath(dir, "Setup directory", cleanDryRun)
			itemsRemoved += removed
			itemsFailed += failed
		}
	}

	// Clean setup files
	if cleanAll {
		for _, file := range app.Setup.Files {
			removed, failed := cleanPath(file.Path, "Setup file", cleanDryRun)
			itemsRemoved += removed
			itemsFailed += failed
		}
	}

	// Print summary
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if cleanDryRun {
		fmt.Printf("Would remove %d item(s)\n", itemsRemoved)
		if itemsFailed > 0 {
			fmt.Printf("⚠️  %d item(s) not found\n", itemsFailed)
		}
	} else {
		fmt.Printf("✅ Removed %d item(s)\n", itemsRemoved)
		if itemsFailed > 0 {
			fmt.Printf("⚠️  %d item(s) failed or not found\n", itemsFailed)
		}
	}

	return nil
}

// cleanPath removes a file or directory
func cleanPath(path string, description string, dryRun bool) (removed int, failed int) {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("  ⏭️  %s: %s (not found)\n", description, path)
		return 0, 1
	}
	if err != nil {
		fmt.Printf("  ❌ %s: %s (error: %v)\n", description, path, err)
		return 0, 1
	}

	if dryRun {
		if info.IsDir() {
			fmt.Printf("  Would remove directory: %s\n", path)
		} else {
			fmt.Printf("  Would remove file: %s\n", path)
		}
		return 1, 0
	}

	// Remove the path
	if err := os.RemoveAll(path); err != nil {
		fmt.Printf("  ❌ Failed to remove %s: %s (%v)\n", description, path, err)
		return 0, 1
	}

	if info.IsDir() {
		fmt.Printf("  ✅ Removed directory: %s\n", path)
	} else {
		fmt.Printf("  ✅ Removed file: %s\n", path)
	}
	return 1, 0
}

// appendUnique adds a string to a slice if it's not already present
func appendUnique(slice []string, item string) []string {
	for _, existing := range slice {
		if existing == item {
			return slice
		}
	}
	return append(slice, item)
}

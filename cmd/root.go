package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	appName string
	verbose bool
	local   bool
)

// getVersion reads the version from the VERSION file
func getVersion() string {
	// Try to find VERSION file relative to the executable
	exePath, err := os.Executable()
	if err != nil {
		return "dev"
	}

	exeDir := filepath.Dir(exePath)

	// Try different possible locations for VERSION file
	possiblePaths := []string{
		filepath.Join(exeDir, "..", "VERSION"),           // build/devup -> VERSION
		filepath.Join(exeDir, "VERSION"),                 // Same directory
		filepath.Join(exeDir, "..", "..", "VERSION"),     // Deeper nesting
		"VERSION",                                        // Current directory
	}

	for _, path := range possiblePaths {
		if data, err := os.ReadFile(path); err == nil {
			version := strings.TrimSpace(string(data))
			if version != "" {
				return version
			}
		}
	}

	return "dev"
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "devup",
	Short: "DevUp - Extensible development environment manager",
	Long: `DevUp is an extensible CLI tool for managing development environments.

It allows you to define multiple applications with their services, modes,
and configurations in a single YAML file, making it easy to manage complex
development setups across different projects.

Features:
  • Multi-app support with isolated configurations
  • Multiple service types (processes, Docker, tmux)
  • Different running modes (mock, development, production)
  • Health checks and dependency management
  • Lifecycle hooks for custom automation
  • Extensible plugin-like architecture`,
	Version: getVersion(),
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: devup.yaml)")
	rootCmd.PersistentFlags().StringVarP(&appName, "app", "a", "", "application name to manage")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "use local directory config, ignore DEVUP_DEFAULT_PROJECT environment variable")
}

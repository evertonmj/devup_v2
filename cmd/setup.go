package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"devup/internal/config"
)

var (
	setupDryRun   bool
	skipPrompts   bool
	useDefaults   bool
	envOutputFile string
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up the application environment",
	Long: `Set up the application environment including directories, files, and environment variables.

This command will:
  • Create required directories
  • Generate or copy configuration files
  • Generate environment variables (.env file)
  • Run setup scripts
  • Validate the setup

Examples:
  # Setup default app
  devup setup

  # Setup specific app
  devup setup -a myapp

  # Dry run (show what would be done)
  devup setup --dry-run

  # Use defaults without prompting
  devup setup --use-defaults

  # Skip all prompts
  devup setup --skip-prompts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false, "show what would be done without making changes")
	setupCmd.Flags().BoolVar(&skipPrompts, "skip-prompts", false, "skip all prompts (use existing values or defaults)")
	setupCmd.Flags().BoolVar(&useDefaults, "use-defaults", false, "use default values without prompting")
	setupCmd.Flags().StringVar(&envOutputFile, "env-output", ".env", "output file for environment variables")
}

func runSetup() error {
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

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Setting Up Environment: %s\n", app.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if setupDryRun {
		fmt.Println("🔍 DRY-RUN MODE: No actual changes will be made")
	}

	// Execute pre-setup hooks
	if len(app.Hooks.PreSetup) > 0 {
		fmt.Println("📋 Executing pre-setup hooks...")
		if err := executeHooks(app.Hooks.PreSetup, app.WorkDir, setupDryRun); err != nil {
			return fmt.Errorf("pre-setup hooks failed: %w", err)
		}
		fmt.Println()
	}

	// Create directories
	if len(app.Setup.Directories) > 0 {
		fmt.Println("📁 Creating directories...")
		if err := createDirectories(app, setupDryRun); err != nil {
			return fmt.Errorf("directory creation failed: %w", err)
		}
		fmt.Println()
	}

	// Create/copy files
	if len(app.Setup.Files) > 0 {
		fmt.Println("📄 Creating files...")
		if err := createFiles(app, setupDryRun); err != nil {
			return fmt.Errorf("file creation failed: %w", err)
		}
		fmt.Println()
	}

	// Generate environment variables
	if len(app.Setup.EnvVars) > 0 {
		fmt.Println("🔐 Generating environment variables...")
		if err := generateEnvFile(app, setupDryRun); err != nil {
			return fmt.Errorf("environment generation failed: %w", err)
		}
		fmt.Println()
	}

	// Run setup scripts
	if len(app.Setup.Scripts) > 0 {
		fmt.Println("⚙️  Running setup scripts...")
		if err := runSetupScripts(app, setupDryRun); err != nil {
			return fmt.Errorf("setup scripts failed: %w", err)
		}
		fmt.Println()
	}

	// Execute post-setup hooks
	if len(app.Hooks.PostSetup) > 0 {
		fmt.Println("📋 Executing post-setup hooks...")
		if err := executeHooks(app.Hooks.PostSetup, app.WorkDir, setupDryRun); err != nil {
			return fmt.Errorf("post-setup hooks failed: %w", err)
		}
		fmt.Println()
	}

	// Run validation checks
	if len(app.Setup.Checks) > 0 && !setupDryRun {
		fmt.Println("✓ Running validation checks...")
		if err := runSetupChecks(app); err != nil {
			return fmt.Errorf("validation checks failed: %w", err)
		}
		fmt.Println()
	}

	fmt.Printf("✅ Setup completed successfully!\n\n")

	// Display environment variables that were set
	if len(app.Setup.EnvVars) > 0 {
		fmt.Printf("🔐 Environment Variables Set:\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		envFilePath := filepath.Join(app.WorkDir, envOutputFile)
		if envVars, err := loadEnvFile(envFilePath); err == nil {
			for key := range envVars {
				fmt.Printf("  ✓ %s\n", key)
			}
		}
		fmt.Printf("\n💡 To export these to your current shell, run:\n")
		if cfgFile != "" {
			fmt.Printf("   eval $(devup env -c %s)\n\n", cfgFile)
		} else {
			fmt.Printf("   eval $(devup env)\n\n")
		}
	}

	fmt.Printf("Next steps:\n")
	fmt.Printf("  • To start the application:\n")
	fmt.Printf("    devup start\n\n")
	return nil
}

func createDirectories(app *config.AppSpec, isDryRun bool) error {
	for _, dir := range app.Setup.Directories {
		fullPath := filepath.Join(app.WorkDir, dir)

		if isDryRun {
			fmt.Printf("  Would create: %s\n", fullPath)
		} else {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if err := os.MkdirAll(fullPath, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
				}
				fmt.Printf("  ✅ Created: %s\n", fullPath)
			} else {
				fmt.Printf("  ⏭️  Already exists: %s\n", fullPath)
			}
		}
	}
	return nil
}

func createFiles(app *config.AppSpec, isDryRun bool) error {
	for _, file := range app.Setup.Files {
		fullPath := filepath.Join(app.WorkDir, file.Path)

		if isDryRun {
			fmt.Printf("  Would create: %s\n", fullPath)
			continue
		}

		// Check if file already exists
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ⏭️  Already exists: %s\n", fullPath)
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", fullPath, err)
		}

		var content string
		if file.Source != "" {
			// Copy from source
			sourceData, err := os.ReadFile(filepath.Join(app.WorkDir, file.Source))
			if err != nil {
				return fmt.Errorf("failed to read source file %s: %w", file.Source, err)
			}
			content = string(sourceData)
		} else if file.Content != "" {
			content = file.Content
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("  ✅ Created: %s\n", fullPath)
	}
	return nil
}

func generateEnvFile(app *config.AppSpec, isDryRun bool) error {
	envVars := make(map[string]string)

	// Load existing .env file if it exists
	envFilePath := filepath.Join(app.WorkDir, envOutputFile)
	if _, err := os.Stat(envFilePath); err == nil {
		existingVars, _ := loadEnvFile(envFilePath)
		for k, v := range existingVars {
			envVars[k] = v
		}
	}

	// Process each environment variable
	for _, envVar := range app.Setup.EnvVars {
		var value string

		// Check if already set
		if existingVal, exists := envVars[envVar.Name]; exists && existingVal != "" {
			fmt.Printf("  ⏭️  %s (already set)\n", envVar.Name)
			continue
		}

		// Generate value if specified
		if envVar.Generate != "" {
			if isDryRun {
				fmt.Printf("  Would generate: %s\n", envVar.Name)
				continue
			}

			cmd := exec.Command("bash", "-c", envVar.Generate)
			output, err := cmd.Output()
			if err != nil {
				if envVar.Required && !skipPrompts {
					return fmt.Errorf("failed to generate required variable %s: %w", envVar.Name, err)
				}
				fmt.Printf("  ⚠️  Failed to generate %s\n", envVar.Name)
				continue
			}
			value = strings.TrimSpace(string(output))
		} else if envVar.Prompt && !skipPrompts && !useDefaults {
			// Prompt user for value
			fmt.Printf("  🔑 %s", envVar.Name)
			if envVar.Description != "" {
				fmt.Printf(" (%s)", envVar.Description)
			}
			if envVar.Default != "" {
				fmt.Printf(" [default: %s]", envVar.Default)
			}
			value = promptUser("")
			if value == "" && envVar.Default != "" {
				value = envVar.Default
			}
		} else {
			// Use default value
			value = envVar.Default
		}

		// Generate random value if still empty and required
		if value == "" && envVar.Required && !skipPrompts {
			if envVar.Name == "JWT_SECRET_KEY" || envVar.Name == "SECRET_KEY" || strings.Contains(strings.ToLower(envVar.Name), "secret") {
				value = generateRandomSecret(32)
			}
		}

		if value != "" {
			envVars[envVar.Name] = value
			if !isDryRun {
				fmt.Printf("  ✅ Set: %s\n", envVar.Name)
			}
		} else if envVar.Required {
			return fmt.Errorf("required environment variable %s is not set", envVar.Name)
		}
	}

	// Write .env file
	if !isDryRun {
		if err := writeEnvFile(envFilePath, envVars); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
		fmt.Printf("\n  📝 Environment variables written to: %s\n", envFilePath)
		fmt.Printf("\n  💡 To export these variables to your current shell, run:\n")
		fmt.Printf("     source %s\n", envFilePath)
		fmt.Printf("     or: export $(cat %s | xargs)\n", envFilePath)
	} else {
		fmt.Printf("\n  Would write %d variables to: %s\n", len(envVars), envFilePath)
	}

	return nil
}

func runSetupScripts(app *config.AppSpec, isDryRun bool) error {
	for i, script := range app.Setup.Scripts {
		if isDryRun {
			fmt.Printf("  %d. Would run: %s\n", i+1, script)
		} else {
			fmt.Printf("  %d. Running: %s\n", i+1, script)
			cmd := exec.Command("bash", "-c", script)
			cmd.Dir = app.WorkDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("script failed: %w", err)
			}
			fmt.Printf("     ✅ Completed\n")
		}
	}
	return nil
}

func runSetupChecks(app *config.AppSpec) error {
	allPassed := true

	for _, check := range app.Setup.Checks {
		cmd := exec.Command("bash", "-c", check.Command)
		cmd.Dir = app.WorkDir

		if err := cmd.Run(); err != nil {
			allPassed = false
			fmt.Printf("  ❌ %s\n", check.Name)
			if check.Message != "" {
				fmt.Printf("     %s\n", check.Message)
			}
		} else {
			fmt.Printf("  ✅ %s\n", check.Name)
		}
	}

	if !allPassed {
		return fmt.Errorf("some validation checks failed")
	}

	return nil
}

func loadEnvFile(path string) (map[string]string, error) {
	vars := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return vars, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			vars[key] = value
		}
	}

	return vars, scanner.Err()
}

func writeEnvFile(path string, vars map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	fmt.Fprintln(writer, "# Environment variables generated by devup")
	fmt.Fprintln(writer, "# DO NOT commit this file to version control")
	fmt.Fprintln(writer, "")

	// Write variables
	for key, value := range vars {
		fmt.Fprintf(writer, "%s=%s\n", key, value)
	}

	return writer.Flush()
}

func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to simpler generation
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", os.Getpid())))
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

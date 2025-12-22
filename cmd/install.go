package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/devup/internal/config"
)

var (
	skipOptional bool
	dryRun       bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies for the application",
	Long: `Install all required dependencies for the application.

This command will:
  • Check for required dependencies
  • Install missing dependencies using appropriate package managers
  • Run custom installation steps
  • Execute pre/post install hooks

Examples:
  # Install all dependencies for default app
  devup install

  # Install for specific app
  devup install -a myapp

  # Skip optional dependencies
  devup install --skip-optional

  # Dry run (show what would be installed)
  devup install --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&skipOptional, "skip-optional", false, "skip optional dependencies")
	installCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be installed without installing")
}

func runInstall() error {
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
	fmt.Printf("Installing Dependencies: %s\n", app.Name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if dryRun {
		fmt.Println("🔍 DRY-RUN MODE: No actual changes will be made\n")
	}

	// Execute pre-install hooks
	if len(app.Hooks.PreInstall) > 0 {
		fmt.Println("📋 Executing pre-install hooks...")
		if err := executeHooks(app.Hooks.PreInstall, app.WorkDir, dryRun); err != nil {
			return fmt.Errorf("pre-install hooks failed: %w", err)
		}
		fmt.Println()
	}

	// Check and install dependencies
	if len(app.Install.Dependencies) > 0 {
		fmt.Println("📦 Checking dependencies...")
		if err := installDependencies(app, dryRun); err != nil {
			return fmt.Errorf("dependency installation failed: %w", err)
		}
		fmt.Println()
	}

	// Run installation steps
	if len(app.Install.Steps) > 0 {
		fmt.Println("⚙️  Running installation steps...")
		if err := runInstallSteps(app, dryRun); err != nil {
			return fmt.Errorf("installation steps failed: %w", err)
		}
		fmt.Println()
	}

	// Execute post-install hooks
	if len(app.Hooks.PostInstall) > 0 {
		fmt.Println("📋 Executing post-install hooks...")
		if err := executeHooks(app.Hooks.PostInstall, app.WorkDir, dryRun); err != nil {
			return fmt.Errorf("post-install hooks failed: %w", err)
		}
		fmt.Println()
	}

	fmt.Printf("✅ Installation completed successfully!\n\n")
	return nil
}

func installDependencies(app *config.AppSpec, isDryRun bool) error {
	required := 0
	installed := 0
	missing := 0
	optional := 0

	for _, dep := range app.Install.Dependencies {
		if dep.Optional && skipOptional {
			continue
		}

		// Check if already installed
		isInstalled := checkDependency(dep)

		statusIcon := "❌"
		statusText := "missing"

		if isInstalled {
			statusIcon = "✅"
			statusText = "installed"
			installed++
		} else {
			missing++
			if !dep.Optional {
				required++
			} else {
				optional++
			}
		}

		optionalTag := ""
		if dep.Optional {
			optionalTag = " (optional)"
		}

		fmt.Printf("  %s %s%s - %s\n", statusIcon, dep.Name, optionalTag, statusText)

		// Install if missing
		if !isInstalled {
			if isDryRun {
				fmt.Printf("     Would install: %s\n", dep.Package)
			} else {
				fmt.Printf("     Installing %s...\n", dep.Name)
				if err := installDependency(dep); err != nil {
					if !dep.Optional {
						return fmt.Errorf("failed to install %s: %w", dep.Name, err)
					}
					fmt.Printf("     ⚠️  Failed to install optional dependency %s: %v\n", dep.Name, err)
				} else {
					fmt.Printf("     ✅ Installed %s\n", dep.Name)
				}
			}
		}
	}

	fmt.Printf("\n  Summary: %d installed, %d missing (%d required, %d optional)\n", installed, missing, required, optional)

	if required > 0 && !isDryRun {
		return fmt.Errorf("failed to install %d required dependencies", required)
	}

	return nil
}

func checkDependency(dep config.Dependency) bool {
	// Use custom check command if provided
	if dep.Check != "" {
		cmd := exec.Command("bash", "-c", dep.Check)
		return cmd.Run() == nil
	}

	// Default checks based on type
	switch dep.Type {
	case "brew":
		cmd := exec.Command("brew", "list", dep.Package)
		return cmd.Run() == nil
	case "npm":
		cmd := exec.Command("npm", "list", "-g", dep.Package)
		return cmd.Run() == nil
	case "pip":
		cmd := exec.Command("pip3", "show", dep.Package)
		return cmd.Run() == nil
	case "go":
		cmd := exec.Command("go", "version")
		return cmd.Run() == nil
	default:
		// For custom type, try to find the package name as a command
		cmd := exec.Command("which", dep.Package)
		return cmd.Run() == nil
	}
}

func installDependency(dep config.Dependency) error {
	// Use custom install command if provided
	if dep.InstallCmd != "" {
		cmd := exec.Command("bash", "-c", dep.InstallCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Default installation based on type
	var cmd *exec.Cmd
	switch dep.Type {
	case "brew":
		cmd = exec.Command("brew", "install", dep.Package)
	case "npm":
		cmd = exec.Command("npm", "install", "-g", dep.Package)
	case "pip":
		if dep.Version != "" {
			cmd = exec.Command("pip3", "install", fmt.Sprintf("%s==%s", dep.Package, dep.Version))
		} else {
			cmd = exec.Command("pip3", "install", dep.Package)
		}
	case "apt":
		cmd = exec.Command("sudo", "apt-get", "install", "-y", dep.Package)
	default:
		return fmt.Errorf("unsupported dependency type: %s", dep.Type)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runInstallSteps(app *config.AppSpec, isDryRun bool) error {
	for i, step := range app.Install.Steps {
		// Check if step should be skipped
		if step.SkipIf != "" {
			cmd := exec.Command("bash", "-c", step.SkipIf)
			if cmd.Run() == nil {
				fmt.Printf("  %d. %s - ⏭️  skipped\n", i+1, step.Name)
				continue
			}
		}

		fmt.Printf("  %d. %s\n", i+1, step.Name)
		if step.Description != "" {
			fmt.Printf("     %s\n", step.Description)
		}

		workDir := app.WorkDir
		if step.WorkDir != "" {
			workDir = step.WorkDir
		}

		for _, cmdStr := range step.Commands {
			if isDryRun {
				fmt.Printf("     Would run: %s\n", cmdStr)
			} else {
				fmt.Printf("     Running: %s\n", cmdStr)
				cmd := exec.Command("bash", "-c", cmdStr)
				cmd.Dir = workDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					return fmt.Errorf("step '%s' failed: %w", step.Name, err)
				}
			}
		}

		if !isDryRun {
			fmt.Printf("     ✅ Completed\n")
		}
		fmt.Println()
	}

	return nil
}

func executeHooks(hooks []string, workDir string, isDryRun bool) error {
	for _, hook := range hooks {
		if isDryRun {
			fmt.Printf("  Would run: %s\n", hook)
		} else {
			fmt.Printf("  Running: %s\n", hook)
			cmd := exec.Command("bash", "-c", hook)
			cmd.Dir = workDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("hook failed: %w", err)
			}
		}
	}
	return nil
}

func promptUser(prompt string) string {
	fmt.Printf("%s: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(response)
}

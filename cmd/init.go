package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initInteractive bool
	initScan        bool
	initForce       bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new DevUp configuration",
	Long: `Initialize a new DevUp configuration in the current directory.

This command will:
- Scan the project directory for common patterns
- Read README.md and other documentation to understand the project
- Detect package managers and build tools
- Create a devup.yaml configuration file with sensible defaults

Examples:
  devup init                    # Interactive mode with project scanning
  devup init --scan             # Auto-detect and create config
  devup init --force            # Overwrite existing config`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&initInteractive, "interactive", "i", true, "Interactive mode with prompts")
	initCmd.Flags().BoolVar(&initScan, "scan", true, "Scan project directory for patterns")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing devup.yaml")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if devup.yaml already exists
	configPath := "devup.yaml"
	if _, err := os.Stat(configPath); err == nil && !initForce {
		return fmt.Errorf("devup.yaml already exists. Use --force to overwrite")
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🚀 DevUp Project Initialization")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	var projectInfo *ProjectInfo
	var err error

	if initScan {
		fmt.Println("🔍 Scanning project directory...")
		projectInfo, err = scanProject(".")
		if err != nil {
			return fmt.Errorf("failed to scan project: %w", err)
		}
		fmt.Printf("✅ Detected: %s\n", projectInfo.Description)
		fmt.Println()
	} else {
		projectInfo = &ProjectInfo{
			Name: filepath.Base(getCurrentDir()),
		}
	}

	if initInteractive {
		if err := promptProjectInfo(projectInfo); err != nil {
			return fmt.Errorf("failed to get project info: %w", err)
		}
	}

	// Generate devup.yaml
	config := generateConfig(projectInfo)

	// Write to file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write devup.yaml: %w", err)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Successfully created devup.yaml")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Println("  1. Review and customize devup.yaml")
	fmt.Println("  2. Test your configuration: devup list")
	fmt.Println("  3. Start your application: devup start")
	fmt.Println()
	fmt.Println("📚 For help, see: https://github.com/evertonmj/devup_v2/blob/main/docs/TUTORIAL.md")
	fmt.Println()

	return nil
}

// ProjectInfo holds detected project information
type ProjectInfo struct {
	Name        string
	Description string
	Type        string // web, api, cli, library, etc.
	RootDir     string
	Services    []ServiceInfo
	Environment map[string]string
	PackageMgr  string // npm, go, pip, cargo, etc.
	BuildCmd    string
	TestCmd     string
	DevCmd      string
	InstallCmds []CommandInfo
	SetupCmds   []CommandInfo
}

// ServiceInfo holds information about a detected service
type ServiceInfo struct {
	Name         string
	Type         string
	Command      string
	Port         int
	Directory    string
	Dependencies []string
}

// CommandInfo holds a command and its working directory
type CommandInfo struct {
	Command string
	WorkDir string
}

func scanProject(dir string) (*ProjectInfo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	info := &ProjectInfo{
		Name:        filepath.Base(absDir),
		RootDir:     absDir,
		Services:    []ServiceInfo{},
		Environment: make(map[string]string),
	}

	// Detect package manager and project type
	detectPackageManager(absDir, info)

	// Read and analyze documentation
	analyzeDocumentation(absDir, info)

	// Detect common project structures
	if len(info.Services) == 0 {
		detectProjectStructure(absDir, info)
	}

	// Detect ports and services
	if len(info.Services) == 0 {
		detectServices(absDir, info)
	}

	return info, nil
}

func detectPackageManager(dir string, info *ProjectInfo) {
	checks := []struct {
		file     string
		pkgMgr   string
		projType string
		devCmd   string
		buildCmd string
		testCmd  string
	}{
		{"package.json", "npm", "nodejs", "npm run dev", "npm run build", "npm test"},
		{"go.mod", "go", "golang", "go run .", "go build", "go test ./..."},
		{"requirements.txt", "pip", "python", "python main.py", "", "pytest"},
		{"Pipfile", "pipenv", "python", "pipenv run python main.py", "", "pipenv run pytest"},
		{"Cargo.toml", "cargo", "rust", "cargo run", "cargo build", "cargo test"},
		{"pom.xml", "maven", "java", "mvn spring-boot:run", "mvn package", "mvn test"},
		{"build.gradle", "gradle", "java", "./gradlew bootRun", "./gradlew build", "./gradlew test"},
		{"Gemfile", "bundle", "ruby", "bundle exec rails s", "", "bundle exec rspec"},
		{"composer.json", "composer", "php", "php artisan serve", "", "phpunit"},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(dir, check.file)); err == nil {
			info.PackageMgr = check.pkgMgr
			info.Type = check.projType
			info.DevCmd = check.devCmd
			info.BuildCmd = check.buildCmd
			info.TestCmd = check.testCmd
			return
		}
	}
}

func analyzeDocumentation(dir string, info *ProjectInfo) {
	// Read README.md
	readmeContent := readFileIfExists(filepath.Join(dir, "README.md"))
	if readmeContent == "" {
		readmeContent = readFileIfExists(filepath.Join(dir, "readme.md"))
	}

	if readmeContent != "" {
		extractReadmeInstructions(dir, readmeContent, info)

		// Extract title/description
		lines := strings.Split(readmeContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") && info.Description == "" {
				info.Description = strings.TrimPrefix(line, "# ")
			}
			// Look for port mentions
			if strings.Contains(strings.ToLower(line), "port") {
				extractPorts(line, info)
			}
			// Look for environment variables
			if strings.Contains(line, "export ") || strings.Contains(line, "ENV ") {
				extractEnvVars(line, info)
			}
			// Look for commands
			if strings.Contains(line, "npm run") || strings.Contains(line, "go run") {
				extractCommands(line, info)
			}
		}
	}

	// Read other markdown files
	markdownFiles := []string{"SETUP.md", "DEVELOPMENT.md", "CONTRIBUTING.md", "docs/setup.md"}
	for _, file := range markdownFiles {
		content := readFileIfExists(filepath.Join(dir, file))
		if content != "" {
			// Extract environment variables and commands
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "export ") {
					extractEnvVars(line, info)
				}
			}
		}
	}

	// Read .env.example if exists
	envExample := readFileIfExists(filepath.Join(dir, ".env.example"))
	if envExample != "" {
		lines := strings.Split(envExample, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if value == "" {
					value = "changeme"
				}
				info.Environment[key] = value
			}
		}
	}
}

func detectProjectStructure(dir string, info *ProjectInfo) {
	// Common patterns
	patterns := map[string]ServiceInfo{
		"frontend": {Name: "frontend", Type: "web", Port: 3000},
		"ui":       {Name: "ui", Type: "web", Port: 3000},
		"web":      {Name: "web", Type: "web", Port: 3000},
		"client":   {Name: "client", Type: "web", Port: 3000},
		"backend":  {Name: "backend", Type: "api", Port: 8000},
		"api":      {Name: "api", Type: "api", Port: 8000},
		"server":   {Name: "server", Type: "api", Port: 8000},
		"service":  {Name: "service", Type: "api", Port: 8080},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if service, ok := patterns[name]; ok {
			service.Directory = entry.Name()
			// Check if it has its own package.json or go.mod
			if _, err := os.Stat(filepath.Join(dir, entry.Name(), "package.json")); err == nil {
				service.Command = "npm run dev"
			} else if _, err := os.Stat(filepath.Join(dir, entry.Name(), "go.mod")); err == nil {
				service.Command = "go run ."
			}
			info.Services = append(info.Services, service)
		}
	}

	// If no services detected, create a default one
	if len(info.Services) == 0 {
		defaultService := ServiceInfo{
			Name:      "app",
			Type:      "process",
			Command:   info.DevCmd,
			Port:      3000,
			Directory: dir,
		}
		info.Services = append(info.Services, defaultService)
	}
}

func detectServices(dir string, info *ProjectInfo) {
	// Check for docker-compose.yml
	dockerCompose := readFileIfExists(filepath.Join(dir, "docker-compose.yml"))
	if dockerCompose != "" {
		// Parse docker-compose for services
		lines := strings.Split(dockerCompose, "\n")
		inServices := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "services:" {
				inServices = true
				continue
			}
			if inServices && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inServices = false
			}
			if inServices && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
				serviceName := strings.TrimSuffix(trimmed, ":")
				// Don't add if already exists
				exists := false
				for _, s := range info.Services {
					if s.Name == serviceName {
						exists = true
						break
					}
				}
				if !exists {
					info.Services = append(info.Services, ServiceInfo{
						Name:      serviceName,
						Type:      "process",
						Command:   fmt.Sprintf("docker-compose up %s", serviceName),
						Directory: ".",
					})
				}
			}
		}
	}
}

func extractPorts(line string, info *ProjectInfo) {
	// Simple port extraction (e.g., "port 3000", ":3000", "PORT=3000")
	words := strings.Fields(line)
	for i, word := range words {
		word = strings.Trim(word, "`:,")
		if strings.HasPrefix(word, "PORT=") {
			word = strings.TrimPrefix(word, "PORT=")
		}
		if len(word) == 4 || len(word) == 5 {
			// Could be a port number
			if i > 0 && strings.ToLower(words[i-1]) == "port" {
				// Found "port XXXX"
				if len(info.Services) > 0 {
					// Assign to first service that doesn't have a port
					for idx := range info.Services {
						if info.Services[idx].Port == 0 {
							fmt.Sscanf(word, "%d", &info.Services[idx].Port)
							break
						}
					}
				}
			}
		}
	}
}

func extractEnvVars(line string, info *ProjectInfo) {
	// Look for export VAR=value or ENV VAR=value
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "export ") {
		line = strings.TrimPrefix(line, "export ")
	} else if strings.HasPrefix(line, "ENV ") {
		line = strings.TrimPrefix(line, "ENV ")
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) == 2 {
		key := strings.TrimSpace(parts[0])
		// Only capture uppercase env vars
		if key == strings.ToUpper(key) && len(key) > 1 {
			if _, exists := info.Environment[key]; !exists {
				info.Environment[key] = "changeme"
			}
		}
	}
}

func extractCommands(line string, info *ProjectInfo) {
	// Update commands if found in documentation
	if info.DevCmd != "" {
		return
	}

	switch {
	case strings.Contains(line, "npm run dev"):
		info.DevCmd = "npm run dev"
	case strings.Contains(line, "npm start"):
		info.DevCmd = "npm start"
	case strings.Contains(line, "yarn dev"):
		info.DevCmd = "yarn dev"
	case strings.Contains(line, "yarn start"):
		info.DevCmd = "yarn start"
	case strings.Contains(line, "pnpm dev"):
		info.DevCmd = "pnpm dev"
	case strings.Contains(line, "pnpm start"):
		info.DevCmd = "pnpm start"
	case strings.Contains(line, "go run"):
		info.DevCmd = "go run ."
	case strings.Contains(line, "docker compose up"):
		info.DevCmd = "docker compose up"
	case strings.Contains(line, "docker-compose up"):
		info.DevCmd = "docker-compose up"
	case strings.Contains(line, "make dev"):
		info.DevCmd = "make dev"
	case strings.Contains(line, "make start"):
		info.DevCmd = "make start"
	case strings.Contains(line, "uvicorn "):
		info.DevCmd = "uvicorn app:app --reload"
	case strings.Contains(line, "flask run"):
		info.DevCmd = "flask run"
	case strings.Contains(line, "rails s"):
		info.DevCmd = "rails s"
	}
}

func promptProjectInfo(info *ProjectInfo) error {
	reader := bufio.NewReader(os.Stdin)

	// Project name
	fmt.Printf("Project name [%s]: ", info.Name)
	if input := readLine(reader); input != "" {
		info.Name = input
	}

	// Description
	if info.Description == "" {
		info.Description = "My application"
	}
	fmt.Printf("Description [%s]: ", info.Description)
	if input := readLine(reader); input != "" {
		info.Description = input
	}

	// Review detected services
	if len(info.Services) > 0 {
		fmt.Println()
		fmt.Printf("Detected %d service(s):\n", len(info.Services))
		for i, svc := range info.Services {
			fmt.Printf("  %d. %s (%s) - Command: %s\n", i+1, svc.Name, svc.Type, svc.Command)
		}
		fmt.Print("\nKeep these services? [Y/n]: ")
		if input := readLine(reader); strings.ToLower(input) == "n" {
			info.Services = []ServiceInfo{}
		}
	}

	return nil
}

func generateConfig(info *ProjectInfo) string {
	var sb strings.Builder

	appWorkDir := info.RootDir
	if appWorkDir == "" {
		appWorkDir = getCurrentDir()
	}

	sb.WriteString("version: \"1.0\"\n\n")
	sb.WriteString("apps:\n")
	sb.WriteString(fmt.Sprintf("  %s:\n", strings.ToLower(strings.ReplaceAll(info.Name, " ", "-"))))
	sb.WriteString(fmt.Sprintf("    name: \"%s\"\n", info.Name))
	sb.WriteString(fmt.Sprintf("    description: \"%s\"\n", info.Description))
	sb.WriteString(fmt.Sprintf("    workdir: \"%s\"\n\n", appWorkDir))

	// Services
	sb.WriteString("    services:\n")
	if len(info.Services) == 0 {
		// Default service
		sb.WriteString("      - name: app\n")
		sb.WriteString("        type: process\n")
		if info.DevCmd != "" {
			sb.WriteString(fmt.Sprintf("        command: \"%s\"\n", info.DevCmd))
		} else {
			sb.WriteString("        command: \"echo 'Update this command'\"\n")
		}
		sb.WriteString("        port: 3000\n")
		sb.WriteString("        logfile: \"logs/app.log\"\n")
	} else {
		for _, svc := range info.Services {
			sb.WriteString(fmt.Sprintf("      - name: %s\n", svc.Name))
			sb.WriteString("        type: process\n")
			if svc.Command != "" {
				sb.WriteString(fmt.Sprintf("        command: \"%s\"\n", svc.Command))
			} else {
				sb.WriteString("        command: \"echo 'Update this command'\"\n")
			}
			serviceDir := resolveAbsolutePath(appWorkDir, svc.Directory)
			if serviceDir != "" && serviceDir != appWorkDir {
				sb.WriteString(fmt.Sprintf("        workdir: \"%s\"\n", serviceDir))
			}
			if svc.Port > 0 {
				sb.WriteString(fmt.Sprintf("        port: %d\n", svc.Port))
			}
			sb.WriteString(fmt.Sprintf("        logfile: \"logs/%s.log\"\n", svc.Name))

			// Add dependencies if multiple services
			if len(info.Services) > 1 && svc.Name == "frontend" {
				sb.WriteString("        dependencies:\n")
				for _, dep := range info.Services {
					if dep.Name != svc.Name && dep.Type == "api" {
						sb.WriteString(fmt.Sprintf("          - %s\n", dep.Name))
					}
				}
			}

			// Add environment variables
			if len(info.Environment) > 0 {
				sb.WriteString("        environment:\n")
				for key, val := range info.Environment {
					sb.WriteString(fmt.Sprintf("          %s: \"%s\"\n", key, val))
				}
			}

			// Add health check for services with ports
			if svc.Port > 0 {
				sb.WriteString("        healthcheck:\n")
				if svc.Type == "web" || svc.Type == "api" {
					sb.WriteString("          type: http\n")
					sb.WriteString(fmt.Sprintf("          endpoint: \"http://localhost:%d\"\n", svc.Port))
				} else {
					sb.WriteString("          type: tcp\n")
					sb.WriteString(fmt.Sprintf("          endpoint: \"localhost:%d\"\n", svc.Port))
				}
				sb.WriteString("          timeout: 30s\n")
				sb.WriteString("          interval: 5s\n")
				sb.WriteString("          retries: 6\n")
			}

			sb.WriteString("\n")
		}
	}

	// Modes
	sb.WriteString("    modes:\n")
	sb.WriteString("      default:\n")
	sb.WriteString("        name: \"Development\"\n")
	sb.WriteString("        description: \"Full development environment\"\n")
	sb.WriteString("        services:\n")
	if len(info.Services) == 0 {
		sb.WriteString("          - app\n")
	} else {
		for _, svc := range info.Services {
			sb.WriteString(fmt.Sprintf("          - %s\n", svc.Name))
		}
	}

	// Add hooks
	sb.WriteString("\n    hooks:\n")
	sb.WriteString("      pre_start:\n")
	sb.WriteString("        - \"mkdir -p logs\"\n")
	sb.WriteString("      post_start:\n")
	sb.WriteString("        - \"echo 'Application started successfully'\"\n")

	appendInstallSetup(&sb, info, appWorkDir)

	return sb.String()
}

func readFileIfExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "my-app"
	}
	return dir
}

func readLine(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func extractReadmeInstructions(projectRoot, content string, info *ProjectInfo) {
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	section := ""
	currentWorkDir := ""
	usedNames := make(map[string]int)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			currentWorkDir = ""
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			section = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			continue
		}

		if !inCodeBlock || trimmed == "" {
			continue
		}

		cmd := strings.TrimSpace(trimmed)
		if strings.HasPrefix(cmd, "$") || strings.HasPrefix(cmd, ">") {
			cmd = strings.TrimSpace(cmd[1:])
		}
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}

		if dir, rest, ok := splitChainedCD(cmd); ok {
			currentWorkDir = resolveWorkDir(projectRoot, currentWorkDir, dir)
			cmd = rest
			if cmd == "" {
				continue
			}
		}

		if strings.HasPrefix(cmd, "cd ") {
			currentWorkDir = resolveWorkDir(projectRoot, currentWorkDir, strings.TrimSpace(strings.TrimPrefix(cmd, "cd ")))
			continue
		}

		workDir := projectRoot
		if currentWorkDir != "" {
			workDir = currentWorkDir
		}

		switch {
		case isInstallCommand(cmd):
			addCommand(&info.InstallCmds, cmd, workDir)
		case isSetupCommand(cmd):
			addCommand(&info.SetupCmds, cmd, workDir)
		case isRunCommand(cmd):
			name := inferServiceName(section, cmd)
			name = ensureUniqueServiceName(name, usedNames, info.Services)
			info.Services = append(info.Services, ServiceInfo{
				Name:      name,
				Type:      "process",
				Command:   cmd,
				Directory: workDir,
			})
		}
	}
}

func splitChainedCD(cmd string) (string, string, bool) {
	for _, sep := range []string{"&&", ";"} {
		parts := strings.SplitN(cmd, sep, 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			if strings.HasPrefix(left, "cd ") {
				dir := strings.TrimSpace(strings.TrimPrefix(left, "cd "))
				rest := strings.TrimSpace(parts[1])
				return dir, rest, true
			}
		}
	}
	return "", "", false
}

func resolveWorkDir(projectRoot, current, dir string) string {
	if dir == "" {
		return current
	}
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	base := projectRoot
	if current != "" {
		base = current
	}
	return filepath.Clean(filepath.Join(base, dir))
}

func addCommand(commands *[]CommandInfo, cmd, workDir string) {
	for _, existing := range *commands {
		if existing.Command == cmd && existing.WorkDir == workDir {
			return
		}
	}
	*commands = append(*commands, CommandInfo{
		Command: cmd,
		WorkDir: workDir,
	})
}

func isInstallCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	return strings.Contains(lower, "npm install") ||
		strings.Contains(lower, "npm ci") ||
		strings.Contains(lower, "yarn install") ||
		strings.Contains(lower, "pnpm install") ||
		strings.Contains(lower, "pip install") ||
		strings.Contains(lower, "pip3 install") ||
		strings.Contains(lower, "poetry install") ||
		strings.Contains(lower, "bundle install") ||
		strings.Contains(lower, "composer install") ||
		strings.Contains(lower, "go mod download") ||
		strings.Contains(lower, "go mod tidy")
}

func isSetupCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	return strings.HasPrefix(lower, "python -m venv") ||
		strings.HasPrefix(lower, "python3 -m venv") ||
		strings.HasPrefix(lower, "virtualenv ") ||
		strings.HasPrefix(lower, "uv venv") ||
		strings.Contains(lower, "cp .env") ||
		strings.Contains(lower, "copy .env") ||
		strings.Contains(lower, "cp env") ||
		strings.Contains(lower, "cp config") ||
		strings.Contains(lower, "source ") ||
		strings.Contains(lower, "export ")
}

func isRunCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	return strings.Contains(lower, "npm run dev") ||
		strings.Contains(lower, "npm start") ||
		strings.Contains(lower, "yarn dev") ||
		strings.Contains(lower, "yarn start") ||
		strings.Contains(lower, "pnpm dev") ||
		strings.Contains(lower, "pnpm start") ||
		strings.Contains(lower, "go run") ||
		strings.Contains(lower, "air ") ||
		strings.Contains(lower, "uvicorn ") ||
		strings.Contains(lower, "flask run") ||
		strings.Contains(lower, "rails s") ||
		strings.Contains(lower, "bundle exec") ||
		strings.Contains(lower, "docker compose up") ||
		strings.Contains(lower, "docker-compose up") ||
		strings.Contains(lower, "make dev") ||
		strings.Contains(lower, "make start")
}

func inferServiceName(section, cmd string) string {
	if section != "" {
		return sanitizeServiceName(section)
	}

	lower := strings.ToLower(cmd)
	switch {
	case strings.Contains(lower, "frontend") || strings.Contains(lower, "client"):
		return "frontend"
	case strings.Contains(lower, "backend") || strings.Contains(lower, "api"):
		return "backend"
	case strings.Contains(lower, "server"):
		return "server"
	default:
		return "app"
	}
}

func sanitizeServiceName(name string) string {
	name = strings.ToLower(name)
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "app"
	}
	return result
}

func ensureUniqueServiceName(name string, used map[string]int, existing []ServiceInfo) string {
	if name == "" {
		name = "app"
	}
	if used[name] == 0 {
		for _, svc := range existing {
			if svc.Name == name {
				used[name] = 1
				break
			}
		}
	}
	if used[name] == 0 {
		used[name] = 1
		return name
	}
	used[name]++
	return fmt.Sprintf("%s-%d", name, used[name])
}

func resolveAbsolutePath(baseDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}

func appendInstallSetup(sb *strings.Builder, info *ProjectInfo, appWorkDir string) {
	installSteps := groupCommandsByWorkdir(info.InstallCmds, appWorkDir)
	if len(installSteps) == 0 {
		switch info.PackageMgr {
		case "npm":
			installSteps = append(installSteps, InstallStepSpec{
				Name:     "Install dependencies",
				WorkDir:  appWorkDir,
				Commands: []string{"npm install"},
			})
		case "go":
			installSteps = append(installSteps, InstallStepSpec{
				Name:     "Install dependencies",
				WorkDir:  appWorkDir,
				Commands: []string{"go mod download"},
			})
		case "pip":
			installSteps = append(installSteps, InstallStepSpec{
				Name:     "Install dependencies",
				WorkDir:  appWorkDir,
				Commands: []string{"pip install -r requirements.txt"},
			})
		}
	}

	if len(installSteps) > 0 {
		sb.WriteString("\n    install:\n")
		sb.WriteString("      steps:\n")
		for _, step := range installSteps {
			sb.WriteString(fmt.Sprintf("        - name: \"%s\"\n", step.Name))
			if step.WorkDir != "" && step.WorkDir != appWorkDir {
				sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", step.WorkDir))
			}
			sb.WriteString("          commands:\n")
			for _, cmd := range step.Commands {
				sb.WriteString(fmt.Sprintf("            - \"%s\"\n", cmd))
			}
		}
	}

	setupScripts := formatSetupScripts(info.SetupCmds, appWorkDir)
	if len(setupScripts) == 0 && info.PackageMgr == "pip" {
		setupScripts = append(setupScripts, "python -m venv .venv")
	}
	if len(setupScripts) > 0 {
		sb.WriteString("\n    setup:\n")
		sb.WriteString("      scripts:\n")
		for _, script := range setupScripts {
			sb.WriteString(fmt.Sprintf("        - \"%s\"\n", script))
		}
	}
}

type InstallStepSpec struct {
	Name     string
	WorkDir  string
	Commands []string
}

func groupCommandsByWorkdir(commands []CommandInfo, appWorkDir string) []InstallStepSpec {
	if len(commands) == 0 {
		return nil
	}

	order := []string{}
	buckets := make(map[string][]string)
	for _, cmd := range commands {
		workDir := cmd.WorkDir
		if workDir == "" {
			workDir = appWorkDir
		}
		if _, exists := buckets[workDir]; !exists {
			order = append(order, workDir)
		}
		buckets[workDir] = append(buckets[workDir], cmd.Command)
	}

	steps := make([]InstallStepSpec, 0, len(order))
	for _, workDir := range order {
		name := "Install dependencies"
		if workDir != appWorkDir {
			name = fmt.Sprintf("Install dependencies (%s)", filepath.Base(workDir))
		}
		steps = append(steps, InstallStepSpec{
			Name:     name,
			WorkDir:  workDir,
			Commands: buckets[workDir],
		})
	}

	return steps
}

func formatSetupScripts(commands []CommandInfo, appWorkDir string) []string {
	if len(commands) == 0 {
		return nil
	}

	scripts := []string{}
	for _, cmd := range commands {
		workDir := cmd.WorkDir
		if workDir == "" || workDir == appWorkDir {
			scripts = append(scripts, cmd.Command)
			continue
		}
		scripts = append(scripts, fmt.Sprintf("cd %s && %s", workDir, cmd.Command))
	}
	return scripts
}

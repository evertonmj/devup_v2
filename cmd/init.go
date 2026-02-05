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
	onlyInit        bool
	onlyInstall     bool
	onlySetup       bool
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
- Run installation phase to download dependencies
- Run setup phase to prepare the environment

Examples:
  devup init                           # Full setup: scan, install, and setup everything
  devup init --only-init               # Only create devup.yaml
  devup init --only-install            # Only run installation phase
  devup init --only-setup              # Only run setup phase
  devup init --scan --force            # Auto-detect and overwrite existing config`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&initInteractive, "interactive", "i", true, "Interactive mode with prompts")
	initCmd.Flags().BoolVar(&initScan, "scan", true, "Scan project directory for patterns")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing devup.yaml")
	initCmd.Flags().BoolVar(&onlyInit, "only-init", false, "Only create devup.yaml, skip install and setup")
	initCmd.Flags().BoolVar(&onlyInstall, "only-install", false, "Only run installation phase")
	initCmd.Flags().BoolVar(&onlySetup, "only-setup", false, "Only run setup phase")
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "devup.yaml"

	// Determine which phases to run
	runInitPhase := !onlyInstall && !onlySetup
	runInstallPhase := !onlyInit && !onlySetup
	runSetupPhase := !onlyInit && !onlyInstall

	// Check if devup.yaml already exists
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
	}

	// Handle existing config file
	if configExists && !initForce {
		// If user wants only init phase, error out
		if onlyInit {
			return fmt.Errorf("devup.yaml already exists. Use --force to overwrite")
		}
		// For full init or when install/setup requested, skip init phase but continue
		if runInitPhase {
			fmt.Println("ℹ️  devup.yaml already exists, skipping file generation...")
			fmt.Println("   (Use --force to regenerate the file)")
			fmt.Println()
			runInitPhase = false
		}
	}

	// Phase 1: Initialize (create devup.yaml)
	if runInitPhase {
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
		fmt.Println("✅ Configuration created successfully")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
	}

	// Phase 2: Install
	if runInstallPhase {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("📦 Installing dependencies...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// Call install command - set local flag to use the just-created config
		savedLocal := local
		local = true
		installCmd := rootCmd.Commands()
		var install *cobra.Command
		for _, c := range installCmd {
			if c.Name() == "install" {
				install = c
				break
			}
		}
		if install == nil {
			local = savedLocal
			return fmt.Errorf("install command not found")
		}
		if err := install.RunE(install, []string{}); err != nil {
			local = savedLocal
			return fmt.Errorf("failed to install: %w", err)
		}
		local = savedLocal

		fmt.Println()
		fmt.Println("✅ Installation completed")
		fmt.Println()
	}

	// Phase 3: Setup
	if runSetupPhase {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("🔧 Running setup...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// Call setup command - set local flag to use the just-created config
		savedLocal := local
		local = true
		setupCmd := rootCmd.Commands()
		var setup *cobra.Command
		for _, c := range setupCmd {
			if c.Name() == "setup" {
				setup = c
				break
			}
		}
		if setup == nil {
			local = savedLocal
			return fmt.Errorf("setup command not found")
		}
		if err := setup.RunE(setup, []string{}); err != nil {
			local = savedLocal
			return fmt.Errorf("failed to setup: %w", err)
		}
		local = savedLocal

		fmt.Println()
		fmt.Println("✅ Setup completed")
		fmt.Println()
	}

	// Show summary
	if !onlyInit && !onlyInstall && !onlySetup {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✅ Project fully initialized and ready to use!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Println("📝 Next steps:")
		fmt.Println("  1. Review and customize devup.yaml if needed")
		fmt.Println("  2. Start your application: devup start")
		fmt.Println("  3. View status: devup status")
		fmt.Println()
		fmt.Println("📚 For help, see: https://github.com/evertonmj/devup_v2/blob/main/docs/TUTORIAL.md")
		fmt.Println()
	}

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
  VenvAppScope    bool              // when true, activate Python venv for entire app (all services including frontend)
	PythonVersion   string            // Python executable for venv (e.g. "python3", "python3.11")
	RuntimeVersions map[string]string 
}

// ServiceInfo holds information about a detected service
type ServiceInfo struct {
	Name         string
	Type         string
	Command      string
	Port         int
	Directory    string
	Dependencies []string
	// Optional Docker settings when the service runs as a container
	DockerImage       string
	DockerContainer   string
	DockerPorts       []string
	DockerVolumes     []string
	DockerEnvironment map[string]string
	// Language hint (e.g., "python", "node", "go")
	Language string
	// Framework hint (e.g., "uvicorn", "flask", "django") used to set command
	Framework string
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

	// Scan .env style files
	scanEnvFiles(dir, info)

	// Detect common project structures
	if len(info.Services) == 0 {
		detectProjectStructure(absDir, info)
	}

	// Detect ports and services
	if len(info.Services) == 0 {
		detectServices(absDir, info)
	}

	// Detect frameworks and set start commands accordingly (e.g. uvicorn, flask, django)
	detectFrameworks(dir, info)

	// Infer related backing services like databases
	detectDatabaseServices(dir, info)

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

// serviceHints defines keywords to match dir names (prefix, suffix, or contains) and the resulting service.
// Order matters: longer/more specific hints first (e.g. "frontend" before "front").
var serviceHints = []struct {
	keyword string // matches as prefix, suffix, or word in dir name
	svc     ServiceInfo
}{
	{"frontend", ServiceInfo{Name: "frontend", Type: "web", Port: 3000}},
	{"backend", ServiceInfo{Name: "backend", Type: "api", Port: 8080}},
	{"fullstack", ServiceInfo{Name: "app", Type: "api", Port: 3000}},
	{"restapi", ServiceInfo{Name: "api", Type: "api", Port: 8080}},
	{"rest-api", ServiceInfo{Name: "api", Type: "api", Port: 8080}},
	{"graphql", ServiceInfo{Name: "api", Type: "api", Port: 8080}},
	{"service", ServiceInfo{Name: "service", Type: "api", Port: 8080}},
	{"services", ServiceInfo{Name: "service", Type: "api", Port: 8080}},
	{"api", ServiceInfo{Name: "api", Type: "api", Port: 8080}},
	{"server", ServiceInfo{Name: "server", Type: "api", Port: 8080}},
	{"client", ServiceInfo{Name: "client", Type: "web", Port: 3000}},
	{"ui", ServiceInfo{Name: "ui", Type: "web", Port: 3000}},
	{"web", ServiceInfo{Name: "web", Type: "web", Port: 3000}},
	{"app", ServiceInfo{Name: "app", Type: "api", Port: 8080}},
	{"front", ServiceInfo{Name: "frontend", Type: "web", Port: 3000}},
	{"back", ServiceInfo{Name: "backend", Type: "api", Port: 8080}},
	{"svc", ServiceInfo{Name: "service", Type: "api", Port: 8080}},
	{"fe", ServiceInfo{Name: "frontend", Type: "web", Port: 3000}},
	{"be", ServiceInfo{Name: "backend", Type: "api", Port: 8080}},
}

// matchesServiceHint returns true if dirName matches the keyword (prefix, suffix, or contains as word).
func matchesServiceHint(dirName, keyword string) bool {
	if len(keyword) > len(dirName) {
		return false
	}
	if dirName == keyword {
		return true
	}
	sep := "-" + keyword
	if strings.HasPrefix(dirName, keyword+"-") || strings.HasPrefix(dirName, keyword+"_") {
		return true
	}
	if strings.HasSuffix(dirName, "-"+keyword) || strings.HasSuffix(dirName, "_"+keyword) {
		return true
	}
	// contains as a word (surrounded by - or _ or at boundary)
	if strings.Contains(dirName, sep) || strings.Contains(dirName, "_"+keyword+"_") || strings.Contains(dirName, "_"+keyword+"-") {
		return true
	}
	if strings.Contains(dirName, "-"+keyword+"_") || strings.Contains(dirName, "-"+keyword+"-") {
		return true
	}
	return false
}

func detectProjectStructure(dir string, info *ProjectInfo) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		for _, h := range serviceHints {
			if matchesServiceHint(name, h.keyword) {
				service := h.svc
				service.Name = deriveServiceName(entry.Name(), h.keyword, h.svc.Name)
				service.Directory = entry.Name()
				subdir := filepath.Join(dir, entry.Name())
				service = inferServiceStack(subdir, service)
				info.Services = append(info.Services, service)
				break
			}
		}
	}

	// Fallback: scan all subdirs for package manifests (catches any subproject)
	detectSubprojects(dir, info)

	// Ensure unique service names (e.g. user-service + auth-service both deriving to "service")
	ensureUniqueServiceNames(info)

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

// inferServiceStack detects the stack (npm, maven, etc.) in subdir and sets Command/Language.
func inferServiceStack(subdir string, svc ServiceInfo) ServiceInfo {
	checks := []struct {
		file     string
		cmd      string
		language string
	}{
		{"package.json", "npm run dev", "node"},
		{"go.mod", "go run .", "go"},
		{"requirements.txt", "python main.py", "python"},
		{"pyproject.toml", "python main.py", "python"},
		{"pom.xml", "mvn spring-boot:run", "java"},
		{"build.gradle", "./gradlew bootRun", "java"},
		{"build.gradle.kts", "./gradlew bootRun", "java"},
		{"Cargo.toml", "cargo run", "rust"},
		{"Gemfile", "bundle exec rails s", "ruby"},
		{"composer.json", "php artisan serve", "php"},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(subdir, c.file)); err == nil {
			svc.Command = c.cmd
			svc.Language = c.language
			if c.language == "node" {
				if pkg := readFileIfExists(filepath.Join(subdir, "package.json")); strings.Contains(pkg, "react-scripts") {
					svc.Command = "npm start"
				}
			}
			return svc
		}
	}
	return svc
}

// detectSubprojects scans all subdirectories for package manifests and adds them as services.
// Skips directories already present in info.Services.
func detectSubprojects(dir string, info *ProjectInfo) {
	existingDirs := make(map[string]bool)
	for _, s := range info.Services {
		if s.Directory != "" && s.Directory != "." {
			existingDirs[s.Directory] = true
		}
	}

	manifests := []struct {
		file        string
		cmd         string
		language    string
		svcType     string
		defaultPort int
	}{
		{"package.json", "npm run dev", "node", "web", 3000},
		{"pom.xml", "mvn spring-boot:run", "java", "api", 8080},
		{"build.gradle", "./gradlew bootRun", "java", "api", 8080},
		{"build.gradle.kts", "./gradlew bootRun", "java", "api", 8080},
		{"go.mod", "go run .", "go", "api", 8080},
		{"requirements.txt", "python main.py", "python", "api", 8000},
		{"pyproject.toml", "python main.py", "python", "api", 8000},
		{"Cargo.toml", "cargo run", "rust", "api", 8080},
		{"Gemfile", "bundle exec rails s", "ruby", "api", 3000},
		{"composer.json", "php artisan serve", "php", "api", 8000},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if existingDirs[entry.Name()] {
			continue
		}
		subdir := filepath.Join(dir, entry.Name())
		for _, m := range manifests {
			if _, err := os.Stat(filepath.Join(subdir, m.file)); err == nil {
				name := deriveServiceName(entry.Name(), "", "api")
				cmd := m.cmd
				svcType := m.svcType
				if m.language == "node" {
					if pkg := readFileIfExists(filepath.Join(subdir, "package.json")); strings.Contains(pkg, "react-scripts") {
						cmd = "npm start"
						svcType = "web"
					}
				}
				info.Services = append(info.Services, ServiceInfo{
					Name:      name,
					Type:      svcType,
					Command:   cmd,
					Port:      m.defaultPort,
					Directory: entry.Name(),
					Language:  m.language,
				})
				existingDirs[entry.Name()] = true
				break
			}
		}
	}
}

// deriveServiceName returns a unique service name from directory name, tied to the scanned app.
// When matchedKeyword is set, extracts the distinguishing part for suffix matches (e.g. "user" from "user-service")
// to avoid duplicates when multiple dirs map to the same hint (user-service, auth-service -> "service").
// defaultName is used for prefix/exact matches and when no hint matched (for detectSubprojects).
func deriveServiceName(dirName, matchedKeyword, defaultName string) string {
	lower := strings.ToLower(dirName)

	// For suffix match with "service"/"api" (often duplicated), extract qualifier: "user-service" -> "user"
	genericKeywords := map[string]bool{"service": true, "services": true, "svc": true, "api": true, "app": true}
	if matchedKeyword != "" && genericKeywords[matchedKeyword] &&
		(strings.HasSuffix(lower, "-"+matchedKeyword) || strings.HasSuffix(lower, "_"+matchedKeyword)) {
		before := strings.TrimSuffix(lower, "-"+matchedKeyword)
		before = strings.TrimSuffix(before, "_"+matchedKeyword)
		if before != "" {
			if idx := strings.LastIndexAny(before, "-_"); idx >= 0 && idx < len(before)-1 {
				return before[idx+1:]
			}
			return before
		}
	}

	// Prefix or exact match: use default (e.g. "frontend", "backend")
	if matchedKeyword != "" {
		return defaultName
	}

	// No hint matched: use first meaningful part from directory
	parts := strings.FieldsFunc(dirName, func(r rune) bool { return r == '-' || r == '_' })
	for _, p := range parts {
		p = strings.ToLower(p)
		if p != "" && p != "service" && p != "app" && len(p) <= 20 {
			return p
		}
	}
	if len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	if defaultName != "" {
		return defaultName
	}
	return "app"
}

// ensureUniqueServiceNames deduplicates service names using directory-derived names when needed.
func ensureUniqueServiceNames(info *ProjectInfo) {
	used := make(map[string]bool)
	for i := range info.Services {
		name := info.Services[i].Name
		dir := info.Services[i].Directory
		base := name
		for used[name] {
			// Disambiguate: use distinguishing part from directory (e.g. "user" from "user-service")
			parts := strings.FieldsFunc(dir, func(r rune) bool { return r == '-' || r == '_' })
			found := false
			for _, p := range parts {
				cand := strings.ToLower(p)
				if cand != "" && cand != base && !used[cand] {
					name = cand
					found = true
					break
				}
			}
			if !found {
				// Use base + first dir segment: "service" + "user" -> "service-user"
				if len(parts) > 0 {
					name = base + "-" + strings.ToLower(parts[0])
				} else {
					slug := strings.ToLower(strings.Trim(dir, "-_"))
					if slug == "" {
						slug = "app"
					}
					name = base + "-" + slug
				}
			}
			if !used[name] {
				break
			}
			// Numeric suffix as last resort
			for n := 1; n < 100; n++ {
				cand := fmt.Sprintf("%s-%d", base, n)
				if !used[cand] {
					name = cand
					break
				}
			}
		}
		used[name] = true
		info.Services[i].Name = name
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

// detectFrameworks detects web/server frameworks (e.g. uvicorn, flask, django) per service
// and overrides the service command so they are started accordingly.
func detectFrameworks(dir string, info *ProjectInfo) {
	for i := range info.Services {
		svc := &info.Services[i]
		if svc.Type == "docker" && svc.DockerImage != "" {
			continue
		}
		base := dir
		if svc.Directory != "" && svc.Directory != "." {
			base = filepath.Join(dir, svc.Directory)
		}
		detectPythonFramework(base, svc)
	}
}

// detectPythonFramework checks requirements.txt/pyproject.toml and Python files in base
// for uvicorn/fastapi, flask, or django, and sets svc.Command and svc.Framework.
func detectPythonFramework(base string, svc *ServiceInfo) {
	reqPath := filepath.Join(base, "requirements.txt")
	pyprojectPath := filepath.Join(base, "pyproject.toml")
	req := readFileIfExists(reqPath)
	pyproject := readFileIfExists(pyprojectPath)
	if req == "" && pyproject == "" {
		return
	}
	lower := strings.ToLower(req + "\n" + pyproject)

	hasUvicorn := strings.Contains(lower, "uvicorn") || strings.Contains(lower, "fastapi")
	hasFlask := strings.Contains(lower, "flask")
	hasDjango := strings.Contains(lower, "django")
	managePy := readFileIfExists(filepath.Join(base, "manage.py")) != ""

	// Prefer uvicorn/fastapi, then flask, then django
	if hasUvicorn {
		module, appName := resolveUvicornApp(base)
		cmd := fmt.Sprintf("uvicorn %s:%s --reload --host 0.0.0.0", module, appName)
		svc.Command = cmd
		svc.Framework = "uvicorn"
		return
	}
	if hasFlask {
		svc.Command = "flask run --host=0.0.0.0"
		svc.Framework = "flask"
		return
	}
	if hasDjango || managePy {
		port := 8000
		if svc.Port > 0 {
			port = svc.Port
		}
		svc.Command = fmt.Sprintf("python manage.py runserver 0.0.0.0:%d", port)
		svc.Framework = "django"
	}
}

// resolveUvicornApp scans Python files for FastAPI/Starlette app and returns "module:app".
// Defaults to "main:app" if not found.
func resolveUvicornApp(base string) (module, appName string) {
	candidates := []string{"main.py", "app.py", "app/main.py", "src/main.py", "application.py"}
	for _, rel := range candidates {
		p := filepath.Join(base, rel)
		data := readFileIfExists(p)
		if data == "" {
			continue
		}
		mod := strings.TrimSuffix(rel, ".py")
		mod = strings.ReplaceAll(mod, string(filepath.Separator), ".")
		// Look for app = FastAPI(...) or app = Application(...) or similar
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			// app = FastAPI( or app = APIRouter( typically not the asgi app
			if strings.Contains(line, "FastAPI(") || strings.Contains(line, "Application(") {
				name := extractAppVarName(line)
				if name != "" {
					return mod, name
				}
			}
		}
	}
	// Fallback: scan any *.py in base for FastAPI
	entries, err := os.ReadDir(base)
	if err != nil {
		return "main", "app"
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".py") {
			continue
		}
		data := readFileIfExists(filepath.Join(base, e.Name()))
		if data == "" {
			continue
		}
		mod := strings.TrimSuffix(e.Name(), ".py")
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.Contains(line, "FastAPI(") || strings.Contains(line, "Application(") {
				name := extractAppVarName(line)
				if name != "" {
					return mod, name
				}
			}
		}
	}
	return "main", "app"
}

func extractAppVarName(line string) string {
	// "app = FastAPI()" or "application = FastAPI()" or "api = FastAPI()"
	idx := strings.Index(line, "=")
	if idx < 0 {
		return ""
	}
	left := strings.TrimSpace(line[:idx])
	// avoid "if app = ..." etc.
	for _, bad := range []string{"if", "elif", "for", "while", "("} {
		if strings.HasPrefix(left, bad) {
			return ""
		}
	}
	// single identifier
	if strings.Contains(left, " ") || strings.Contains(left, ".") {
		return ""
	}
	if left != "" {
		return left
	}
	return ""
}

// detectDatabaseServices scans project files and environment for common database usage
// and adds a Docker-backed database service to the configuration when appropriate.
func detectDatabaseServices(dir string, info *ProjectInfo) {
	// Helper to check if a service with a given name already exists
	serviceExists := func(name string) bool {
		for _, s := range info.Services {
			if strings.EqualFold(s.Name, name) {
				return true
			}
		}
		return false
	}

	// Signals gathered from files and env vars
	hasPostgres := false
	hasMySQL := false
	hasMongo := false

	// Inspect go.mod for drivers
	if data := readFileIfExists(filepath.Join(dir, "go.mod")); data != "" {
		lower := strings.ToLower(data)
		if strings.Contains(lower, "lib/pq") || strings.Contains(lower, "pgx") || strings.Contains(lower, "gorm.io/driver/postgres") {
			hasPostgres = true
		}
		if strings.Contains(lower, "go-sql-driver/mysql") || strings.Contains(lower, "gorm.io/driver/mysql") {
			hasMySQL = true
		}
	}

	// Inspect package.json dependencies for Node
	if data := readFileIfExists(filepath.Join(dir, "package.json")); data != "" {
		lower := strings.ToLower(data)
		if strings.Contains(lower, "\"pg\"") || strings.Contains(lower, "pg-promise") || strings.Contains(lower, "sequelize") {
			// sequelize may imply either postgres or mysql; keep postgres as a default
			hasPostgres = hasPostgres || strings.Contains(lower, "pg")
		}
		if strings.Contains(lower, "mysql2") || strings.Contains(lower, "\"mysql\"") {
			hasMySQL = true
		}
		if strings.Contains(lower, "mongoose") || strings.Contains(lower, "mongodb") {
			hasMongo = true
		}
	}

	// Inspect Python requirements
	if data := readFileIfExists(filepath.Join(dir, "requirements.txt")); data != "" {
		lower := strings.ToLower(data)
		if strings.Contains(lower, "psycopg2") || strings.Contains(lower, "asyncpg") || strings.Contains(lower, "sqlalchemy[postgresql]") {
			hasPostgres = true
		}
		if strings.Contains(lower, "mysqlclient") || strings.Contains(lower, "pymysql") {
			hasMySQL = true
		}
		if strings.Contains(lower, "pymongo") {
			hasMongo = true
		}
	}

	// Env var hints
	for k := range info.Environment {
		key := strings.ToUpper(k)
		if key == "DATABASE_URL" || key == "POSTGRES_URL" || key == "PGHOST" || key == "PGPORT" || key == "PGUSER" {
			hasPostgres = true
		}
		if key == "DB_HOST" || key == "MYSQL_HOST" || key == "MYSQL_PORT" || key == "MYSQL_USER" {
			hasMySQL = true
		}
		if key == "MONGO_URL" || key == "MONGODB_URI" || key == "MONGO_HOST" {
			hasMongo = true
		}
	}

	// docker-compose hints (service names)
	if data := readFileIfExists(filepath.Join(dir, "docker-compose.yml")); data != "" {
		lower := strings.ToLower(data)
		if strings.Contains(lower, "postgres") || strings.Contains(lower, "postgresql") {
			hasPostgres = true
		}
		if strings.Contains(lower, "mysql") {
			hasMySQL = true
		}
		if strings.Contains(lower, "mongo") {
			hasMongo = true
		}
	}

	// Prefer a single relational database; add service with sensible defaults
	if hasPostgres && !serviceExists("postgres") && !serviceExists("database") {
		env := map[string]string{
			"POSTGRES_USER":     "devuser",
			"POSTGRES_PASSWORD": "devpass",
			"POSTGRES_DB":       "devdb",
		}
		// If user has env vars, use them
		if v, ok := info.Environment["POSTGRES_USER"]; ok && v != "" {
			env["POSTGRES_USER"] = v
		}
		if v, ok := info.Environment["POSTGRES_PASSWORD"]; ok && v != "" {
			env["POSTGRES_PASSWORD"] = v
		}
		if v, ok := info.Environment["POSTGRES_DB"]; ok && v != "" {
			env["POSTGRES_DB"] = v
		}

		info.Services = append(info.Services, ServiceInfo{
			Name:              "postgres",
			Type:              "docker",
			Port:              5432,
			DockerImage:       "postgres:15-alpine",
			DockerContainer:   "devup-postgres",
			DockerPorts:       []string{"5432:5432"},
			DockerVolumes:     []string{"postgres_data:/var/lib/postgresql/data"},
			DockerEnvironment: env,
		})
	} else if hasMySQL && !serviceExists("mysql") && !serviceExists("database") {
		env := map[string]string{
			"MYSQL_ROOT_PASSWORD": "devpass",
			"MYSQL_DATABASE":      "devdb",
			"MYSQL_USER":          "devuser",
			"MYSQL_PASSWORD":      "devpass",
		}
		if v, ok := info.Environment["MYSQL_DATABASE"]; ok && v != "" {
			env["MYSQL_DATABASE"] = v
		}
		if v, ok := info.Environment["MYSQL_USER"]; ok && v != "" {
			env["MYSQL_USER"] = v
		}
		if v, ok := info.Environment["MYSQL_PASSWORD"]; ok && v != "" {
			env["MYSQL_PASSWORD"] = v
		}

		info.Services = append(info.Services, ServiceInfo{
			Name:              "mysql",
			Type:              "docker",
			Port:              3306,
			DockerImage:       "mysql:8",
			DockerContainer:   "devup-mysql",
			DockerPorts:       []string{"3306:3306"},
			DockerVolumes:     []string{"mysql_data:/var/lib/mysql"},
			DockerEnvironment: env,
		})
	} else if hasMongo && !serviceExists("mongo") && !serviceExists("mongodb") {
		env := map[string]string{}
		info.Services = append(info.Services, ServiceInfo{
			Name:              "mongo",
			Type:              "docker",
			Port:              27017,
			DockerImage:       "mongo:6",
			DockerContainer:   "devup-mongo",
			DockerPorts:       []string{"27017:27017"},
			DockerVolumes:     []string{"mongo_data:/data/db"},
			DockerEnvironment: env,
		})
	}
}

func hasPythonServices(info *ProjectInfo) bool {
	if info.PackageMgr == "pip" || info.PackageMgr == "pipenv" {
		return true
	}
	for _, s := range info.Services {
		if strings.EqualFold(s.Language, "python") {
			return true
		}
	}
	return false
}

func hasNodeServices(info *ProjectInfo) bool {
	if info.PackageMgr == "npm" {
		return true
	}
	for _, s := range info.Services {
		if strings.EqualFold(s.Language, "node") {
			return true
		}
	}
	return false
}

func pythonServiceDirs(info *ProjectInfo) []string {
	dirs := []string{}
	if info.PackageMgr == "pip" || info.PackageMgr == "pipenv" {
		dirs = append(dirs, ".")
	}
	for _, s := range info.Services {
		if strings.EqualFold(s.Language, "python") {
			d := s.Directory
			if d == "" {
				d = "."
			}
			// avoid duplicates
			exists := false
			for _, existing := range dirs {
				if existing == d {
					exists = true
					break
				}
			}
			if !exists {
				dirs = append(dirs, d)
			}
		}
	}
	return dirs
}

func extractPorts(line string, info *ProjectInfo) {
	// Simple port extraction (e.g., "port 3000", ":3000", "PORT=3000")
	words := strings.Fields(line)
	for i, word := range words {
		word = strings.Trim(word, "`:,")
		word = strings.TrimPrefix(word, "PORT=")
		if len(word) == 4 || len(word) == 5 {
			// Could be a port number
			if i > 0 && strings.ToLower(words[i-1]) == "port" {
				// Found "port XXXX"
				if len(info.Services) > 0 {
					// Assign to first service that doesn't have a port
					for idx := range info.Services {
						if info.Services[idx].Port == 0 {
							_, _ = fmt.Sscanf(word, "%d", &info.Services[idx].Port)
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

// scanEnvFiles loads common .env files and extracts variables
func scanEnvFiles(dir string, info *ProjectInfo) {
	// Check common .env files in the project root
	files := []string{".env", ".env.local", ".env.development", ".env.example"}

	for _, f := range files {
		path := filepath.Join(dir, f)
		if data := readFileIfExists(path); data != "" {
			parseDotEnvContent(data, info)
		}
	}
}

// parseDotEnvContent parses KEY=VALUE pairs and populates info.Environment
func parseDotEnvContent(content string, info *ProjectInfo) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// tolerate export
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// strip surrounding quotes (handle multiple layers)
		for len(val) >= 2 {
			if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
				(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
				val = val[1 : len(val)-1]
				val = strings.TrimSpace(val) // trim after removing quotes
			} else {
				break
			}
		}
		if key != "" {
			if _, exists := info.Environment[key]; !exists {
				if val == "" {
					val = "changeme"
				}
				info.Environment[key] = val
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

	// Python venv app scope: when Python service(s) detected, offer to activate venv for all services
	if hasPythonServices(info) && len(info.Services) > 0 {
		fmt.Print("\nActivate Python venv for entire app scope (all services including frontend)? [y/N]: ")
		if input := readLine(reader); strings.ToLower(strings.TrimSpace(input)) == "y" || strings.ToLower(strings.TrimSpace(input)) == "yes" {
			info.VenvAppScope = true
			fmt.Print("Python version for venv [python3]: ")
			if v := strings.TrimSpace(readLine(reader)); v != "" {
				info.PythonVersion = v
			} else {
				info.PythonVersion = "python3"
			}
		}
	}

	// Framework/runtime versions: when Node detected, offer to set Node version
	if hasNodeServices(info) {
		if info.RuntimeVersions == nil {
			info.RuntimeVersions = make(map[string]string)
		}
		fmt.Print("\nNode.js version (e.g. 18, 20) [18]: ")
		if v := strings.TrimSpace(readLine(reader)); v != "" {
			info.RuntimeVersions["node"] = v
		} else {
			info.RuntimeVersions["node"] = "18"
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
	workdirAbs, _ := filepath.Abs(".")
	sb.WriteString(fmt.Sprintf("    workdir: \"%s\"\n\n", workdirAbs))

	// Python venv (app scope) when user opted in
	if info.VenvAppScope && hasPythonServices(info) {
		pyVer := info.PythonVersion
		if pyVer == "" {
			pyVer = "python3"
		}
		sb.WriteString("    python:\n")
		sb.WriteString("      venv:\n")
		sb.WriteString(fmt.Sprintf("        version: \"%s\"\n", pyVer))
		sb.WriteString("        dir: \".venv\"\n")
		sb.WriteString("        app_scope: true\n\n")
	}

	// Runtimes: framework versions (node, go, etc.)
	if len(info.RuntimeVersions) > 0 {
		sb.WriteString("    runtimes:\n")
		order := []string{"node", "go", "python"}
		seen := make(map[string]bool)
		for _, k := range order {
			if v, ok := info.RuntimeVersions[k]; ok {
				sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", k, v))
				seen[k] = true
			}
		}
		for k, v := range info.RuntimeVersions {
			if !seen[k] {
				sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", k, v))
			}
		}
		sb.WriteString("\n")
	}

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
			// Docker-backed service
			if svc.Type == "docker" && svc.DockerImage != "" {
				sb.WriteString("        type: docker\n")
				if svc.Port > 0 {
					sb.WriteString(fmt.Sprintf("        port: %d\n", svc.Port))
				}
				sb.WriteString("        docker:\n")
				sb.WriteString(fmt.Sprintf("          image: \"%s\"\n", svc.DockerImage))
				if svc.DockerContainer != "" {
					sb.WriteString(fmt.Sprintf("          container: \"%s\"\n", svc.DockerContainer))
				}
				if len(svc.DockerPorts) > 0 {
					sb.WriteString("          ports:\n")
					for _, p := range svc.DockerPorts {
						sb.WriteString(fmt.Sprintf("            - \"%s\"\n", p))
					}
				}
				if len(svc.DockerVolumes) > 0 {
					sb.WriteString("          volumes:\n")
					for _, v := range svc.DockerVolumes {
						sb.WriteString(fmt.Sprintf("            - \"%s\"\n", v))
					}
				}
				if len(svc.DockerEnvironment) > 0 {
					sb.WriteString("          environment:\n")
					for k, v := range svc.DockerEnvironment {
						sb.WriteString(fmt.Sprintf("            %s: \"%s\"\n", k, v))
					}
				}
				if len(svc.Dependencies) > 0 {
					sb.WriteString("        dependencies:\n")
					for _, dep := range svc.Dependencies {
						sb.WriteString(fmt.Sprintf("          - %s\n", dep))
					}
				}
			} else {
				// Process service
				sb.WriteString("        type: process\n")
				if svc.Command != "" {
					sb.WriteString(fmt.Sprintf("        command: \"%s\"\n", svc.Command))
				} else {
					sb.WriteString("        command: \"echo 'Update this command'\"\n")
				}
				if svc.Directory != "" && svc.Directory != "." {
					// Only set workdir if directory actually exists in project
					if _, err := os.Stat(svc.Directory); err == nil {
						// Use relative path for service workdir to avoid double-joining with app workdir
						sb.WriteString(fmt.Sprintf("        workdir: \"%s\"\n", svc.Directory))
					}
				}
				if svc.Port > 0 {
					sb.WriteString(fmt.Sprintf("        port: %d\n", svc.Port))
				}
				if len(svc.Dependencies) > 0 {
					sb.WriteString("        dependencies:\n")
					for _, dep := range svc.Dependencies {
						sb.WriteString(fmt.Sprintf("          - %s\n", dep))
					}
				}
				sb.WriteString(fmt.Sprintf("        logfile: \"logs/%s.log\"\n", svc.Name))

				// Add environment variables for process services
				if len(info.Environment) > 0 {
					sb.WriteString("        environment:\n")
					for key, val := range info.Environment {
						sb.WriteString(fmt.Sprintf("          %s: %s\n", key, val))
					}
				}
			}

			sb.WriteString("\n")
		}
	}

	// Setup: environment variables and Python virtualenvs if needed
	if len(info.Environment) > 0 || hasPythonServices(info) {
		sb.WriteString("\n    setup:\n")
		if len(info.Environment) > 0 {
			sb.WriteString("      env_vars:\n")
			for key, val := range info.Environment {
				// default simple shape: name + default
				sb.WriteString("        - name: \"" + key + "\"\n")
				if val != "" {
					sb.WriteString("          default: " + val + "\n")
				}
				sb.WriteString("          required: false\n")
			}
		}

		// Python venv setup scripts
		pyDirs := pythonServiceDirs(info)
		pyVer := info.PythonVersion
		if pyVer == "" {
			pyVer = "python3"
		}
		if len(pyDirs) > 0 {
			sb.WriteString("      scripts:\n")
			if info.VenvAppScope {
				// Single app-level venv: create .venv at root, install all Python deps
				sb.WriteString(fmt.Sprintf("        - \"%s -m venv .venv || true\"\n", pyVer))
				for _, d := range pyDirs {
					if d == "" {
						d = "."
					}
					sb.WriteString("        - \"if [ -f " + d + "/requirements.txt ]; then . .venv/bin/activate && pip install -r " + d + "/requirements.txt; fi\"\n")
				}
			} else {
				for _, d := range pyDirs {
					if d == "" {
						d = "."
					}
					sb.WriteString(fmt.Sprintf("        - \"cd %s && %s -m venv .venv || true\"\n", d, pyVer))
					sb.WriteString("        - \"cd " + d + " && if [ -f requirements.txt ]; then . .venv/bin/activate && pip install -r requirements.txt; fi\"\n")
				}
			}
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
	sb.WriteString("\n    install:\n")
	sb.WriteString("      steps:\n")

	// When app-scope venv, create .venv first so Python install steps can use it
	if info.VenvAppScope && hasPythonServices(info) {
		pyVer := info.PythonVersion
		if pyVer == "" {
			pyVer = "python3"
		}
		sb.WriteString("        - name: \"Create app Python venv\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString(fmt.Sprintf("            - \"%s -m venv .venv || true\"\n", pyVer))
	}

	// Root-level package manager steps using commands arrays
	// Always check for common manifests to ensure root installs are included
	if _, err := os.Stat("package.json"); err == nil || info.PackageMgr == "npm" {
		sb.WriteString("        - name: \"Install npm dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"npm install\"\n")
	}
	if _, err := os.Stat("go.mod"); err == nil || info.PackageMgr == "go" {
		sb.WriteString("        - name: \"Download Go modules (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"go mod download\"\n")
	}
	if _, err := os.Stat("requirements.txt"); err == nil || info.PackageMgr == "pip" {
		sb.WriteString("        - name: \"Install Python dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		if info.VenvAppScope {
			sb.WriteString("            - \". .venv/bin/activate && pip install -r requirements.txt\"\n")
		} else {
			sb.WriteString("            - \"pip install -r requirements.txt\"\n")
		}
	}
	if _, err := os.Stat("Pipfile"); err == nil || info.PackageMgr == "pipenv" {
		sb.WriteString("        - name: \"Install Pipenv dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"pipenv install\"\n")
	}
	if _, err := os.Stat("Cargo.toml"); err == nil || info.PackageMgr == "cargo" {
		sb.WriteString("        - name: \"Build Rust dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"cargo build\"\n")
	}
	if _, err := os.Stat("pom.xml"); err == nil || info.PackageMgr == "maven" {
		sb.WriteString("        - name: \"Resolve Maven dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"mvn -q -DskipTests package\"\n")
	}
	if _, err := os.Stat("build.gradle"); err == nil || info.PackageMgr == "gradle" {
		sb.WriteString("        - name: \"Resolve Gradle dependencies (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"gradle build -x test\"\n")
	}
	if _, err := os.Stat("Gemfile"); err == nil || info.PackageMgr == "bundle" {
		sb.WriteString("        - name: \"Install Ruby gems (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"bundle install\"\n")
	}
	if _, err := os.Stat("composer.json"); err == nil || info.PackageMgr == "composer" {
		sb.WriteString("        - name: \"Install Composer packages (root)\"\n")
		sb.WriteString("          commands:\n")
		sb.WriteString("            - \"composer install\"\n")
	}

	// Per-service installs: add a step with workdir for each detected service
	for _, svc := range info.Services {
		wd := svc.Directory
		if wd == "" {
			wd = "."
		}
		// npm
		if _, err := os.Stat(filepath.Join(wd, "package.json")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Install npm dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"npm install\"\n")
		}
		// Go
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Download Go modules (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"go mod download\"\n")
		}
		// Python requirements
		if _, err := os.Stat(filepath.Join(wd, "requirements.txt")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Install Python dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			if info.VenvAppScope {
				relVenv := ".venv"
				if wd != "." {
					relVenv = filepath.Join("..", ".venv")
				}
				sb.WriteString(fmt.Sprintf("            - \". %s/bin/activate && pip install -r requirements.txt\"\n", relVenv))
			} else {
				sb.WriteString("            - \"pip install -r requirements.txt\"\n")
			}
		}
		// Pipenv
		if _, err := os.Stat(filepath.Join(wd, "Pipfile")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Install Pipenv dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"pipenv install\"\n")
		}
		// Cargo
		if _, err := os.Stat(filepath.Join(wd, "Cargo.toml")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Build Rust dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"cargo build\"\n")
		}
		// Maven
		if _, err := os.Stat(filepath.Join(wd, "pom.xml")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Resolve Maven dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"mvn -q -DskipTests package\"\n")
		}
		// Gradle
		if _, err := os.Stat(filepath.Join(wd, "build.gradle")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Resolve Gradle dependencies (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"gradle build -x test\"\n")
		}
		// Ruby Bundler
		if _, err := os.Stat(filepath.Join(wd, "Gemfile")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Install Ruby gems (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"bundle install\"\n")
		}
		// Composer
		if _, err := os.Stat(filepath.Join(wd, "composer.json")); err == nil {
			sb.WriteString(fmt.Sprintf("        - name: \"Install Composer packages (%s)\"\n", svc.Name))
			sb.WriteString(fmt.Sprintf("          workdir: \"%s\"\n", wd))
			sb.WriteString("          commands:\n")
			sb.WriteString("            - \"composer install\"\n")
		}
	}

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

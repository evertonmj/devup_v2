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
	Services    []ServiceInfo
	Environment map[string]string
	PackageMgr  string // npm, go, pip, cargo, etc.
	BuildCmd    string
	TestCmd     string
	DevCmd      string
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
}

func scanProject(dir string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Name:        filepath.Base(getCurrentDir()),
		Services:    []ServiceInfo{},
		Environment: make(map[string]string),
	}

	// Detect package manager and project type
	detectPackageManager(dir, info)

	// Read and analyze documentation
	analyzeDocumentation(dir, info)

	// Scan .env style files
	scanEnvFiles(dir, info)

	// Detect common project structures
	detectProjectStructure(dir, info)

	// Detect ports and services
	detectServices(dir, info)

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
	// Common patterns - only add if directory actually exists
	patterns := map[string]ServiceInfo{
		"frontend": {Name: "frontend", Type: "web", Port: 3000},
		"ui":       {Name: "ui", Type: "web", Port: 3000},
		"web":      {Name: "web", Type: "web", Port: 3000},
		"client":   {Name: "client", Type: "web", Port: 3000},
		"backend":  {Name: "backend", Type: "api", Port: 8000},
		"api":      {Name: "api", Type: "api", Port: 8000},
		"server":   {Name: "server", Type: "api", Port: 8000},
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
				service.Language = "node"
			} else if _, err := os.Stat(filepath.Join(dir, entry.Name(), "go.mod")); err == nil {
				service.Command = "go run ."
				service.Language = "go"
			} else if _, err := os.Stat(filepath.Join(dir, entry.Name(), "requirements.txt")); err == nil {
				service.Command = "python main.py"
				service.Language = "python"
			} else if _, err := os.Stat(filepath.Join(dir, entry.Name(), "pyproject.toml")); err == nil {
				service.Command = "python main.py"
				service.Language = "python"
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
			Directory: ".",
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
	if strings.Contains(line, "npm run dev") && info.DevCmd == "" {
		info.DevCmd = "npm run dev"
	} else if strings.Contains(line, "npm start") && info.DevCmd == "" {
		info.DevCmd = "npm start"
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

	sb.WriteString("version: \"1.0\"\n\n")
	sb.WriteString("apps:\n")
	sb.WriteString(fmt.Sprintf("  %s:\n", strings.ToLower(strings.ReplaceAll(info.Name, " ", "-"))))
	sb.WriteString(fmt.Sprintf("    name: \"%s\"\n", info.Name))
	sb.WriteString(fmt.Sprintf("    description: \"%s\"\n", info.Description))
	workdirAbs, _ := filepath.Abs(".")
	sb.WriteString(fmt.Sprintf("    workdir: \"%s\"\n\n", workdirAbs))

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
		if len(pyDirs) > 0 {
			sb.WriteString("      scripts:\n")
			for _, d := range pyDirs {
				if d == "" {
					d = "."
				}
				sb.WriteString("        - \"cd " + d + " && python3 -m venv .venv || true\"\n")
				// Install requirements only if present
				sb.WriteString("        - \"cd " + d + " && if [ -f requirements.txt ]; then . .venv/bin/activate && pip install -r requirements.txt; fi\"\n")
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
		sb.WriteString("            - \"pip install -r requirements.txt\"\n")
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
			sb.WriteString("            - \"pip install -r requirements.txt\"\n")
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

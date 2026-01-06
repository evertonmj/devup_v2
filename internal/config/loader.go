package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing configuration files
type Loader struct {
	configPath string
	useLocal   bool
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
		useLocal:   false,
	}
}

// NewLoaderWithLocal creates a new configuration loader with local flag
func NewLoaderWithLocal(configPath string, useLocal bool) *Loader {
	return &Loader{
		configPath: configPath,
		useLocal:   useLocal,
	}
}

// Load reads and parses the configuration file
func (l *Loader) Load() (*AppConfig, error) {
	// Resolve config path
	configPath, err := l.resolveConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := l.validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// resolveConfigPath finds the configuration file
func (l *Loader) resolveConfigPath() (string, error) {
	// If explicit path provided, use it
	if l.configPath != "" {
		if _, err := os.Stat(l.configPath); err != nil {
			return "", fmt.Errorf("config file not found at %s", l.configPath)
		}
		return l.configPath, nil
	}

	// If local flag not set, check for DEVUP_DEFAULT_PROJECT environment variable
	if !l.useLocal {
		if envPath := os.Getenv("DEVUP_DEFAULT_PROJECT"); envPath != "" {
			configPath := filepath.Join(envPath, "devup.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}

		// If environment variable not provided or invalid, check saved default project
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "devup")
		defaultProjectFile := filepath.Join(configDir, "default_project")
		if data, err := os.ReadFile(defaultProjectFile); err == nil {
			// Trim whitespace and newlines
			projectPath := string(data)
			projectPath = filepath.Clean(projectPath)
			if projectPath != "" {
				configPath := filepath.Join(projectPath, "devup.yaml")
				if _, err := os.Stat(configPath); err == nil {
					return configPath, nil
				}
			}
		}
	}

	// Search in common locations (current directory first)
	searchPaths := []string{
		"devup.yaml",
		".devup.yaml",
		"config/devup.yaml",
		".config/devup.yaml",
		filepath.Join(os.Getenv("HOME"), ".devup.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "devup", "devup.yaml"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no configuration file found. Searched: %v", searchPaths)
}

// validate performs basic validation on the configuration
func (l *Loader) validate(config *AppConfig) error {
	if len(config.Apps) == 0 {
		return fmt.Errorf("no apps defined in configuration")
	}

	for appName, app := range config.Apps {
		if app.Name == "" {
			app.Name = appName
			config.Apps[appName] = app
		}

		if len(app.Services) == 0 {
			return fmt.Errorf("app '%s' has no services defined", appName)
		}

		// Validate services
		serviceNames := make(map[string]bool)
		for i, service := range app.Services {
			if service.Name == "" {
				return fmt.Errorf("app '%s': service at index %d has no name", appName, i)
			}
			if serviceNames[service.Name] {
				return fmt.Errorf("app '%s': duplicate service name '%s'", appName, service.Name)
			}
			serviceNames[service.Name] = true

			if service.Command == "" {
				return fmt.Errorf("app '%s': service '%s' has no command", appName, service.Name)
			}

			// Validate workdir
			if service.WorkDir != "" {
				workdir := filepath.Join(app.WorkDir, service.WorkDir)
				if _, err := os.Stat(workdir); os.IsNotExist(err) {
					return fmt.Errorf("app '%s': service '%s' workdir '%s' does not exist", appName, service.Name, workdir)
				}
			}

			// Validate healthcheck
			if service.HealthCheck.Type != "" {
				switch service.HealthCheck.Type {
				case "http", "tcp", "exec":
					// valid types
				default:
					return fmt.Errorf("app '%s': service '%s' has invalid healthcheck type '%s'", appName, service.Name, service.HealthCheck.Type)
				}
				if service.HealthCheck.Endpoint == "" && service.HealthCheck.Type != "exec" {
					return fmt.Errorf("app '%s': service '%s' healthcheck endpoint is not set", appName, service.Name)
				}
			}
		}

		// Validate service dependencies
		for _, service := range app.Services {
			for _, dep := range service.Dependencies {
				if !serviceNames[dep] {
					return fmt.Errorf("app '%s': service '%s' has unknown dependency '%s'", appName, service.Name, dep)
				}
			}
		}

		// Validate modes reference existing services
		for modeName, mode := range app.Modes {
			for _, serviceName := range mode.Services {
				if !serviceNames[serviceName] {
					return fmt.Errorf("app '%s': mode '%s' references unknown service '%s'", appName, modeName, serviceName)
				}
			}
		}
	}

	return nil
}

// GetApp retrieves a specific app configuration by name
func (l *Loader) GetApp(config *AppConfig, appName string) (*AppSpec, error) {
	app, exists := config.Apps[appName]
	if !exists {
		return nil, fmt.Errorf("app '%s' not found in configuration", appName)
	}
	return &app, nil
}

// ListApps returns all available app names
func ListApps(config *AppConfig) []string {
	apps := make([]string, 0, len(config.Apps))
	for name := range config.Apps {
		apps = append(apps, name)
	}
	return apps
}

package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TemplateResolver handles variable substitution in configuration strings
type TemplateResolver struct {
	context map[string]string
}

// NewTemplateResolver creates a new template resolver with the given context
func NewTemplateResolver(appName, projectRoot, workdir, cwd string) *TemplateResolver {
	context := map[string]string{
		"HOME":         getHomeDir(),
		"PROJECT_ROOT": projectRoot,
		"WORKDIR":      workdir,
		"APP_NAME":     appName,
		"CWD":          cwd,
	}

	// Add all environment variables to context for fallback
	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx > 0 {
			context[env[:idx]] = env[idx+1:]
		}
	}

	return &TemplateResolver{
		context: context,
	}
}

// ResolveString resolves template variables in a single string
func (tr *TemplateResolver) ResolveString(s string) string {
	if s == "" {
		return s
	}

	// Match ${VAR_NAME} pattern
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := match[2 : len(match)-1]

		// Look up value in context (includes both built-ins and env vars)
		if value, exists := tr.context[varName]; exists {
			return value
		}

		// Return original if variable not found
		return match
	})
}

// ResolveStringMap resolves template variables in a map of strings
func (tr *TemplateResolver) ResolveStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range m {
		result[key] = tr.ResolveString(value)
	}
	return result
}

// ResolveStringSlice resolves template variables in a slice of strings
func (tr *TemplateResolver) ResolveStringSlice(s []string) []string {
	if s == nil {
		return nil
	}

	result := make([]string, len(s))
	for i, value := range s {
		result[i] = tr.ResolveString(value)
	}
	return result
}

// ResolveService resolves template variables in a Service
func (tr *TemplateResolver) ResolveService(svc *Service) *Service {
	if svc == nil {
		return nil
	}

	return &Service{
		Name:         svc.Name,
		Type:         svc.Type,
		Command:      tr.ResolveString(svc.Command),
		WorkDir:      tr.ResolveString(svc.WorkDir),
		Port:         svc.Port,
		Environment:  tr.ResolveStringMap(svc.Environment),
		HealthCheck:  svc.HealthCheck, // HealthCheck doesn't have string fields that need resolution
		LogFile:      tr.ResolveString(svc.LogFile),
		Dependencies: svc.Dependencies,
	}
}

// ResolveAppSpec resolves template variables in an AppSpec
func (tr *TemplateResolver) ResolveAppSpec(app *AppSpec) *AppSpec {
	if app == nil {
		return nil
	}

	// Resolve services
	resolvedServices := make([]Service, len(app.Services))
	for i, svc := range app.Services {
		resolved := tr.ResolveService(&svc)
		resolvedServices[i] = *resolved
	}

	// Resolve modes
	resolvedModes := make(map[string]Mode)
	for modeName, mode := range app.Modes {
		resolvedMode := Mode{
			Name:        mode.Name,
			Description: mode.Description,
			Services:    mode.Services,
			Overrides:   mode.Overrides, // ServiceOverrides would need resolution too
			Environment: tr.ResolveStringMap(mode.Environment),
		}

		// Resolve overrides
		resolvedOverrides := make([]ServiceOverride, len(mode.Overrides))
		for i, override := range mode.Overrides {
			resolvedOverrides[i] = ServiceOverride{
				ServiceName: override.ServiceName,
				Command:     tr.ResolveString(override.Command),
				Environment: tr.ResolveStringMap(override.Environment),
				Port:        override.Port,
			}
		}
		resolvedMode.Overrides = resolvedOverrides
		resolvedModes[modeName] = resolvedMode
	}

	return &AppSpec{
		Name:        app.Name,
		Description: app.Description,
		WorkDir:     tr.ResolveString(app.WorkDir),
		Services:    resolvedServices,
		Modes:       resolvedModes,
		Environment: tr.ResolveStringMap(app.Environment),
		Hooks:       app.Hooks, // Hooks would need resolution for command strings
		Health:      app.Health,
		Install:     app.Install, // Install would need resolution for paths
		Setup:       app.Setup,   // Setup would need resolution for paths
	}
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE") // Windows
		}
	}
	return home
}

// ResolveFilePath resolves template variables and returns an absolute or relative path
func (tr *TemplateResolver) ResolveFilePath(filePath string) string {
	resolved := tr.ResolveString(filePath)

	// If path starts with ~, expand home directory
	if strings.HasPrefix(resolved, "~") {
		home := getHomeDir()
		resolved = filepath.Join(home, resolved[1:])
	}

	return resolved
}

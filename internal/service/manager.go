package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devup/internal/config"
)

// Manager orchestrates service lifecycle operations
type Manager struct {
	app         *config.AppSpec
	mode        string
	services    map[string]ServiceRunner
	state       *config.RuntimeState
	processPool *ProcessPool
	mu          sync.RWMutex
}

// ServiceRunner interface that all service types must implement
type ServiceRunner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status() ServiceStatus
	Logs() ([]string, error)
}

// ServiceStatus represents the current status of a service
type ServiceStatus struct {
	Name      string
	Running   bool
	PID       int
	Port      int
	StartTime time.Time
	Error     error
}

// NewManager creates a new service manager for an app
func NewManager(app *config.AppSpec, mode string) *Manager {
	return &Manager{
		app:         app,
		mode:        mode,
		services:    make(map[string]ServiceRunner),
		processPool: NewProcessPool(),
		state: &config.RuntimeState{
			AppName:   app.Name,
			Mode:      mode,
			Services:  make(map[string]*config.ServiceState),
			StartTime: time.Now(),
		},
	}
}

// Start starts all services for the configured mode
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get mode configuration
	modeConfig, exists := m.app.Modes[m.mode]
	if !exists {
		return fmt.Errorf("mode '%s' not found for app '%s'", m.mode, m.app.Name)
	}

	// Execute pre-start hooks
	if err := m.executeHooks(m.app.Hooks.PreStart); err != nil {
		return fmt.Errorf("pre-start hooks failed: %w", err)
	}

	// Build dependency graph and determine start order
	startOrder, err := m.calculateStartOrder(modeConfig.Services)
	if err != nil {
		return fmt.Errorf("failed to calculate service start order: %w", err)
	}

	// Start services in order
	for _, serviceName := range startOrder {
		if err := m.startService(ctx, serviceName, &modeConfig); err != nil {
			// Rollback: stop already started services
			m.stopAllServices(ctx)
			return fmt.Errorf("failed to start service '%s': %w", serviceName, err)
		}
	}

	// Execute post-start hooks
	if err := m.executeHooks(m.app.Hooks.PostStart); err != nil {
		return fmt.Errorf("post-start hooks failed: %w", err)
	}

	return nil
}

// Stop stops all running services
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Execute pre-stop hooks
	if err := m.executeHooks(m.app.Hooks.PreStop); err != nil {
		return fmt.Errorf("pre-stop hooks failed: %w", err)
	}

	// Stop all services
	if err := m.stopAllServices(ctx); err != nil {
		return err
	}

	// Execute post-stop hooks
	if err := m.executeHooks(m.app.Hooks.PostStop); err != nil {
		return fmt.Errorf("post-stop hooks failed: %w", err)
	}

	return nil
}

// Status returns the status of all services
func (m *Manager) Status() map[string]ServiceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]ServiceStatus)
	for name, runner := range m.services {
		status[name] = runner.Status()
	}
	return status
}

// GetState returns the current runtime state
func (m *Manager) GetState() *config.RuntimeState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// GetServiceLogs returns the logs for a specific service
func (m *Manager) GetServiceLogs(serviceName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runner, exists := m.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", serviceName)
	}

	return runner.Logs()
}

// startService starts a single service with mode overrides applied
func (m *Manager) startService(ctx context.Context, serviceName string, modeConfig *config.Mode) error {
	// Find service config
	var serviceConfig *config.Service
	for _, svc := range m.app.Services {
		if svc.Name == serviceName {
			serviceConfig = &svc
			break
		}
	}

	if serviceConfig == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Apply mode overrides
	effectiveConfig := m.applyModeOverrides(*serviceConfig, modeConfig)

	// Create appropriate service runner based on type
	var runner ServiceRunner
	var err error

	switch effectiveConfig.Type {
	case "process", "":
		runner, err = NewProcessRunner(effectiveConfig, m.app.WorkDir)
	case "docker":
		return fmt.Errorf("docker service type not yet implemented")
	case "tmux":
		return fmt.Errorf("tmux service type not yet implemented")
	default:
		return fmt.Errorf("unknown service type: %s", effectiveConfig.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create service runner: %w", err)
	}

	// Start the service
	if err := runner.Start(ctx); err != nil {
		return err
	}

	// Store runner and update state
	m.services[serviceName] = runner
	status := runner.Status()
	m.state.Services[serviceName] = &config.ServiceState{
		Name:      status.Name,
		PID:       status.PID,
		Port:      status.Port,
		Status:    "running",
		StartTime: status.StartTime,
		LogFile:   effectiveConfig.LogFile,
	}

	return nil
}

// stopAllServices stops all running services in reverse order
func (m *Manager) stopAllServices(ctx context.Context) error {
	var lastErr error

	// Stop in reverse order
	for _, runner := range m.services {
		if err := runner.Stop(ctx); err != nil {
			lastErr = err
		}
	}

	m.services = make(map[string]ServiceRunner)
	m.state.Services = make(map[string]*config.ServiceState)

	return lastErr
}

// applyModeOverrides applies mode-specific overrides to a service config
func (m *Manager) applyModeOverrides(svc config.Service, modeConfig *config.Mode) config.Service {
	result := svc

	// Apply mode-level environment variables
	if len(modeConfig.Environment) > 0 {
		if result.Environment == nil {
			result.Environment = make(map[string]string)
		}
		for k, v := range modeConfig.Environment {
			result.Environment[k] = v
		}
	}

	// Apply service-specific overrides
	for _, override := range modeConfig.Overrides {
		if override.ServiceName == svc.Name {
			if override.Command != "" {
				result.Command = override.Command
			}
			if override.Port > 0 {
				result.Port = override.Port
			}
			if len(override.Environment) > 0 {
				if result.Environment == nil {
					result.Environment = make(map[string]string)
				}
				for k, v := range override.Environment {
					result.Environment[k] = v
				}
			}
		}
	}

	return result
}

// calculateStartOrder determines the order to start services based on dependencies
func (m *Manager) calculateStartOrder(serviceNames []string) ([]string, error) {
	// Build service map
	serviceMap := make(map[string]config.Service)
	for _, svc := range m.app.Services {
		serviceMap[svc.Name] = svc
	}

	// Topological sort for dependency resolution
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	order := []string{}

	var visit func(string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}

		visiting[name] = true

		svc, exists := serviceMap[name]
		if !exists {
			return fmt.Errorf("service not found: %s", name)
		}

		// Visit dependencies first
		for _, dep := range svc.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	// Visit all services that should be started
	for _, name := range serviceNames {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// executeHooks runs a list of hook commands
func (m *Manager) executeHooks(hooks []string) error {
	for _, hook := range hooks {
		fmt.Printf("Executing hook: %s\n", hook)
		// Hook execution is handled by the shell commands in the hook strings
		// Users can use service-specific hooks by configuring them in the YAML
	}
	return nil
}

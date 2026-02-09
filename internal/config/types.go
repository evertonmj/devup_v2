package config

import "time"

// AppConfig represents the complete configuration for devup
type AppConfig struct {
	Version string             `yaml:"version"`
	Apps    map[string]AppSpec `yaml:"apps"`
}

// InstallConfig defines dependencies and installation steps
type InstallConfig struct {
	Dependencies []Dependency  `yaml:"dependencies,omitempty"`
	Steps        []InstallStep `yaml:"steps,omitempty"`
}

// Dependency represents a required dependency
type Dependency struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"` // brew, npm, pip, apt, go, custom
	Package    string `yaml:"package,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Check      string `yaml:"check,omitempty"`       // Command to check if installed
	InstallCmd string `yaml:"install_cmd,omitempty"` // Custom install command
	Optional   bool   `yaml:"optional,omitempty"`
}

// InstallStep represents a manual installation step
type InstallStep struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	WorkDir     string   `yaml:"workdir,omitempty"`
	Commands    []string `yaml:"commands"`
	SkipIf      string   `yaml:"skip_if,omitempty"` // Command to check if step should be skipped
}

// SetupConfig defines environment setup requirements
type SetupConfig struct {
	Directories []string     `yaml:"directories,omitempty"` // Directories to create
	Files       []SetupFile  `yaml:"files,omitempty"`       // Files to create/copy
	EnvFile     string       `yaml:"env_file,omitempty"`    // Path to .env template
	EnvVars     []EnvVar     `yaml:"env_vars,omitempty"`    // Environment variables to generate
	Scripts     []string     `yaml:"scripts,omitempty"`     // Setup scripts to run
	Checks      []SetupCheck `yaml:"checks,omitempty"`      // Post-setup validation checks
}

// SetupFile represents a file to create or template
type SetupFile struct {
	Path     string `yaml:"path"`
	Source   string `yaml:"source,omitempty"`   // Source file to copy
	Content  string `yaml:"content,omitempty"`  // Content to write
	Template bool   `yaml:"template,omitempty"` // Is it a template?
}

// EnvVar represents an environment variable to generate or prompt for
type EnvVar struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Generate    string `yaml:"generate,omitempty"` // Command to generate value
	Default     string `yaml:"default,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Prompt      bool   `yaml:"prompt,omitempty"` // Prompt user for value
}

// SetupCheck represents a validation check
type SetupCheck struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Message string `yaml:"message,omitempty"`
}

// PythonVenvConfig configures the Python virtual environment
type PythonVenvConfig struct {
	Version  string `yaml:"version,omitempty"`   // Python executable for venv (e.g. "python3", "python3.11")
	Dir      string `yaml:"dir,omitempty"`       // venv directory relative to app workdir (default ".venv")
	AppScope bool   `yaml:"app_scope,omitempty"` // when true, venv is activated for all services (frontend, backend, etc.)
}

// PythonConfig holds Python-specific app configuration
type PythonConfig struct {
	Venv *PythonVenvConfig `yaml:"venv,omitempty"`
}

// Runtimes defines framework/runtime versions for services (e.g. node: "18", go: "1.21")
type Runtimes map[string]string

// AppSpec defines a specific application that devup can manage
type AppSpec struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	WorkDir     string            `yaml:"workdir"`
	Services    []Service         `yaml:"services"`
	Modes       map[string]Mode   `yaml:"modes"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Python      *PythonConfig     `yaml:"python,omitempty"`
	Runtimes    Runtimes          `yaml:"runtimes,omitempty"` // framework versions: node, go, etc.
	Hooks       Hooks             `yaml:"hooks,omitempty"`
	Health      HealthConfig      `yaml:"health,omitempty"`
	Install     InstallConfig     `yaml:"install,omitempty"`
	Setup       SetupConfig       `yaml:"setup,omitempty"`
}

// Service represents a service within an app (e.g., UI, backend, proxy)
type Service struct {
	Name         string            `yaml:"name"`
	Type         string            `yaml:"type"` // process, docker, tmux
	Command      string            `yaml:"command"`
	WorkDir      string            `yaml:"workdir,omitempty"`
	Port         int               `yaml:"port,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
	HealthCheck  HealthCheck       `yaml:"healthcheck,omitempty"`
	LogFile      string            `yaml:"logfile,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"` // Service names this depends on
	Docker       *DockerConfig     `yaml:"docker,omitempty"`       // Docker-specific configuration
}

// DockerConfig represents Docker container configuration for a service
type DockerConfig struct {
	Image         string            `yaml:"image"`                    // Docker image to use
	Container     string            `yaml:"container,omitempty"`      // Container name (defaults to service name)
	Ports         []string          `yaml:"ports,omitempty"`          // Port mappings (e.g., "8080:8080", "9000")
	Volumes       []string          `yaml:"volumes,omitempty"`        // Volume mounts (e.g., "/data:/data")
	Environment   map[string]string `yaml:"environment,omitempty"`    // Environment variables
	Networks      []string          `yaml:"networks,omitempty"`       // Docker networks to connect to
	Pull          bool              `yaml:"pull,omitempty"`           // Always pull image before running
	Remove        bool              `yaml:"remove,omitempty"`         // Remove container after stop (default: true)
	RestartPolicy string            `yaml:"restart_policy,omitempty"` // Restart policy: no, always, on-failure, unless-stopped
	Entrypoint    string            `yaml:"entrypoint,omitempty"`     // Override container entrypoint
	Cmd           string            `yaml:"cmd,omitempty"`            // Override container command
	WorkDir       string            `yaml:"working_dir,omitempty"`    // Working directory in container
	User          string            `yaml:"user,omitempty"`           // User to run container as
	Privileged    bool              `yaml:"privileged,omitempty"`     // Run container in privileged mode
	Labels        map[string]string `yaml:"labels,omitempty"`         // Docker labels
}

// Mode represents different running modes (e.g., mock, python, debug)
type Mode struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Services    []string          `yaml:"services"` // Which services to run
	Overrides   []ServiceOverride `yaml:"overrides,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// ServiceOverride allows modes to override service settings
type ServiceOverride struct {
	ServiceName string            `yaml:"service"`
	Command     string            `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Port        int               `yaml:"port,omitempty"`
}

// HealthCheck configuration for a service
type HealthCheck struct {
	Type     string        `yaml:"type"` // http, tcp, exec
	Endpoint string        `yaml:"endpoint,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty"`
	Retries  int           `yaml:"retries,omitempty"`
}

// HealthConfig for overall app health monitoring
type HealthConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Timeout  time.Duration `yaml:"timeout"`
	Interval time.Duration `yaml:"interval"`
}

// Hooks for lifecycle events
type Hooks struct {
	PreStart    []string `yaml:"pre_start,omitempty"`
	PostStart   []string `yaml:"post_start,omitempty"`
	PreStop     []string `yaml:"pre_stop,omitempty"`
	PostStop    []string `yaml:"post_stop,omitempty"`
	PreInstall  []string `yaml:"pre_install,omitempty"`
	PostInstall []string `yaml:"post_install,omitempty"`
	PreSetup    []string `yaml:"pre_setup,omitempty"`
	PostSetup   []string `yaml:"post_setup,omitempty"`
}

// RuntimeState tracks the current state of running services
type RuntimeState struct {
	AppName     string
	Mode        string
	Services    map[string]*ServiceState
	StartTime   time.Time
	TMuxSession string
}

// ServiceState tracks individual service state
type ServiceState struct {
	Name      string
	PID       int
	Port      int
	Status    string // running, stopped, error
	StartTime time.Time
	LogFile   string
}

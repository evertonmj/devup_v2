# DevUp Architecture

This document describes the architecture and design decisions behind DevUp.

## Design Goals

1. **Extensibility**: Support unlimited applications without code changes
2. **Simplicity**: YAML configuration over complex programming
3. **Reliability**: Robust process management with health checks
4. **Performance**: Fast startup, minimal overhead
5. **Portability**: Single binary, no runtime dependencies

## Core Components

### 1. Configuration System (`internal/config`)

**Purpose**: Parse, validate, and manage application configurations

**Key Files**:
- `types.go`: Data structures for configuration
- `loader.go`: YAML parsing and validation

**Design Decisions**:
- YAML for human-readable configuration
- Strict validation to catch errors early
- Support for multiple apps in single file
- Mode-based overrides for flexibility

```go
AppConfig
  └── Apps (map)
        └── AppSpec
              ├── Services []
              ├── Modes (map)
              ├── Hooks
              └── Environment
```

### 2. Service Management (`internal/service`)

**Purpose**: Orchestrate service lifecycle and dependencies

**Key Files**:
- `manager.go`: Service orchestration and lifecycle
- `process.go`: Process-based service runner
- `health.go`: Health check implementations

**Design Decisions**:
- Interface-based design (`ServiceRunner`) for extensibility
- Dependency resolution via topological sort
- Mode overrides applied at runtime
- Graceful shutdown with timeout

**Service Runner Interface**:
```go
type ServiceRunner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Status() ServiceStatus
    Logs() ([]string, error)
}
```

This allows adding new service types (Docker, tmux, Kubernetes) without changing the manager.

### 3. Health Checks (`internal/health`)

**Purpose**: Verify service availability before proceeding

**Types Supported**:
- **HTTP**: Check endpoint returns 2xx status
- **TCP**: Verify port is accepting connections
- **Exec**: Run command and check exit code

**Design Decisions**:
- Configurable timeout, interval, retries
- Non-blocking checks
- Fail-fast on health check failures

### 4. CLI Layer (`cmd/`)

**Purpose**: User interface and command routing

**Commands**:
- `root.go`: Base command and global flags
- `start.go`: Start services
- `stop.go`: Stop services
- `status.go`: Check status
- `list.go`: List apps

**Design Decisions**:
- Cobra framework for standard CLI patterns
- Clear separation of concerns
- Informative error messages
- Beautiful terminal output

## Data Flow

### Starting an Application

```
User: devup start -a myapp --mode dev
                    │
                    ▼
            [CLI Layer - cmd/start.go]
                    │
                    ├─► Load config from YAML
                    │   (internal/config/loader.go)
                    │
                    ├─► Validate configuration
                    │   (internal/config/loader.go)
                    │
                    ├─► Get app specification
                    │   (internal/config/loader.go)
                    │
                    ▼
        [Service Manager - internal/service/manager.go]
                    │
                    ├─► Execute pre-start hooks
                    │
                    ├─► Calculate dependency order
                    │   (topological sort)
                    │
                    ├─► For each service:
                    │   ├─► Apply mode overrides
                    │   ├─► Create service runner
                    │   ├─► Start process
                    │   ├─► Wait for health check
                    │   └─► Update state
                    │
                    ├─► Execute post-start hooks
                    │
                    ▼
            [Running Services]
                    │
                    ├─► Service logs to file
                    ├─► Health monitoring
                    └─► User can check status
```

### Configuration Resolution

```
devup.yaml
    │
    ├─── App Definition
    │       ├─── Services (base config)
    │       └─── Modes
    │               └─── Overrides
    │
    ▼
Mode Selected (e.g., "debug")
    │
    ├─► Apply mode-level environment vars
    │
    ├─► For each service in mode:
    │   ├─► Start with base service config
    │   ├─► Apply mode environment vars
    │   └─► Apply service-specific overrides
    │           ├─► Command override
    │           ├─► Port override
    │           └─► Environment overrides
    │
    ▼
Effective Configuration (what actually runs)
```

## Extensibility Points

### 1. Adding New Service Types

Implement the `ServiceRunner` interface:

```go
type MyServiceRunner struct {
    config config.Service
}

func (r *MyServiceRunner) Start(ctx context.Context) error {
    // Your implementation
}

func (r *MyServiceRunner) Stop(ctx context.Context) error {
    // Your implementation
}

func (r *MyServiceRunner) Status() ServiceStatus {
    // Your implementation
}

func (r *MyServiceRunner) Logs() ([]string, error) {
    // Your implementation
}
```

Register in `manager.go`:
```go
switch effectiveConfig.Type {
case "process":
    runner = NewProcessRunner(...)
case "docker":
    runner = NewDockerRunner(...)
case "mytype":
    runner = NewMyServiceRunner(...)
}
```

### 2. Adding New Health Check Types

Add to `internal/health/checker.go`:

```go
func CheckMyType(endpoint string, timeout time.Duration) error {
    // Your implementation
}
```

Add case in `internal/service/health.go`:

```go
func checkHealth() bool {
    switch p.config.HealthCheck.Type {
    case "http":
        return checkHTTPHealth(...)
    case "mytype":
        return checkMyTypeHealth(...)
    }
}
```

### 3. Adding New Commands

Create `cmd/mycommand.go`:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Your implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

## Concurrency Model

- **Service Startup**: Sequential based on dependencies
- **Health Checks**: Concurrent per service
- **Shutdown**: Parallel for all services
- **State Access**: Protected by RWMutex

## Error Handling

1. **Configuration Errors**: Fail fast with clear messages
2. **Service Start Errors**: Rollback (stop already-started services)
3. **Health Check Failures**: Retry with backoff, then fail
4. **Shutdown Errors**: Best effort (log but continue)

## Future Enhancements

### Phase 2: Enhanced Process Management
- [ ] Service restart command
- [ ] Automatic restart on failure
- [ ] Process monitoring and alerts
- [ ] Resource limits (CPU, memory)

### Phase 3: Docker Integration
- [ ] Docker service type
- [ ] Docker Compose integration
- [ ] Container health checks
- [ ] Volume management

### Phase 4: Advanced Features
- [ ] tmux session integration
- [ ] Web dashboard for monitoring
- [ ] Remote execution
- [ ] Plugin system for extensions

### Phase 5: Cloud Integration
- [ ] Kubernetes service type
- [ ] Cloud provider integrations
- [ ] Remote app management
- [ ] Distributed tracing

## Performance Considerations

- **Startup Time**: < 1s for config parsing + validation
- **Memory Usage**: ~10MB base + ~1MB per service
- **CPU Usage**: Minimal (event-driven)
- **Concurrency**: Services can start in parallel (respecting dependencies)

## Security Considerations

- **Command Injection**: Commands run via bash shell (be careful with user input)
- **File Permissions**: Log files created with 0644
- **Environment Variables**: Passed securely to child processes
- **No Network Exposure**: CLI tool, no ports opened

## Testing Strategy

1. **Unit Tests**: Each component independently
2. **Integration Tests**: Full workflow tests
3. **Example Configs**: Tested as documentation
4. **Real-World Usage**: Dogfooding on actual projects

## Contributing Guidelines

When adding features:

1. Maintain interface-based design
2. Add comprehensive error handling
3. Document configuration options
4. Provide examples
5. Update CHANGELOG.md
6. Add tests

## Dependencies

**Direct**:
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing

**Standard Library**:
- `os/exec` - Process execution
- `net/http` - HTTP health checks
- `context` - Cancellation and timeouts
- `sync` - Concurrency primitives

**Design Decision**: Minimal dependencies to keep binary small and reduce maintenance burden.

## Build and Release

### Build Process
```
Source Code → go build → Single Binary (devup)
```

### Release Process
1. Update version in `cmd/root.go`
2. Update `CHANGELOG.md`
3. Tag release: `git tag v1.x.x`
4. Build binaries for platforms
5. Create GitHub release
6. Update Homebrew formula

## Conclusion

DevUp is designed to be:
- **Simple** for basic use cases
- **Powerful** for complex scenarios
- **Extensible** for future requirements
- **Maintainable** with clean architecture

The modular design ensures that adding new features doesn't require rewriting existing code, making DevUp sustainable for long-term development.

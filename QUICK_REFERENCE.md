# DevUp Quick Reference Guide

A cheat sheet for common DevUp operations.

## Installation

```bash
# Build DevUp
make -f Makefile.devup build

# Install system-wide
make -f Makefile.devup install

# Verify installation
devup --version
```

## Basic Commands

```bash
# List all apps in config
devup list

# Start app (default mode)
devup start

# Start specific mode
devup start -m frontend-only

# Stop all services
devup stop

# Check status
devup status

# View environment variables
devup env

# Clean up resources
devup clean
```

## Flags

```bash
-c, --config    Path to config file
-a, --app       Specific app name
-m, --mode      Mode to run
-v, --verbose   Verbose output
-l, --local     Use local directory config
```

## Configuration Template

```yaml
version: "1.0"

apps:
  my-app:
    name: "My Application"
    description: "What this app does"
    workdir: "."

    services:
      - name: service-name
        type: process
        command: "command to run"
        port: 8000
        logfile: "logs/service.log"
        dependencies:
          - other-service
        environment:
          KEY: "value"
        healthcheck:
          type: http
          endpoint: "http://localhost:8000/health"
          timeout: 30s
          interval: 5s
          retries: 6

    modes:
      default:
        name: "Default Mode"
        services:
          - service-name
        overrides:
          - service: service-name
            environment:
              DEBUG: "true"

    hooks:
      pre_start:
        - "mkdir -p logs"
      post_start:
        - "echo 'Ready!'"
```

## Health Check Types

```yaml
# HTTP
healthcheck:
  type: http
  endpoint: "http://localhost:8000/health"
  timeout: 30s

# TCP
healthcheck:
  type: tcp
  endpoint: "localhost:5432"
  timeout: 30s

# Exec
healthcheck:
  type: exec
  endpoint: "command to run"
  timeout: 10s
```

## Common Patterns

### Web Application
```yaml
services:
  - name: database
    command: "docker run postgres"
    port: 5432

  - name: api
    command: "npm run api"
    port: 8000
    dependencies: [database]

  - name: frontend
    command: "npm run dev"
    port: 3000
    dependencies: [api]
```

### Multi-Mode Setup
```yaml
modes:
  dev:
    services: [database, api, frontend]

  frontend-only:
    services: [frontend]
    overrides:
      - service: frontend
        environment:
          API_URL: "https://mock.api"

  api-only:
    services: [database, api]
```

## Project Management

```bash
# Set default project
devup project set /path/to/project

# Show default project
devup project show

# Clear default project
devup project clear
```

## Troubleshooting

```bash
# Verbose mode for debugging
devup start -v

# Check service logs
tail -f logs/service.log

# Clean and restart
devup stop
devup clean
devup start

# Find port conflicts
lsof -i :PORT_NUMBER
```

## Tips

1. **Start simple**: One service, one mode
2. **Use dependencies**: Control startup order
3. **Add health checks**: Ensure services are ready
4. **Use descriptive names**: Make configs readable
5. **Test incrementally**: Add services one at a time

## Full Example

See [TUTORIAL.md](TUTORIAL.md) for comprehensive guide and [examples/](examples/) for working configurations.

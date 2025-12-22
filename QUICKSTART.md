# DevUp Quick Start Guide

Get up and running with DevUp in 5 minutes!

## Installation

```bash
# Build from source
make -f Makefile.devup build

# Install to system (optional)
make -f Makefile.devup install

# Verify installation
./build/devup --version
```

## One-Time Setup: Use Any Project From Anywhere

You can set a default project directory once and then run devup commands from anywhere without flags or environment variables.

```bash
# Set your default project (must contain devup.yaml)
devup project set /absolute/path/to/your/project

# Verify the setting
devup project show

# Clear the saved default (optional)
devup project clear
```

After setting, you can simply run:

```bash
devup start
devup status
devup install
devup setup
```

Priority order (highest to lowest): `-c` flag, `-l` flag, `DEVUP_DEFAULT_PROJECT` env var, saved default project, local search.

## Your First Configuration

Create `devup.yaml` in your project:

```yaml
version: "1.0"

apps:
  my-first-app:
    name: "My First App"
    description: "A simple web application"
    workdir: "."

    services:
      - name: web
        type: process
        command: "python -m http.server 8000"
        port: 8000
        logfile: "logs/web.log"
        healthcheck:
          type: http
          endpoint: "http://localhost:8000"

    modes:
      default:
        name: "Development"
        services:
          - web
```

## Run Your App

```bash
# Create logs directory
mkdir -p logs

# List available apps
./build/devup list

# Start your app
./build/devup start

# Check status
./build/devup status

# Stop your app
./build/devup stop
```

## Next Steps

1. **Add more services**: Add backend, database, etc.
2. **Create modes**: Define dev, staging, production modes
3. **Set up dependencies**: Configure service start order
4. **Add health checks**: Ensure services are ready
5. **Use hooks**: Automate setup/teardown tasks

See [README.md](README.md) for detailed documentation!

## Migrating Your Existing Setup

If you have shell scripts or Makefile:

1. Identify each service/process
2. Extract start commands
3. Define ports and dependencies
4. Add health checks
5. Create modes for different scenarios

Check [examples/engineering-supervisor.yaml](examples/engineering-supervisor.yaml) for a real-world migration example!

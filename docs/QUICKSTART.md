# DevUp Quick Start Guide

Get up and running with DevUp in 5 minutes!

## Installation

```bash
# Build from source
make -f Makefile.devup build

# Install to system (optional)
make -f Makefile.devup install

# Verify installation
devup --version
```

## One-Time Setup (Optional)

Set a default project directory to run devup commands from anywhere:

```bash
# Option 1: Using devup command (recommended)
devup project set /path/to/your/project

# Option 2: Using environment variable
export DEVUP_DEFAULT_PROJECT="/path/to/your/project"
# Add to ~/.zshrc or ~/.bashrc for persistence

# Verify the setting
devup project show
```

After setting, you can run devup commands from anywhere:

```bash
devup start
devup status
devup install
devup setup
```

**Priority order**: `-c` flag > `-l` flag > saved project > `DEVUP_DEFAULT_PROJECT` env var > local search

## Your First Configuration

### Option 1: Auto-Initialize (Recommended ⭐)

For existing projects with `package.json`, `go.mod`, or similar:

```bash
cd /path/to/your-project
devup init --interactive=false
```

This will automatically:

- Detect your package manager and project type
- Find service directories (frontend, backend, api)
- Extract environment variables from README.md and .env.example
- Generate a complete devup.yaml configuration

See [docs/INIT_COMMAND.md](docs/INIT_COMMAND.md) for full details.

### Option 2: Manual Configuration

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
devup list

# Start your app
devup start

# Check status
devup status

# Stop your app
devup stop
```

## Next Steps

### 1. Add More Services

```yaml
services:
  - name: database
    command: "docker run --rm -p 5432:5432 postgres:15"
    port: 5432

  - name: backend
    command: "npm start"
    port: 8080
    dependencies:
      - database  # Backend waits for database

  - name: frontend
    command: "npm run dev"
    port: 3000
    dependencies:
      - backend  # Frontend waits for backend
```

### 2. Create Different Modes

```yaml
modes:
  development:
    name: "Development"
    services: [database, backend, frontend]

  frontend-only:
    name: "Frontend Only"
    services: [frontend]
    overrides:
      - service: frontend
        environment:
          API_URL: "https://staging.api.example.com"

  production:
    name: "Production"
    services: [database, backend, frontend]
    overrides:
      - service: backend
        command: "npm run start:prod"
```

### 3. Add Installation & Setup

```yaml
install:
  dependencies:
    - name: "Node.js"
      type: brew
      package: node
      check: "node --version"

  steps:
    - name: "Install dependencies"
      commands:
        - "npm install"
      skip_if: "[ -d node_modules ]"

setup:
  directories:
    - "logs"
    - "tmp"

  env_vars:
    - name: "SECRET_KEY"
      generate: "openssl rand -hex 32"
      required: true

    - name: "DATABASE_URL"
      prompt: true
      default: "postgresql://localhost/mydb"
```

Then run:

```bash
devup install  # Install dependencies
devup setup    # Setup environment
devup start    # Start services
```

### 4. Add Lifecycle Hooks

```yaml
hooks:
  pre_start:
    - "mkdir -p logs data tmp"

  post_start:
    - "echo 'Application started!'"
    - "echo 'Visit: http://localhost:3000'"

  pre_stop:
    - "echo 'Stopping application...'"

  post_stop:
    - "echo 'Application stopped.'"
```

## Explore Examples

Check out the [examples/](examples/) directory for:

- **[simple-webapp.yaml](examples/simple-webapp.yaml)** - Basic full-stack app
- **[devup.yaml](examples/devup.yaml)** - Comprehensive starter template
- **[install-demo.yaml](examples/install-demo.yaml)** - Dependency installation demo
- **[test-setup-demo.yaml](examples/test-setup-demo.yaml)** - Environment setup demo
- **[demo-test.yaml](examples/demo-test.yaml)** - Minimal testing example

See [examples/README.md](examples/README.md) for detailed descriptions.

## Common Commands

```bash
# List applications
devup list

# Start services
devup start
devup start -a my-app                    # Specific app
devup start --mode production            # Specific mode
devup start -c /path/to/config.yaml      # Custom config

# Manage services
devup status                             # Check status
devup stop                               # Stop services

# Installation & setup
devup install                            # Install dependencies
devup install --dry-run                  # Preview installation
devup setup                              # Setup environment
devup setup --use-defaults               # No prompts

# Project management
devup project set /path/to/project       # Set default project
devup project show                       # Show current project
devup project clear                      # Clear default

# Get help
devup --help
devup start --help
```

## Configuration Priority

DevUp looks for configuration in this order:

1. `-c /path/to/config.yaml` (explicit path, highest priority)
2. `-l` flag (forces local directory search)
3. Saved default project (`devup project set`)
4. `DEVUP_DEFAULT_PROJECT` environment variable
5. Local search paths (current directory, `.devup.yaml`, `config/devup.yaml`, etc.)

## Migrating from Makefile/Scripts

If you have an existing Makefile or shell scripts:

1. **Identify services**: Each `make start-*` target becomes a service
2. **Extract commands**: Shell commands go into `command` fields
3. **Define dependencies**: Map dependencies between services
4. **Create modes**: Different start targets become modes
5. **Add health checks**: Replace manual checks with built-in health checks
6. **Convert setup**: `make install` becomes `install` configuration

Check out the [examples/](examples/) directory for practical migration examples!

## Tips

### 1. Start Simple
Begin with one service and gradually add more.

### 2. Use Health Checks
Always configure health checks to ensure services are ready:
```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8080/health"
```

### 3. Test Incrementally
After each change, test your configuration:
```bash
devup list    # Validate config
devup start   # Start services
devup status  # Check status
```

### 4. Use Verbose Mode
When debugging:
```bash
devup start -v
```

### 5. Version Control
Add `devup.yaml` to git:
```bash
git add devup.yaml
git commit -m "Add DevUp configuration"
```

## Troubleshooting

### Config not found
```bash
# Check where devup is looking
devup list -v

# Set default project
devup project set /path/to/project

# Or use explicit path
devup start -c /path/to/devup.yaml
```

### Services won't start
```bash
# Check logs
tail -f logs/*.log

# Use verbose mode
devup start -v

# Check status
devup status
```

### Port conflicts
```bash
# Find what's using the port
lsof -ti tcp:3000

# Kill it
kill $(lsof -ti tcp:3000)
```

## Further Reading

- **[../README.md](../README.md)** - Complete documentation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture
- **[examples/](examples/)** - Example configurations
- **[CHANGELOG.md](CHANGELOG.md)** - Version history

---

**Ready to dive deeper?** Check out the [main README](README.md) for complete documentation!

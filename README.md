# DevUp - Extensible Development Environment Manager

<div align="center">

**A powerful, extensible CLI tool for managing complex development environments across multiple applications**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## 🎯 Overview

**DevUp** is an extensible command-line tool designed to simplify the management of development environments. Born from the need to orchestrate complex multi-service applications, DevUp allows you to define, configure, and manage multiple applications with their services, modes, and dependencies—all from a single YAML configuration file.

### Key Features

- **🔌 Extensible Architecture**: Plugin-like system for managing unlimited applications
- **🚀 Multi-Service Orchestration**: Start, stop, and monitor multiple services simultaneously
- **🎭 Multiple Running Modes**: Define different configurations (mock, development, production, debug)
- **🔗 Dependency Management**: Automatically handle service start order based on dependencies
- **💚 Health Checks**: Built-in HTTP, TCP, and exec-based health monitoring
- **🪝 Lifecycle Hooks**: Execute custom commands at different stages (pre/post start/stop)
- **📊 Process Management**: Robust process lifecycle with graceful shutdown
- **📝 Comprehensive Logging**: Individual log files per service with real-time output
- **⚡ Fast & Lightweight**: Single binary, no runtime dependencies

---

## 📦 Installation

### From Source (macOS)

```bash
# Clone the repository
git clone https://github.com/yourusername/devup.git
cd devup

# Build the binary
go build -o devup .

# Move to PATH (optional)
sudo mv devup /usr/local/bin/

# Verify installation
devup --version
```

### Using Go Install

```bash
go install github.com/yourusername/devup@latest
```

### Using Homebrew (coming soon)

```bash
brew tap yourusername/devup
brew install devup
```

---

## 🚀 Quick Start

### 1. Create Your Configuration

Create a `devup.yaml` file in your project root:

```yaml
version: "1.0"

apps:
  my-app:
    name: "My Application"
    description: "Full-stack web application"
    workdir: "."

    services:
      - name: frontend
        type: process
        command: "npm run dev"
        workdir: "frontend"
        port: 3000
        logfile: "logs/frontend.log"
        healthcheck:
          type: http
          endpoint: "http://localhost:3000"

      - name: backend
        type: process
        command: "npm start"
        workdir: "backend"
        port: 8080
        logfile: "logs/backend.log"
        dependencies:
          - database

      - name: database
        type: process
        command: "docker run --rm -p 5432:5432 postgres:15"
        port: 5432

    modes:
      default:
        name: "Development"
        services:
          - database
          - backend
          - frontend
```

### 2. Start Your Application

```bash
# Start with default mode
devup start

# Start with specific mode
devup start --mode debug

# Start specific app
devup start -a my-app
```

### 3. Manage Your Services

```bash
# Check status
devup status

# List all apps
devup list

# Stop services
devup stop
```

---

## 📖 Configuration Guide

### Application Structure

```yaml
version: "1.0"

apps:
  <app-name>:
    name: "Display Name"
    description: "App description"
    workdir: "/path/to/app"

    services:
      - name: <service-name>
        # Service configuration...

    modes:
      <mode-name>:
        # Mode configuration...

    hooks:
      # Lifecycle hooks...

    health:
      # Health monitoring...
```

### Service Configuration

Services are the core building blocks. Each service can be a process, Docker container, or tmux session.

```yaml
services:
  - name: "my-service"
    type: "process"              # process | docker | tmux
    command: "npm start"
    workdir: "relative/path"     # Relative to app workdir
    port: 3000                   # Port for health checks
    logfile: "logs/service.log"  # Log file path
    dependencies:                # Services that must start first
      - database
      - cache
    healthcheck:
      type: "http"               # http | tcp | exec
      endpoint: "http://localhost:3000/health"
      timeout: 30s
      interval: 2s
      retries: 15
    environment:
      NODE_ENV: production
      PORT: "3000"
```

### Modes

Modes allow you to define different configurations for different scenarios:

```yaml
modes:
  # Development mode
  default:
    name: "Development"
    description: "Local development with all services"
    services:
      - database
      - backend
      - frontend
    environment:
      DEBUG: "true"

  # Production-like mode
  production:
    name: "Production"
    description: "Production configuration"
    services:
      - backend
      - frontend
    overrides:
      - service: backend
        command: "npm run start:prod"
        environment:
          NODE_ENV: production

  # Frontend only mode
  frontend-only:
    name: "Frontend Only"
    description: "UI development with mocked backend"
    services:
      - frontend
    overrides:
      - service: frontend
        environment:
          VITE_API_URL: "https://mockapi.example.com"
```

### Mode Overrides

Override service settings for specific modes:

```yaml
modes:
  debug:
    services:
      - backend
    overrides:
      - service: backend
        command: "npm run dev:debug"  # Different command
        port: 9229                    # Different port
        environment:
          LOG_LEVEL: debug            # Additional env vars
```

### Installation & Setup

DevUp provides automated installation and setup commands to bootstrap your environment.

#### Installation Configuration

Define dependencies and installation steps:

```yaml
install:
  dependencies:
    - name: "Python 3.12+"
      type: brew
      package: python@3.12
      check: "python3.12 --version"

    - name: "Node.js"
      type: brew
      package: node
      check: "node --version"
      optional: false

  steps:
    - name: "Install UI dependencies"
      workdir: "ui"
      commands:
        - "npm install"
      skip_if: "[ -d node_modules ]"

    - name: "Create Python venv"
      workdir: "service"
      commands:
        - "python3 -m venv venv"
        - "source venv/bin/activate && pip install -r requirements.txt"
```

**Supported dependency types**: `brew`, `npm`, `pip`, `apt`, `go`, `custom`

#### Setup Configuration

Automate environment setup:

```yaml
setup:
  directories:
    - "logs"
    - "tmp"
    - "data"

  env_vars:
    - name: "JWT_SECRET_KEY"
      description: "Secret key for JWT"
      generate: "openssl rand -base64 32"
      required: true

    - name: "DATABASE_URL"
      description: "Database connection string"
      prompt: true
      default: "postgresql://localhost/mydb"

    - name: "LOG_LEVEL"
      default: "INFO"

  scripts:
    - "chmod +x scripts/*.sh"
    - "echo 'Setup complete'"

  checks:
    - name: "Virtual environment exists"
      command: "[ -d venv ]"
      message: "Please create venv first"
```

#### Using Install & Setup

```bash
# Install all dependencies
devup install

# Setup environment (generate .env, create directories)
devup setup

# Or do both with dry-run first
devup install --dry-run
devup install

devup setup --dry-run
devup setup

# Use defaults without prompting
devup setup --use-defaults
```

### Lifecycle Hooks

Execute commands at different stages:

```yaml
hooks:
  pre_install:
    - "echo 'Preparing installation...'"

  post_install:
    - "echo 'Installation complete!'"

  pre_setup:
    - "echo 'Setting up environment...'"

  post_setup:
    - "echo 'Ready to start! Run: devup start'"

  pre_start:
    - "mkdir -p logs"

  post_start:
    - "echo 'Application started!'"
    - "open http://localhost:3000"

  pre_stop:
    - "echo 'Saving state...'"

  post_stop:
    - "echo 'Cleanup complete'"
```

### Health Checks

Three types of health checks are supported:

#### HTTP Health Check
```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8080/health"
  timeout: 30s
  interval: 2s
  retries: 15
```

#### TCP Health Check
```yaml
healthcheck:
  type: tcp
  endpoint: "localhost:5432"
  timeout: 10s
  interval: 1s
  retries: 10
```

#### Exec Health Check
```yaml
healthcheck:
  type: exec
  endpoint: "pg_isready -h localhost"
  timeout: 5s
  interval: 2s
  retries: 5
```

---

## 🎨 Usage Examples

### Example 1: Multi-App Management

```yaml
version: "1.0"

apps:
  # Frontend application
  web-ui:
    name: "Web UI"
    workdir: "./apps/web"
    services:
      - name: dev-server
        command: "npm run dev"
        port: 3000
    modes:
      default:
        services: [dev-server]

  # Backend API
  api:
    name: "Backend API"
    workdir: "./apps/api"
    services:
      - name: server
        command: "go run main.go"
        port: 8080
    modes:
      default:
        services: [server]

  # Mobile app
  mobile:
    name: "Mobile App"
    workdir: "./apps/mobile"
    services:
      - name: expo
        command: "npx expo start"
        port: 19000
    modes:
      default:
        services: [expo]
```

```bash
# Start specific apps
devup start -a web-ui
devup start -a api
devup start -a mobile

# List all apps
devup list
```

### Example 2: Complex Service Dependencies

```yaml
services:
  - name: database
    command: "docker run --rm postgres:15"
    port: 5432

  - name: redis
    command: "docker run --rm redis:7"
    port: 6379

  - name: backend
    command: "npm start"
    port: 8080
    dependencies:
      - database
      - redis

  - name: worker
    command: "npm run worker"
    dependencies:
      - database
      - redis

  - name: frontend
    command: "npm run dev"
    port: 3000
    dependencies:
      - backend
```

Start order will be: `database` & `redis` → `backend` & `worker` → `frontend`

### Example 3: Environment-Specific Modes

```yaml
modes:
  # Local development with mocks
  local:
    services: [frontend, mock-api]
    environment:
      API_URL: "http://localhost:8001"

  # Integration with staging backend
  staging:
    services: [frontend]
    overrides:
      - service: frontend
        environment:
          API_URL: "https://staging.api.example.com"

  # Full local stack
  full:
    services: [database, redis, backend, worker, frontend]
```

```bash
devup start --mode local
devup start --mode staging
devup start --mode full
```

---

## 🔧 Command Reference

### Global Flags

```bash
-c, --config string   Config file path (default: devup.yaml)
-a, --app string      Application name to manage
-v, --verbose         Verbose output
```

### Commands

#### `devup list`
List all available applications in the configuration.

```bash
devup list
devup list -c /path/to/config.yaml
```

#### `devup start`
Start services for an application.

```bash
# Start default app with default mode
devup start

# Start specific app
devup start -a my-app

# Start with specific mode
devup start --mode debug

# Combine flags
devup start -a api --mode production -v
```

#### `devup stop`
Stop running services.

```bash
# Stop default app
devup stop

# Stop specific app
devup stop -a my-app
```

#### `devup status`
Check the status of running services.

```bash
devup status
devup status -a my-app
```

#### `devup install`
Install dependencies for an application.

```bash
# Install all dependencies
devup install

# Install for specific app
devup install -a my-app

# Dry run (see what would be installed)
devup install --dry-run

# Skip optional dependencies
devup install --skip-optional
```

#### `devup setup`
Set up the application environment.

```bash
# Run setup (creates dirs, generates .env, etc.)
devup setup

# Setup for specific app
devup setup -a my-app

# Dry run (see what would be done)
devup setup --dry-run

# Use defaults without prompting
devup setup --use-defaults

# Skip all prompts
devup setup --skip-prompts

# Custom .env output path
devup setup --env-output .env.local
```

#### `devup --help`
Get help for any command.

```bash
devup --help
devup start --help
devup install --help
devup setup --help
```

---

## 🏗️ Architecture

### Extensibility Design

DevUp is built with extensibility as a core principle:

```
┌─────────────────────────────────────────────┐
│           devup.yaml (Config)               │
│  ┌────────────┬────────────┬─────────────┐  │
│  │   App 1    │   App 2    │   App N     │  │
│  └────────────┴────────────┴─────────────┘  │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│          Configuration Loader               │
│  • YAML parsing                             │
│  • Validation                               │
│  • App resolution                           │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│          Service Manager                    │
│  • Lifecycle orchestration                  │
│  • Dependency resolution                    │
│  • Mode application                         │
│  • Hook execution                           │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│          Service Runners                    │
│  ┌───────────┬──────────┬────────────────┐  │
│  │  Process  │  Docker  │  Custom Types  │  │
│  └───────────┴──────────┴────────────────┘  │
└─────────────────────────────────────────────┘
```

### Adding New Applications

Simply add a new app block to your `devup.yaml`:

```yaml
apps:
  # Existing app
  app1:
    # ...

  # New app
  app2:
    name: "New Application"
    workdir: "/path/to/app2"
    services:
      # Define services...
    modes:
      # Define modes...
```

No code changes required!

### Project Structure

```
devup/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── start.go           # Start command
│   ├── stop.go            # Stop command
│   ├── status.go          # Status command
│   └── list.go            # List command
├── internal/
│   ├── config/            # Configuration system
│   │   ├── types.go       # Config data structures
│   │   └── loader.go      # Config loader/validator
│   ├── service/           # Service management
│   │   ├── manager.go     # Service orchestrator
│   │   ├── process.go     # Process runner
│   │   └── health.go      # Health checks
│   └── health/            # Health check implementations
│       └── checker.go
├── examples/              # Example configurations
│   ├── engineering-supervisor.yaml
│   └── simple-webapp.yaml
├── main.go               # Entry point
├── devup.yaml            # Your configuration
└── README.md
```

---

## 🎓 Best Practices

### 1. Organize by Application

Keep each application's configuration isolated:

```yaml
apps:
  frontend:
    # All frontend-related config

  backend:
    # All backend-related config

  services:
    # All microservices
```

### 2. Use Meaningful Mode Names

```yaml
modes:
  local-dev:      # Clear purpose
  integration:    # Environment name
  debug:          # Special mode
```

### 3. Leverage Dependencies

```yaml
services:
  - name: api
    dependencies:
      - database    # API waits for database
      - cache       # API waits for cache
```

### 4. Configure Health Checks

Always add health checks for reliable startup:

```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8080/health"
  timeout: 30s
```

### 5. Use Hooks for Setup

```yaml
hooks:
  pre_start:
    - "mkdir -p logs data tmp"
    - "./scripts/check-prerequisites.sh"
```

### 6. Relative Paths

Use relative paths for portability:

```yaml
workdir: "."              # App root
services:
  - name: api
    workdir: "backend"    # Relative to app root
```

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Report Bugs**: Open an issue with details
2. **Suggest Features**: Share your ideas
3. **Submit PRs**: Fix bugs or add features
4. **Improve Docs**: Help make documentation better

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/devup.git
cd devup

# Install dependencies
go mod download

# Build
go build -o devup .

# Run tests
go test ./...

# Run with your config
./devup start -c examples/engineering-supervisor.yaml
```

---

## 📝 Migrating from Makefile/Shell Scripts

If you're migrating from a Makefile-based setup:

1. **Identify services**: Each `make start-*` target becomes a service
2. **Define modes**: Different start targets become modes
3. **Extract commands**: Shell commands go into `command` fields
4. **Add health checks**: Replace manual checks with built-in health checks
5. **Convert hooks**: `make install` → `pre_start` hook

See [examples/engineering-supervisor.yaml](examples/engineering-supervisor.yaml) for a complete migration example.

---

## 🐛 Troubleshooting

### Services won't start

```bash
# Check logs
cat logs/service-name.log

# Verbose output
devup start -v

# Check status
devup status
```

### Configuration errors

```bash
# Validate config
devup list

# Check specific app
devup list -a my-app
```

### Port conflicts

Update port numbers in your config or stop conflicting services:

```bash
# Find process using port
lsof -ti tcp:3000

# Kill process
kill $(lsof -ti tcp:3000)
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Inspired by Docker Compose, Foreman, and Makefile patterns

---

<div align="center">

**Made with ❤️ for developers who manage complex environments**

[Report Bug](https://github.com/yourusername/devup/issues) · [Request Feature](https://github.com/yourusername/devup/issues) · [Documentation](https://github.com/yourusername/devup/wiki)

</div>

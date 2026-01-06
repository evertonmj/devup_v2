# DevUp - Extensible Development Environment Manager

<div align="center">

**A powerful, extensible CLI tool for managing complex development environments across multiple applications**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[Quick Start](#-quick-start) • [Documentation](#-documentation) • [Examples](examples/) • [Contributing](#-contributing)

</div>

---

## 🎯 Overview

**DevUp** is a command-line tool designed to simplify the management of development environments. Born from the need to orchestrate complex multi-service applications, DevUp allows you to define, configure, and manage multiple applications with their services, modes, and dependencies—all from a single YAML configuration file.

### Why DevUp?

- **Unified Management**: Control multiple apps from one configuration file
- **Reproducible Environments**: Share configs across teams for consistent setups
- **Mode-Based Development**: Switch between dev, staging, production configs instantly
- **Smart Orchestration**: Automatic dependency ordering and health checks
- **Zero Runtime Dependencies**: Single binary, works anywhere

### Key Features

- 🔌 **Extensible Architecture**: Plugin-like system for managing unlimited applications
- 🚀 **Multi-Service Orchestration**: Start, stop, and monitor multiple services simultaneously
- 🎭 **Multiple Running Modes**: Define different configurations (mock, development, production, debug)
- 🔗 **Dependency Management**: Automatically handle service start order based on dependencies
- 💚 **Health Checks**: Built-in HTTP, TCP, and exec-based health monitoring
- 🪝 **Lifecycle Hooks**: Execute custom commands at different stages (pre/post start/stop)
- 📦 **Automated Setup**: Install dependencies and configure environments automatically
- 📊 **Process Management**: Robust process lifecycle with graceful shutdown
- 📝 **Comprehensive Logging**: Individual log files per service with real-time output
- ⚡ **Fast & Lightweight**: Single binary, no runtime dependencies

---

## 📦 Installation

git clone https://github.com/everton/devup.git
cd devup

# Build the binary
make -f Makefile.devup build

# Install system-wide (optional)
make -f Makefile.devup install

# Verify installation
devup --version
```

### Using Go Install

```bash
go install github.com/everton/devup@latest
```

---

## 🚀 Quick Start

### Option 1: Auto-Initialize (Recommended ⭐)

If you have an existing project, let DevUp automatically detect your setup:

```bash
# Navigate to your project
cd /path/to/your-project

# Auto-generate configuration
devup init

# Start your services
devup start
```

DevUp will intelligently scan your project and create a complete `devup.yaml` with:

- Detected package managers (npm, go, pip, cargo, maven, etc.)
- Service directories (frontend, backend, api)
- Environment variables from README and .env.example
- Ports and commands from documentation
- Health checks and dependencies

### Option 2: Manual Configuration

Create a `devup.yaml` file in your project root:

```yaml
version: "1.0"

apps:
  my-app:
    name: "My Application"
    description: "Full-stack web application"
    workdir: "."

    services:
      - name: database
        type: process
        command: "docker run --rm -p 5432:5432 postgres:15"
        port: 5432
        healthcheck:
          type: tcp
          endpoint: "localhost:5432"

      - name: backend
        type: process
        command: "npm start"
        workdir: "backend"
        port: 8080
        dependencies:
          - database
        healthcheck:
          type: http
          endpoint: "http://localhost:8080/health"

      - name: frontend
        type: process
        command: "npm run dev"
        workdir: "frontend"
        port: 3000
        dependencies:
          - backend

    modes:
      default:
        name: "Development"
        services:
          - database
          - backend
          - frontend
```

### 2. Set Default Project (Optional)

Configure DevUp to work from anywhere:

```bash
# One-time setup - set your default project
devup project set /path/to/your/project

# Now run from anywhere:
devup start
devup status
devup stop
```

Or use environment variable:

```bash
# Add to ~/.zshrc or ~/.bashrc
export DEVUP_DEFAULT_PROJECT="/path/to/your/project"
```

### 3. Start Your Application

```bash
# Install dependencies (if configured)
devup install

# Setup environment (if configured)
devup setup

# Start services
devup start

# Check status
devup status

# Stop services
devup stop
```

---

## 📖 Core Concepts

### Applications

Applications are the top-level containers in DevUp. Each app has its own services, modes, and configuration:

```yaml
apps:
  web-app:
    name: "Web Application"
    workdir: "/path/to/app"
    services: [...]
    modes: [...]
```

### Services

Services are the building blocks - processes, containers, or any executable:

```yaml
services:
  - name: "api"
    type: "process"
    command: "npm start"
    port: 8080
    dependencies:
      - database
    healthcheck:
      type: http
      endpoint: "http://localhost:8080/health"
    environment:
      NODE_ENV: production
```

### Modes

Modes let you define different configurations for different scenarios:

```yaml
modes:
  development:
    name: "Development"
    services: [database, backend, frontend]
    environment:
      DEBUG: "true"

  frontend-only:
    name: "Frontend Only"
    services: [frontend]
    overrides:
      - service: frontend
        environment:
          API_URL: "https://staging.api.example.com"
```

### Dependencies

DevUp automatically orders service startup based on dependencies:

```yaml
services:
  - name: backend
    dependencies:
      - database  # Backend waits for database
      - cache     # Backend waits for cache
```

### Health Checks

Ensure services are ready before proceeding:

```yaml
healthcheck:
  type: http                    # http | tcp | exec
  endpoint: "http://localhost:8080/health"
  timeout: 30s
  interval: 2s
  retries: 15
```

---

## 🔧 Command Reference

### Core Commands

```bash
devup init              # Initialize configuration (auto-detect project)
devup list              # List all applications
devup start             # Start services
devup stop              # Stop services
devup status            # Check service status
devup install           # Install dependencies
devup setup             # Setup environment
devup clean             # Clean resources
devup env               # Show environment variables
```

### Project Management

```bash
devup project set <path>     # Set default project directory
devup project show           # Show current default project
devup project clear          # Clear default project
```

### Global Flags

```bash
-c, --config <file>   # Specify config file path
-l, --local           # Use local directory config
-a, --app <name>      # Select specific application
-v, --verbose         # Enable verbose output
```

### Common Usage Patterns

```bash
# Start with specific mode
devup start --mode production

# Start specific app
devup start -a my-app

# Use custom config file
devup start -c /path/to/config.yaml

# Dry-run install
devup install --dry-run

# Setup with defaults
devup setup --use-defaults
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **[docs/QUICKSTART.md](docs/QUICKSTART.md)** | Get up and running in 5 minutes |
| **[docs/TUTORIAL.md](docs/TUTORIAL.md)** | Comprehensive beginner-friendly guide |
| **[docs/INIT_COMMAND.md](docs/INIT_COMMAND.md)** | Auto-initialization guide for existing projects |
| **[docs/QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md)** | Command cheat sheet |
| **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** | Technical architecture and design decisions |
| **[docs/CHANGELOG.md](docs/CHANGELOG.md)** | Version history and release notes |
| **[docs/FEATURE_SUMMARY.md](docs/FEATURE_SUMMARY.md)** | Feature summary for install and setup |
| **[docs/INSTALL_SETUP_GUIDE.md](docs/INSTALL_SETUP_GUIDE.md)** | Install and setup guide |
| **[docs/DEVUP_ENV_CONFIG.md](docs/DEVUP_ENV_CONFIG.md)** | Environment configuration guide |
| **[docs/VERSIONING.md](docs/VERSIONING.md)** | Versioning guide |
| **[docs/MAKEFILE.md](docs/MAKEFILE.md)** | Makefile usage guide |
| **[examples/](examples/)** | Example configurations for different use cases |

### Topic Guides

- **Configuration**: See examples in [examples/](examples/)
- **Environment Setup**: Read [docs/QUICKSTART.md](docs/QUICKSTART.md)
- **Versioning**: See [docs/VERSIONING.md](docs/VERSIONING.md)
- **Makefile Usage**: See [docs/MAKEFILE.md](docs/MAKEFILE.md) (legacy reference)

---

## 🎨 Usage Examples

### Multi-Application Workspace

```yaml
apps:
  frontend:
    name: "Frontend App"
    workdir: "./apps/frontend"
    services:
      - name: dev-server
        command: "npm run dev"
        port: 3000

  backend:
    name: "Backend API"
    workdir: "./apps/backend"
    services:
      - name: api
        command: "go run main.go"
        port: 8080

  mobile:
    name: "Mobile App"
    workdir: "./apps/mobile"
    services:
      - name: expo
        command: "npx expo start"
        port: 19000
```

```bash
# Start individual apps
devup start -a frontend
devup start -a backend
devup start -a mobile
```

### Environment-Specific Modes

```yaml
modes:
  local:
    name: "Local Development"
    services: [frontend, mock-api]
    environment:
      API_URL: "http://localhost:8001"

  staging:
    name: "Staging Environment"
    services: [frontend]
    overrides:
      - service: frontend
        environment:
          API_URL: "https://staging.api.example.com"

  production:
    name: "Production"
    services: [database, redis, backend, frontend]
    environment:
      NODE_ENV: production
```

### Automated Installation & Setup

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

  steps:
    - name: "Install UI dependencies"
      workdir: "frontend"
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

More examples in the [examples/](examples/) directory!

---

## 🏗️ Architecture

DevUp is built with extensibility as a core principle:

```
┌─────────────────────────────────────┐
│        devup.yaml (Config)          │
│  ┌──────┬──────┬──────┬──────────┐  │
│  │ App1 │ App2 │ App3 │   ...    │  │
│  └──────┴──────┴──────┴──────────┘  │
└─────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│      Configuration Loader            │
│  • YAML parsing & validation        │
│  • App resolution                   │
└─────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│       Service Manager                │
│  • Lifecycle orchestration          │
│  • Dependency resolution            │
│  • Health monitoring                │
└─────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│       Service Runners                │
│  ┌─────────┬──────────┬──────────┐  │
│  │ Process │  Docker  │  Custom  │  │
│  └─────────┴──────────┴──────────┘  │
└─────────────────────────────────────┘
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed technical documentation.

---

## 🎓 Best Practices

### 1. Organize by Application

Keep each application's configuration isolated:

```yaml
apps:
  frontend:    # UI application
  backend:     # API service
  workers:     # Background jobs
```

### 2. Use Meaningful Mode Names

```yaml
modes:
  local-dev:        # Local development
  integration:      # Integration testing
  staging:          # Staging environment
  debug:            # Debug mode
```

### 3. Configure Health Checks

Always add health checks for reliable startup:

```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8080/health"
  timeout: 30s
  retries: 15
```

### 4. Leverage Dependencies

Let DevUp handle the startup order:

```yaml
services:
  - name: backend
    dependencies:
      - database
      - cache
```

### 5. Use Lifecycle Hooks

Automate common tasks:

```yaml
hooks:
  pre_start:
    - "mkdir -p logs data tmp"
  post_start:
    - "echo 'Application started at http://localhost:3000'"
```

### 6. Version Control Your Config

```bash
git add devup.yaml
git commit -m "Add DevUp configuration"
```

---

## 🤝 Contributing

We welcome contributions from the community!

For details on how to contribute, please see our [Contributing Guide](CONTRIBUTING.md).

We also have a [Code of Conduct](CODE_OF_CONDUCT.md) that we expect all contributors to adhere to.


---

## 🐛 Troubleshooting

### Services won't start

```bash
# Check logs
tail -f logs/*.log

# Verbose output
devup start -v

# Check status
devup status
```

### Configuration errors

```bash
# Validate config
devup list -v

# Check specific app
devup list -a my-app
```

### Port conflicts

```bash
# Find process using port
lsof -ti tcp:3000

# Kill process
kill $(lsof -ti tcp:3000)
```

### Config not found

```bash
# Check current config location
devup env -v

# Set default project
devup project set /path/to/project

# Or use explicit path
devup start -c /path/to/devup.yaml
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

[Report Bug](https://github.com/everton/devup/issues) · [Request Feature](https://github.com/everton/devup/issues) · [Documentation](https://github.com/everton/devup/wiki)

</div>

# DevUp Configuration Examples

This directory contains example configurations demonstrating various DevUp features and use cases.

## 📑 Table of Contents

- [Getting Started](#getting-started)
- [Available Examples](#available-examples)
- [How to Use Examples](#how-to-use-examples)
- [Example Categories](#example-categories)

---

## Getting Started

### Prerequisites

Before running these examples, ensure you have:

1. **DevUp installed**:
   ```bash
   cd ..
   make -f Makefile.devup build
   ./build/devup --version
   ```

2. **Required tools** (varies by example):
   - Node.js 18+
   - Python 3.12+
   - Docker (for some examples)

### Running an Example

```bash
# From the devup_v2 directory
./build/devup list -c examples/<example-file>.yaml

# Start services
./build/devup start -c examples/<example-file>.yaml

# Check status
./build/devup status -c examples/<example-file>.yaml

# Stop services
./build/devup stop -c examples/<example-file>.yaml
```

---

## Available Examples

### 1. Simple Web Application
**File**: [simple-webapp.yaml](simple-webapp.yaml)

A basic full-stack application demonstrating fundamental DevUp concepts.

**Features**:
- PostgreSQL database
- Express.js backend API
- React frontend
- Service dependencies
- Multiple modes (full stack, frontend-only, backend-only)
- Health checks

**Services**:
- `database` - PostgreSQL (port 5432)
- `backend` - Express API (port 8080)
- `frontend` - React dev server (port 3000)

**Modes**:
- `default` - All services
- `frontend-only` - Just the UI (useful for frontend development)
- `backend-only` - API and database

**Use Case**: Perfect for learning DevUp basics or starting a new full-stack project.

```bash
# Start full stack
./build/devup start -c examples/simple-webapp.yaml

# Start frontend only (for UI development)
./build/devup start -c examples/simple-webapp.yaml --mode frontend-only
```

---

### 2. Installation Demo
**File**: [install-demo.yaml](install-demo.yaml)

Demonstrates the `devup install` command capabilities.

**Features**:
- Multiple package managers (brew, npm, pip)
- Dependency checking
- Optional dependencies
- Custom installation commands
- Installation steps with skip conditions
- Pre/post install hooks

**What It Does**:
- Checks for Python 3.12+
- Checks for Node.js
- Installs npm packages
- Creates Python virtual environment
- Installs Python dependencies

**Use Case**: Learning dependency management, automating project setup.

```bash
# Preview what would be installed
./build/devup install -c examples/install-demo.yaml --dry-run

# Install everything
./build/devup install -c examples/install-demo.yaml

# Skip optional dependencies
./build/devup install -c examples/install-demo.yaml --skip-optional
```

---

### 3. Setup Demo
**File**: [test-setup-demo.yaml](test-setup-demo.yaml)

Demonstrates the `devup setup` command capabilities.

**Features**:
- Directory creation
- File generation and copying
- Environment variable management
- Secret generation
- User prompts
- Setup validation
- Pre/post setup hooks

**What It Does**:
- Creates logs, tmp, data directories
- Generates secure random secrets
- Prompts for database URL
- Sets default environment values
- Creates .env file
- Runs validation checks

**Use Case**: Learning environment setup, automating configuration.

```bash
# Preview what would be done
./build/devup setup -c examples/test-setup-demo.yaml --dry-run

# Interactive setup (prompts for values)
./build/devup setup -c examples/test-setup-demo.yaml

# Use all defaults (no prompts)
./build/devup setup -c examples/test-setup-demo.yaml --use-defaults

# Custom .env output location
./build/devup setup -c examples/test-setup-demo.yaml --env-output .env.local
```

---

### 4. Basic Demo
**File**: [demo-test.yaml](demo-test.yaml)

Minimal example for quick testing.

**Features**:
- Single service
- Simple health check
- No complex dependencies

**Use Case**: Quick DevUp testing, CI/CD integration testing.

```bash
./build/devup start -c examples/demo-test.yaml
```

---

### 5. Starter Template
**File**: [devup.yaml](devup.yaml)

A comprehensive template ready to customize for your project.

**Features**:
- Multi-service setup (web, api, database)
- Multiple modes (development, frontend-only, production)
- Installation configuration
- Setup configuration
- Lifecycle hooks
- Health checks
- Dependencies

**Use Case**: Starting point for new DevUp configurations.

```bash
# Copy to your project
cp examples/devup.yaml /path/to/your/project/

# Edit and customize
vim /path/to/your/project/devup.yaml

# Use it
cd /path/to/your/project
devup start
```

---

## How to Use Examples

### 1. Explore an Example

```bash
# View the configuration
cat examples/simple-webapp.yaml

# List applications and modes
./build/devup list -c examples/simple-webapp.yaml
```

### 2. Run an Example

```bash
# Start with default mode
./build/devup start -c examples/simple-webapp.yaml

# Start with specific mode
./build/devup start -c examples/simple-webapp.yaml --mode frontend-only

# Check status
./build/devup status -c examples/simple-webapp.yaml
```

### 3. Stop an Example

```bash
./build/devup stop -c examples/simple-webapp.yaml
```

### 4. Customize for Your Project

```bash
# Copy an example
cp examples/simple-webapp.yaml myproject.yaml

# Edit the configuration
vim myproject.yaml

# Update paths, commands, ports, etc.
# Test your configuration
./build/devup list -c myproject.yaml
./build/devup start -c myproject.yaml
```

---

## Example Categories

### By Complexity

#### Beginner
- [simple-webapp.yaml](simple-webapp.yaml) - Basic concepts
- [demo-test.yaml](demo-test.yaml) - Minimal example
- [devup.yaml](devup.yaml) - Comprehensive template

#### Intermediate
- [install-demo.yaml](install-demo.yaml) - Dependency management
- [test-setup-demo.yaml](test-setup-demo.yaml) - Environment setup

### By Use Case

#### Learning DevUp
- Start with [simple-webapp.yaml](simple-webapp.yaml)
- Then try [install-demo.yaml](install-demo.yaml)
- Finally explore [devup.yaml](devup.yaml) template

#### Frontend Development
- [simple-webapp.yaml](simple-webapp.yaml) in `frontend-only` mode
- [devup.yaml](devup.yaml) template customized for your stack

#### Full-Stack Development
- [simple-webapp.yaml](simple-webapp.yaml) in `default` mode
- [devup.yaml](devup.yaml) template with all services

#### Team Onboarding
- [devup.yaml](devup.yaml) template with install + setup
- Includes automated dependency installation and environment setup

---

## Common Patterns

### Pattern 1: Development Modes

Most examples include multiple modes for different development scenarios:

```yaml
modes:
  default:        # Quick start, usually full stack
  frontend-only:  # UI development only
  backend-only:   # API development only
  production:     # Production-like environment
```

### Pattern 2: Service Dependencies

Services start in dependency order:

```yaml
services:
  - name: database
    # No dependencies, starts first

  - name: backend
    dependencies:
      - database  # Waits for database

  - name: frontend
    dependencies:
      - backend  # Waits for backend
```

### Pattern 3: Health Checks

Ensure services are ready:

```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8080/health"
  timeout: 30s
  interval: 2s
  retries: 15
```

### Pattern 4: Mode Overrides

Customize services per mode:

```yaml
modes:
  production:
    overrides:
      - service: backend
        command: "npm run start:prod"
        environment:
          NODE_ENV: production
```

---

## Tips for Creating Your Own Config

### 1. Start Small

Begin with a simple config:
```yaml
version: "1.0"
apps:
  my-app:
    name: "My App"
    workdir: "."
    services:
      - name: main
        command: "npm start"
    modes:
      default:
        services: [main]
```

### 2. Add Services Incrementally

Add one service at a time and test:
```bash
# After adding each service
devup list -c myconfig.yaml
devup start -c myconfig.yaml
devup status -c myconfig.yaml
```

### 3. Configure Health Checks

Always add health checks:
```yaml
healthcheck:
  type: http  # or tcp or exec
  endpoint: "http://localhost:8080"
```

### 4. Use Dependencies

Let DevUp manage startup order:
```yaml
services:
  - name: api
    dependencies:
      - database
```

### 5. Create Multiple Modes

Support different development scenarios:
```yaml
modes:
  local:         # Local dev
  integration:   # Integration testing
  production:    # Prod-like environment
```

### 6. Add Lifecycle Hooks

Automate common tasks:
```yaml
hooks:
  pre_start:
    - "mkdir -p logs"
  post_start:
    - "echo 'Started at http://localhost:3000'"
```

---

## Troubleshooting Examples

### Services Don't Start

```bash
# Check verbose output
./build/devup start -c examples/simple-webapp.yaml -v

# Check service logs
tail -f logs/*.log
```

### Port Conflicts

```bash
# Find what's using the port
lsof -ti tcp:3000

# Kill the process
kill $(lsof -ti tcp:3000)

# Or update port in the config
```

### Config Not Valid

```bash
# Validate by listing apps
./build/devup list -c examples/yourconfig.yaml -v
```

---

## Contributing Examples

Have a great DevUp configuration? Share it!

1. Create your example config
2. Add it to this directory
3. Update this README with a description
4. Submit a pull request

**Good examples include**:
- Clear comments
- Multiple modes
- Health checks
- Realistic use cases
- Different tech stacks

---

## Additional Resources

- **Main README**: [../README.md](../README.md)
- **Quick Start**: [../QUICKSTART.md](../QUICKSTART.md)
- **Architecture**: [../ARCHITECTURE.md](../ARCHITECTURE.md)
- **Changelog**: [../CHANGELOG.md](../CHANGELOG.md)

---

## Questions or Issues?

- Check the main [README](../README.md) for documentation
- Review [QUICKSTART.md](../QUICKSTART.md) for getting started
- Open an issue on GitHub

---

**Happy DevUp-ing!** 🚀

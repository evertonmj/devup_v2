# DevUp Install & Setup Guide

Complete guide for using DevUp's automated installation and environment setup features.

## Overview

DevUp provides two powerful commands to automate project bootstrapping:

- **`devup install`** - Install dependencies (Homebrew, npm, pip, etc.)
- **`devup setup`** - Set up environment (directories, .env file, scripts)

## Installation Command

### What It Does

The `install` command automates dependency installation by:

1. Checking which dependencies are already installed
2. Installing missing dependencies using appropriate package managers
3. Running custom installation steps
4. Executing pre/post install hooks

### Configuration

```yaml
install:
  dependencies:
    - name: "Python 3.12+"
      type: brew              # Package manager type
      package: python@3.12    # Package name
      check: "python3.12 --version"  # Command to verify installation
      optional: false         # Required or optional

    - name: "Node.js"
      type: brew
      package: node
      check: "node --version"

    - name: "Redis"
      type: brew
      package: redis
      optional: true          # Won't fail if installation fails

  steps:
    - name: "Install UI dependencies"
      description: "Install npm packages"
      workdir: "ui"
      commands:
        - "npm install"
      skip_if: "[ -d node_modules ]"  # Skip if already done

    - name: "Create Python venv"
      workdir: "service"
      commands:
        - "python3 -m venv venv"
        - "source venv/bin/activate && pip install -r requirements.txt"
```

### Supported Dependency Types

| Type | Package Manager | Example |
|------|----------------|---------|
| `brew` | Homebrew (macOS) | `python@3.12`, `node`, `postgresql` |
| `npm` | npm (global) | `typescript`, `yarn`, `pm2` |
| `pip` | pip/pip3 | `flask`, `django`, `requests` |
| `apt` | apt-get (Linux) | `build-essential`, `curl` |
| `go` | Go | N/A (checks go installation) |
| `custom` | Custom command | Use `install_cmd` |

### Usage Examples

```bash
# Basic installation
devup install

# Install for specific app
devup install -a myapp

# Preview what would be installed (no changes)
devup install --dry-run

# Skip optional dependencies
devup install --skip-optional

# Verbose output
devup install -v
```

### Custom Installation Commands

For dependencies not supported by standard package managers:

```yaml
dependencies:
  - name: "Custom Tool"
    type: custom
    check: "my-tool --version"
    install_cmd: "curl -sSL https://example.com/install.sh | bash"
```

---

## Setup Command

### What It Does

The `setup` command automates environment setup by:

1. Creating required directories
2. Generating or copying configuration files
3. Creating .env file with environment variables
4. Running setup scripts
5. Validating the setup

### Configuration

```yaml
setup:
  # Directories to create
  directories:
    - "logs"
    - "tmp"
    - "data"
    - "uploads"

  # Files to create
  files:
    - path: "config/app.yml"
      source: "config/app.yml.template"  # Copy from template

    - path: "scripts/init.sh"
      content: |
        #!/bin/bash
        echo "Initialization script"

  # Environment variables
  env_vars:
    # Generated values
    - name: "JWT_SECRET_KEY"
      description: "Secret key for JWT tokens"
      generate: "openssl rand -base64 32"
      required: true

    # Prompted values
    - name: "DATABASE_URL"
      description: "Database connection string"
      prompt: true
      default: "postgresql://localhost/mydb"
      required: true

    # Static defaults
    - name: "LOG_LEVEL"
      default: "INFO"
      required: false

  # Scripts to run
  scripts:
    - "chmod +x scripts/*.sh"
    - "./scripts/init-database.sh"

  # Validation checks
  checks:
    - name: "Virtual environment exists"
      command: "[ -d venv ]"
      message: "Python virtual environment not found"

    - name: "Node modules installed"
      command: "[ -d node_modules ]"
      message: "Run npm install first"
```

### Environment Variable Types

#### 1. Generated Values

Automatically generate values using shell commands:

```yaml
env_vars:
  - name: "SECRET_KEY"
    generate: "openssl rand -hex 32"
    required: true

  - name: "ENCRYPTION_KEY"
    generate: "python3 -c 'from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())'"
```

#### 2. Prompted Values

Ask user for input:

```yaml
env_vars:
  - name: "API_KEY"
    description: "Your API key from example.com"
    prompt: true
    required: true

  - name: "DATABASE_HOST"
    description: "Database hostname"
    prompt: true
    default: "localhost"
```

#### 3. Static Defaults

Use default values:

```yaml
env_vars:
  - name: "NODE_ENV"
    default: "development"

  - name: "PORT"
    default: "3000"
```

### Usage Examples

```bash
# Basic setup
devup setup

# Setup for specific app
devup setup -a myapp

# Preview what would be done (no changes)
devup setup --dry-run

# Use all defaults without prompting
devup setup --use-defaults

# Skip all prompts (use existing or defaults)
devup setup --skip-prompts

# Custom .env file location
devup setup --env-output .env.local

# Verbose output
devup setup -v
```

---

## Complete Workflow Example

### Configuration File

```yaml
version: "1.0"

apps:
  my-fullstack-app:
    name: "My Full-Stack App"
    workdir: "."

    # Installation
    install:
      dependencies:
        - name: "Python 3.11+"
          type: brew
          package: python@3.11
          check: "python3.11 --version"

        - name: "Node.js 18+"
          type: brew
          package: node
          check: "node --version"

        - name: "PostgreSQL"
          type: brew
          package: postgresql@15
          optional: true

      steps:
        - name: "Install UI dependencies"
          workdir: "frontend"
          commands:
            - "npm install"
          skip_if: "[ -d node_modules ]"

        - name: "Create Python virtual environment"
          workdir: "backend"
          commands:
            - "python3.11 -m venv venv"
          skip_if: "[ -d venv ]"

        - name: "Install Python dependencies"
          workdir: "backend"
          commands:
            - "source venv/bin/activate && pip install -r requirements.txt"

    # Setup
    setup:
      directories:
        - "logs"
        - "uploads"
        - "tmp"

      env_vars:
        - name: "SECRET_KEY"
          description: "Application secret key"
          generate: "openssl rand -hex 32"
          required: true

        - name: "DATABASE_URL"
          description: "PostgreSQL connection string"
          prompt: true
          default: "postgresql://localhost:5432/myapp"
          required: true

        - name: "REDIS_URL"
          description: "Redis connection string"
          default: "redis://localhost:6379"

        - name: "DEBUG"
          default: "true"

      scripts:
        - "chmod +x scripts/*.sh"

      checks:
        - name: "Backend venv exists"
          command: "[ -d backend/venv ]"

        - name: "Frontend node_modules exists"
          command: "[ -d frontend/node_modules ]"

    # Services
    services:
      - name: frontend
        command: "cd frontend && npm run dev"
        port: 3000

      - name: backend
        command: "cd backend && source venv/bin/activate && python app.py"
        port: 8000
        dependencies:
          - frontend

    modes:
      default:
        services:
          - frontend
          - backend

    # Hooks
    hooks:
      pre_install:
        - "echo 'Starting installation...'"

      post_install:
        - "echo '✅ Installation complete!'"
        - "echo 'Next: Run devup setup'"

      pre_setup:
        - "echo 'Setting up environment...'"

      post_setup:
        - "echo '✅ Setup complete!'"
        - "echo 'Next: Run devup start'"

      pre_start:
        - "mkdir -p logs"

      post_start:
        - "echo '🚀 Application running!'"
        - "echo 'Frontend: http://localhost:3000'"
        - "echo 'Backend:  http://localhost:8000'"
```

### Complete Setup Workflow

```bash
# 1. Check prerequisites
devup install --dry-run

# 2. Install dependencies
devup install

# 3. Preview setup
devup setup --dry-run

# 4. Run setup
devup setup

# 5. Start the application
devup start

# 6. Check status
devup status

# Later: stop services
devup stop
```

---

## Tips & Best Practices

### 1. Always Use Dry-Run First

```bash
# Preview before making changes
devup install --dry-run
devup setup --dry-run

# Then execute
devup install
devup setup
```

### 2. Use Skip Conditions

Avoid re-running expensive operations:

```yaml
steps:
  - name: "Install dependencies"
    commands:
      - "npm install"
    skip_if: "[ -d node_modules ]"  # Skip if already installed
```

### 3. Mark Optional Dependencies

Don't fail the entire installation for optional tools:

```yaml
dependencies:
  - name: "Redis"
    type: brew
    package: redis
    optional: true  # Won't fail if unavailable
```

### 4. Use Environment Variable Generation

Automate secret generation:

```yaml
env_vars:
  - name: "SECRET_KEY"
    generate: "openssl rand -hex 32"  # Auto-generate secure random value
```

### 5. Add Validation Checks

Verify setup completed correctly:

```yaml
checks:
  - name: "Dependencies installed"
    command: "[ -d node_modules ] && [ -d venv ]"
    message: "Run devup install first"
```

### 6. Use Hooks for Automation

```yaml
hooks:
  post_install:
    - "echo 'Run: devup setup' to continue"

  post_setup:
    - "echo 'Run: devup start' to launch app"
```

---

## Troubleshooting

### Installation Issues

**Problem**: Dependency check fails

```bash
# Check manually
which python3.12
brew list python@3.12

# Force reinstall
brew reinstall python@3.12
```

**Problem**: Custom check command not working

```yaml
# Make sure check command returns 0 on success
check: "python3.12 --version > /dev/null 2>&1"
```

### Setup Issues

**Problem**: Environment variables not generated

```bash
# Verify generation command works
openssl rand -base64 32

# Check .env file was created
cat .env
```

**Problem**: Prompts not working

```bash
# Use --use-defaults to skip prompts
devup setup --use-defaults

# Or specify values in config
default: "your-value"
```

### Permission Issues

```bash
# Make scripts executable
chmod +x tools/*.sh

# Or add to setup
scripts:
  - "chmod +x tools/*.sh"
```

---

## Real-World Examples

### Example 1: Django Application

```yaml
install:
  dependencies:
    - name: Python
      type: brew
      package: python@3.11

    - name: PostgreSQL
      type: brew
      package: postgresql@15

  steps:
    - name: "Create venv"
      commands:
        - "python3 -m venv venv"

    - name: "Install Django"
      commands:
        - "source venv/bin/activate && pip install -r requirements.txt"

setup:
  directories:
    - "media"
    - "static"
    - "logs"

  env_vars:
    - name: "SECRET_KEY"
      generate: "python3 -c 'from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())'"

    - name: "DATABASE_URL"
      default: "postgresql://localhost/mydb"

  scripts:
    - "source venv/bin/activate && python manage.py migrate"
```

### Example 2: React + Node.js

```yaml
install:
  dependencies:
    - name: Node.js
      type: brew
      package: node

  steps:
    - name: "Install frontend"
      workdir: "client"
      commands: ["npm install"]

    - name: "Install backend"
      workdir: "server"
      commands: ["npm install"]

setup:
  env_vars:
    - name: "JWT_SECRET"
      generate: "openssl rand -base64 32"

    - name: "PORT"
      default: "3000"
```

---

## Next Steps

After running install and setup:

1. **Start your application**: `devup start`
2. **Check status**: `devup status`
3. **View logs**: Check `logs/` directory
4. **Stop when done**: `devup stop`

For more information, see:
- [../README.md](../README.md) - Main documentation
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [examples/](examples/) - Example configurations

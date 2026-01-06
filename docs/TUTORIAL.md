# DevUp Tutorial for Beginners

Welcome to DevUp! This comprehensive tutorial will guide you through everything you need to know to use DevUp, even if you're new to development tools. By the end of this tutorial, you'll be able to manage complex development environments with ease.

## Table of Contents

1. [What is DevUp?](#what-is-devup)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Your First DevUp Project](#your-first-devup-project)
5. [Understanding the Configuration File](#understanding-the-configuration-file)
6. [Working with Services](#working-with-services)
7. [Using Modes](#using-modes)
8. [Health Checks](#health-checks)
9. [Advanced Features](#advanced-features)
10. [Common Workflows](#common-workflows)
11. [Troubleshooting](#troubleshooting)

---

## What is DevUp?

DevUp is a tool that helps you **manage multiple services** (like a web server, database, API, etc.) for your development projects. Instead of opening multiple terminal windows and typing commands to start each service, DevUp:

- Starts all your services with a single command
- Ensures services start in the correct order (dependencies)
- Monitors service health
- Manages different configurations (dev, test, production)
- Makes development environment setup reproducible

**Think of it as**: A conductor for an orchestra, where each musician (service) needs to play at the right time.

---

## Prerequisites

Before you begin, make sure you have:

### Required Software

1. **A Terminal/Command Line**
   - Mac: Terminal.app or iTerm2
   - Linux: Your preferred terminal
   - Windows: WSL2 (Windows Subsystem for Linux)

2. **Go Programming Language** (version 1.20 or higher)
   ```bash
   # Check if Go is installed
   go version

   # If not installed, visit: https://golang.org/dl/
   ```

3. **Git** (for version control)
   ```bash
   # Check if Git is installed
   git --version

   # If not installed:
   # Mac: xcode-select --install
   # Linux: sudo apt-get install git
   ```

### Optional but Recommended

- **A text editor** (VS Code, Sublime Text, or Vim)
- **Basic terminal knowledge** (navigating directories, running commands)

---

## Installation

### Step 1: Clone or Download DevUp

```bash
# Option 1: Clone from your repository
git clone <your-devup-repo-url>
cd devup_v2

# Option 2: Or navigate to your existing devup_v2 directory
cd /path/to/devup_v2
```

### Step 2: Build DevUp

```bash
# Build the devup binary
make build

# This creates a binary at: build/devup
```

### Step 3: Install System-Wide (Optional but Recommended)

```bash
# Install devup to /usr/local/bin so you can use it from anywhere
make install

# Now you can run 'devup' from any directory
devup --version
```

### Step 4: Verify Installation

```bash
# Check that devup is working
devup --help

# You should see a list of available commands
```

---

## Your First DevUp Project

Let's create a simple project to understand how DevUp works.

### Quick Start: Auto-Initialize (Recommended)

If you already have a project with a `package.json`, `go.mod`, or similar:

```bash
# Navigate to your project
cd /path/to/your-project

# Let DevUp create the configuration automatically
devup init

# DevUp will:
# - Detect your project type
# - Find services
# - Extract environment variables
# - Create devup.yaml

# Start your application
devup start
```

**Skip to [Understanding the Configuration File](#understanding-the-configuration-file)** if you used `devup init`.

---

### Manual Setup: Create Configuration from Scratch

### Step 1: Create a Project Directory

```bash
# Create a new directory for your project
mkdir my-first-devup-project
cd my-first-devup-project
```

### Step 2: Create a Configuration File

Create a file named `devup.yaml` in your project directory:

```bash
# Using your text editor, create devup.yaml
# Or use this command to create it from the terminal:
cat > devup.yaml << 'EOF'
version: "1.0"

apps:
  hello-world:
    name: "Hello World App"
    description: "My first DevUp application"
    workdir: "."

    services:
      - name: greeter
        command: "echo 'Hello from DevUp!'; sleep 5"
        type: process

    modes:
      default:
        services:
          - greeter
EOF
```

### Step 3: Run Your First Application

```bash
# List available applications
devup list

# Start the application
devup start
```

**What just happened?**
1. DevUp read your `devup.yaml` file
2. Found the `hello-world` app
3. Started the `greeter` service
4. The service ran for 5 seconds and exited

---

## Understanding the Configuration File

The `devup.yaml` file is the heart of DevUp. Let's break down each section:

### Basic Structure

```yaml
version: "1.0"              # DevUp configuration version

apps:                        # You can define multiple applications
  my-app-name:              # Unique identifier for your app
    name: "Display Name"    # Human-readable name
    description: "..."      # What this app does
    workdir: "."           # Where to run commands (. = current directory)

    services: []           # List of services to run
    modes: {}              # Different ways to run the app
```

### Services Section

Services are the individual programs that make up your application:

```yaml
services:
  - name: frontend          # Unique name for this service
    type: process          # How to run it (process, docker, tmux)
    command: "npm start"   # The actual command to execute
    workdir: "frontend"    # Where to run this command
    port: 3000            # Port this service listens on
    logfile: "logs/frontend.log"  # Where to save logs
    dependencies:          # Services that must start first
      - database
    environment:           # Environment variables
      NODE_ENV: development
    healthcheck:          # How to verify service is ready
      type: http
      endpoint: "http://localhost:3000"
      timeout: 30s
```

**Field Explanations:**

- **name**: A unique identifier (no spaces)
- **type**:
  - `process` - Run as a regular command
  - `docker` - Run in Docker (coming soon)
  - `tmux` - Run in tmux session (coming soon)
- **command**: The actual command line to execute
- **workdir**: Directory to run the command in (relative to app workdir)
- **port**: Port number the service uses (optional, for documentation)
- **logfile**: Where to save output (DevUp creates this automatically)
- **dependencies**: List of service names that must start before this one
- **environment**: Key-value pairs of environment variables
- **healthcheck**: Configuration for checking if service is ready

### Modes Section

Modes let you run different combinations of services:

```yaml
modes:
  default:                  # Mode name (use with: devup start -m default)
    name: "Development"     # Display name
    description: "Full stack with all services"
    services:              # Which services to run
      - database
      - backend
      - frontend
    environment:           # Environment variables for this mode
      DEBUG: "true"
    overrides:            # Override service settings for this mode
      - service: backend
        command: "npm run dev:debug"  # Different command
        port: 8081                    # Different port
        environment:
          LOG_LEVEL: debug
```

---

## Working with Services

### Creating a Multi-Service Application

Let's create a more realistic example with multiple services:

```yaml
version: "1.0"

apps:
  blog-app:
    name: "Simple Blog"
    description: "A blog with frontend and API"
    workdir: "."

    services:
      # Service 1: Database
      - name: database
        type: process
        command: "echo 'Database simulation running...'; sleep 1000"
        port: 5432
        logfile: "logs/database.log"
        healthcheck:
          type: exec
          endpoint: "echo 'Database is ready'"
          timeout: 5s
          interval: 2s
          retries: 3

      # Service 2: API (depends on database)
      - name: api
        type: process
        command: "echo 'API server starting...'; sleep 1000"
        port: 8000
        logfile: "logs/api.log"
        dependencies:
          - database
        environment:
          DATABASE_URL: "postgres://localhost:5432/blog"
          PORT: "8000"
        healthcheck:
          type: exec
          endpoint: "curl -f http://localhost:8000/health || exit 1"
          timeout: 10s
          interval: 3s
          retries: 5

      # Service 3: Frontend (depends on API)
      - name: frontend
        type: process
        command: "echo 'Frontend server starting...'; sleep 1000"
        port: 3000
        logfile: "logs/frontend.log"
        dependencies:
          - api
        environment:
          API_URL: "http://localhost:8000"

    modes:
      default:
        name: "Full Stack"
        description: "All services"
        services:
          - database
          - api
          - frontend

      api-only:
        name: "API Development"
        description: "Just API and database"
        services:
          - database
          - api

      frontend-only:
        name: "Frontend Development"
        description: "Frontend with mocked API"
        services:
          - frontend
        overrides:
          - service: frontend
            environment:
              API_URL: "https://mock-api.example.com"

    hooks:
      pre_start:
        - "mkdir -p logs"
        - "echo 'Starting blog app...'"
      post_start:
        - "echo 'Blog is ready at http://localhost:3000'"
```

### Save this to `devup.yaml` and try:

```bash
# List all apps
devup list

# Start in default mode (all services)
devup start

# In a new terminal, check status
devup status

# Stop all services
devup stop

# Start in api-only mode
devup start -m api-only

# Stop
devup stop
```

---

## Using Modes

Modes are different ways to run your application. They're useful for:

- **Development vs Production**: Different configurations
- **Frontend vs Backend work**: Run only what you need
- **Testing**: Isolated environments

### Example: Creating Modes for Different Workflows

```yaml
modes:
  # For frontend developers
  frontend-dev:
    name: "Frontend Development"
    description: "UI work with mocked backend"
    services:
      - frontend
    environment:
      REACT_APP_API_MODE: "mock"
    overrides:
      - service: frontend
        environment:
          REACT_APP_DEBUG: "true"

  # For backend developers
  backend-dev:
    name: "Backend Development"
    description: "API with database"
    services:
      - database
      - api
    environment:
      DEBUG: "true"
      LOG_LEVEL: "debug"

  # For full integration
  integration:
    name: "Integration Testing"
    description: "All services with test data"
    services:
      - database
      - api
      - frontend
    environment:
      NODE_ENV: "test"
      USE_TEST_DATA: "true"
```

### Using Modes

```bash
# Start in a specific mode
devup start -m frontend-dev

# Or
devup start --mode backend-dev
```

---

## Health Checks

Health checks ensure services are ready before depending services start.

### Types of Health Checks

#### 1. HTTP Health Check

For web services that expose an HTTP endpoint:

```yaml
healthcheck:
  type: http
  endpoint: "http://localhost:8000/health"
  timeout: 30s       # How long to wait for response
  interval: 5s       # How often to check
  retries: 6         # How many times to retry
```

**When to use**: Web servers, APIs, any service with HTTP endpoint

#### 2. TCP Health Check

For services that listen on a TCP port:

```yaml
healthcheck:
  type: tcp
  endpoint: "localhost:5432"
  timeout: 30s
  interval: 3s
  retries: 10
```

**When to use**: Databases, caches, message queues

#### 3. Exec Health Check

Run a custom command to check health:

```yaml
healthcheck:
  type: exec
  endpoint: "pg_isready -h localhost -p 5432"
  timeout: 10s
  interval: 2s
  retries: 5
```

**When to use**: Complex checks, custom validation

### Best Practices for Health Checks

1. **Set realistic timeouts**: Services need time to start
2. **Use appropriate intervals**: Don't check too frequently
3. **Enough retries**: Give services time to become healthy
4. **Create health endpoints**: For your services, create `/health` endpoints

Example health endpoint (Express.js):

```javascript
// In your Express app
app.get('/health', (req, res) => {
  // Check database connection, etc.
  res.status(200).json({ status: 'healthy' });
});
```

---

## Advanced Features

### Hooks

Hooks let you run commands at specific points:

```yaml
hooks:
  pre_start:           # Before starting any service
    - "mkdir -p logs data tmp"
    - "echo 'Preparing environment...'"

  post_start:          # After all services are running
    - "echo 'Application ready!'"
    - "open http://localhost:3000"  # Open browser (Mac)

  pre_stop:           # Before stopping services
    - "echo 'Saving state...'"

  post_stop:          # After all services stopped
    - "echo 'Cleanup complete'"
```

### Environment Variables

Three ways to set environment variables:

#### 1. Service-Level

```yaml
services:
  - name: api
    environment:
      DATABASE_URL: "postgres://localhost/mydb"
      DEBUG: "true"
```

#### 2. Mode-Level (applies to all services in that mode)

```yaml
modes:
  development:
    environment:
      NODE_ENV: "development"
      DEBUG: "*"
```

#### 3. Mode Overrides (specific to one service in a mode)

```yaml
modes:
  development:
    overrides:
      - service: api
        environment:
          EXTRA_DEBUG: "true"
```

### Dependencies

Control startup order:

```yaml
services:
  - name: database
    command: "..."

  - name: cache
    command: "..."
    dependencies:
      - database      # Waits for database first

  - name: api
    command: "..."
    dependencies:
      - database
      - cache        # Waits for both database and cache

  - name: frontend
    command: "..."
    dependencies:
      - api          # Waits for API
```

**DevUp will:**
1. Start `database` first
2. Wait for `database` health check
3. Start `cache`
4. Wait for `cache` health check
5. Start `api`
6. Wait for `api` health check
7. Start `frontend`

---

## Common Workflows

### Workflow 1: Daily Development

```bash
# Morning: Start your project
cd /path/to/project
devup start

# Check everything is running
devup status

# View environment variables
devup env

# Evening: Stop everything
devup stop
```

### Workflow 2: Working on Frontend Only

```bash
# Start just the frontend with mocked backend
devup start -m frontend-only

# Make your changes...

# Stop when done
devup stop
```

### Workflow 3: Switching Modes

```bash
# Start in one mode
devup start -m frontend-only

# Stop
devup stop

# Start in different mode
devup start -m full-stack
```

### Workflow 4: Multiple Projects

```bash
# Project 1
cd /path/to/project1
devup start -c devup.yaml -a my-app-1

# Project 2 (in another terminal)
cd /path/to/project2
devup start -c devup.yaml -a my-app-2

# Stop specific project
cd /path/to/project1
devup stop
```

### Workflow 5: Setting Default Project

```bash
# Set a default project directory
devup project set /path/to/my-main-project

# Now from anywhere:
devup start    # Uses default project

# Clear default
devup project clear
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: "Port already in use"

**Symptom**: Service fails to start, error about port being used

**Solution**:
```bash
# Find what's using the port (example: port 3000)
lsof -i :3000

# Kill the process
kill -9 <PID>

# Or use devup's clean command
devup clean -c devup.yaml
```

#### Issue 2: "Config file not found"

**Symptom**: `devup` can't find your configuration

**Solution**:
```bash
# Make sure you're in the right directory
pwd

# Or specify the config file explicitly
devup start -c /full/path/to/devup.yaml

# Or use the local flag to search current directory
devup start -l
```

#### Issue 3: "Service health check timeout"

**Symptom**: Service starts but DevUp says it's not healthy

**Solution**:
```yaml
# Increase timeout and retries
healthcheck:
  type: http
  endpoint: "http://localhost:8000"
  timeout: 60s      # Increase from 30s
  interval: 5s
  retries: 12       # Increase from 6
```

#### Issue 4: "Services starting in wrong order"

**Symptom**: A service tries to connect to another before it's ready

**Solution**:
```yaml
# Add explicit dependencies
services:
  - name: api
    dependencies:
      - database    # API waits for database
```

#### Issue 5: "Command not found"

**Symptom**: DevUp says command doesn't exist

**Solution**:
```yaml
# Use full path to command
services:
  - name: my-service
    command: "/usr/local/bin/node server.js"  # Instead of: node server.js

# Or ensure PATH is set
services:
  - name: my-service
    command: "node server.js"
    environment:
      PATH: "/usr/local/bin:/usr/bin:/bin"
```

---

## Tips for Success

### 1. Start Simple

Begin with a simple configuration:
- One service
- Default mode
- No health checks

Then gradually add complexity.

### 2. Use Descriptive Names

```yaml
# Good
services:
  - name: user-authentication-api
  - name: payment-processing-service

# Avoid
services:
  - name: service1
  - name: api2
```

### 3. Document Your Configuration

```yaml
services:
  - name: api
    # This is our main REST API
    # It connects to PostgreSQL database
    # Default port: 8000
    command: "npm run start:api"
```

### 4. Use Environment Variables for Secrets

```yaml
# Don't hardcode secrets
environment:
  DB_PASSWORD: "secret123"  # ❌ Bad

# Instead, reference environment variables
environment:
  DB_PASSWORD: "${DB_PASSWORD}"  # ✅ Good
```

### 5. Test Your Configuration

```bash
# Validate your config
devup list -c devup.yaml

# Try starting services one at a time
# Add services to your config gradually
```

### 6. Keep Logs

```yaml
services:
  - name: api
    logfile: "logs/api.log"  # Always specify logfiles
```

Then check logs when troubleshooting:
```bash
tail -f logs/api.log
```

---

## Next Steps

Now that you understand the basics, explore:

1. **Example Configurations**: Check `examples/` directory
   ```bash
   devup list -c examples/simple-webapp.yaml
   ```

2. **Advanced Documentation**:
   - [ARCHITECTURE.md](ARCHITECTURE.md) - How DevUp works internally
   - [INSTALL_SETUP_GUIDE.md](INSTALL_SETUP_GUIDE.md) - Dependency installation

3. **Create Your Own Project**:
   - Take an existing project
   - Write a `devup.yaml` for it
   - Share with your team!

---

## Getting Help

- **Check existing examples**: `examples/` directory
- **Read error messages carefully**: They usually tell you what's wrong
- **Start with verbose mode**: `devup start -v` shows more details
- **Check logs**: Each service has its own log file

---

## Summary

You've learned:

✅ What DevUp is and why it's useful
✅ How to install and configure DevUp
✅ How to write configuration files
✅ How to work with services, modes, and health checks
✅ Common workflows and troubleshooting

**Remember**: DevUp is a tool to make your development easier. Start simple, add complexity as needed, and don't hesitate to experiment!

Happy developing! 🚀

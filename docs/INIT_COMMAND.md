# DevUp Init Command

The `devup init` command intelligently bootstraps a DevUp configuration for your existing project by scanning and analyzing your codebase.

## Features

### 🔍 **Automatic Detection**

The init command automatically detects:

1. **Package Managers**: npm, go, pip, cargo, maven, gradle, bundle, composer
2. **Project Type**: Node.js, Go, Python, Rust, Java, Ruby, PHP
3. **Project Structure**: Frontend, backend, API, service directories
4. **Environment Variables**: From README.md, .env.example, and other docs
5. **Commands**: Dev, build, and test commands
6. **Ports**: Service ports mentioned in documentation
7. **Dependencies**: Service relationships

### 📚 **Documentation Analysis**

Scans and extracts information from:
- `README.md` - Project description, setup instructions, commands
- `SETUP.md` - Environment setup details
- `DEVELOPMENT.md` - Development workflow
- `CONTRIBUTING.md` - Contribution guidelines
- `.env.example` - Environment variable templates
- `docker-compose.yml` - Service definitions

---

## Usage

### Basic Usage

```bash
# Initialize with interactive prompts
devup init

# Auto-detect without prompts
devup init --interactive=false

# Overwrite existing config
devup init --force
```

### Flags

```bash
-i, --interactive   Interactive mode with prompts (default: true)
    --scan          Scan project directory for patterns (default: true)
-f, --force         Overwrite existing devup.yaml
```

---

## How It Works

### 1. **Project Scanning**

The command scans your project directory for:

```
📁 project/
├── package.json          → Detects Node.js, npm commands
├── go.mod                → Detects Go project
├── requirements.txt      → Detects Python project
├── Cargo.toml           → Detects Rust project
├── README.md            → Extracts description, ports, env vars
├── .env.example         → Reads environment variables
├── docker-compose.yml   → Discovers services
├── frontend/            → Detects frontend service
├── backend/             → Detects backend service
└── api/                 → Detects API service
```

### 2. **Documentation Analysis**

Extracts information from markdown files:

**From README.md:**
```markdown
# My Web App

A full-stack application.

## Setup

export DATABASE_URL=postgres://localhost/mydb
export API_KEY=your_key_here

## Running

Frontend: port 3000
Backend: port 8000

npm run dev
```

**Detected:**
- Description: "My Web App"
- Environment: DATABASE_URL, API_KEY
- Ports: 3000, 8000
- Command: npm run dev

### 3. **Intelligent Generation**

Creates a complete `devup.yaml` with:
- ✅ Detected services with proper configuration
- ✅ Health checks based on service type
- ✅ Dependencies (frontend depends on backend)
- ✅ Environment variables from documentation
- ✅ Pre/post hooks for common tasks
- ✅ Multiple modes (development, etc.)

---

## Examples

### Example 1: Node.js Monorepo

**Project Structure:**
```
my-app/
├── frontend/
│   └── package.json
├── backend/
│   └── package.json
└── README.md
```

**README.md:**
```markdown
# My Application

Full-stack app with React and Express.

Frontend runs on port 3000
Backend API on port 8000
```

**Command:**
```bash
cd my-app
devup init --interactive=false
```

**Generated devup.yaml:**
```yaml
version: "1.0"

apps:
  my-app:
    name: "my-app"
    description: "My Application"
    workdir: "."

    services:
      - name: backend
        type: process
        command: "npm run dev"
        workdir: "backend"
        port: 8000
        logfile: "logs/backend.log"
        healthcheck:
          type: http
          endpoint: "http://localhost:8000"
          timeout: 30s
          interval: 5s
          retries: 6

      - name: frontend
        type: process
        command: "npm run dev"
        workdir: "frontend"
        port: 3000
        logfile: "logs/frontend.log"
        dependencies:
          - backend
        healthcheck:
          type: http
          endpoint: "http://localhost:3000"
          timeout: 30s
          interval: 5s
          retries: 6

    modes:
      default:
        name: "Development"
        description: "Full development environment"
        services:
          - backend
          - frontend

    hooks:
      pre_start:
        - "mkdir -p logs"
        - "npm install"
      post_start:
        - "echo 'Application started successfully'"
```

### Example 2: Go Microservices

**Project Structure:**
```
services/
├── user-service/
│   └── go.mod
├── auth-service/
│   └── go.mod
└── README.md
```

**Command:**
```bash
cd services
devup init
```

**Result:**
Detects Go projects and creates configuration with `go run .` commands.

### Example 3: Python API

**Project Structure:**
```
api/
├── requirements.txt
├── main.py
└── README.md
```

**README.md:**
```markdown
# API Service

export DATABASE_URL=postgresql://localhost/api
export SECRET_KEY=changeme

Run: python main.py
Port: 8000
```

**Command:**
```bash
cd api
devup init --interactive=false
```

**Result:**
- Detects Python project
- Extracts DATABASE_URL and SECRET_KEY
- Creates service with `python main.py`
- Sets port to 8000
- Adds `pip install -r requirements.txt` to pre_start hooks

---

## Interactive Mode

When run with `--interactive` (default), the command prompts for confirmation:

```bash
$ devup init

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 DevUp Project Initialization
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 Scanning project directory...
✅ Detected: My Web Application

Project name [my-app]:
Description [My Web Application]:
Detected 2 service(s):
  1. frontend (web) - Command: npm run dev
  2. backend (api) - Command: npm run dev

Keep these services? [Y/n]: y

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Successfully created devup.yaml
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Detection Rules

### Package Managers

| File | Package Manager | Project Type | Dev Command |
|------|----------------|--------------|-------------|
| `package.json` | npm | nodejs | `npm run dev` |
| `go.mod` | go | golang | `go run .` |
| `requirements.txt` | pip | python | `python main.py` |
| `Pipfile` | pipenv | python | `pipenv run python main.py` |
| `Cargo.toml` | cargo | rust | `cargo run` |
| `pom.xml` | maven | java | `mvn spring-boot:run` |
| `build.gradle` | gradle | java | `./gradlew bootRun` |
| `Gemfile` | bundle | ruby | `bundle exec rails s` |
| `composer.json` | composer | php | `php artisan serve` |

### Service Detection

| Directory Name | Service Type | Default Port |
|----------------|--------------|--------------|
| `frontend`, `ui`, `web`, `client` | web | 3000 |
| `backend`, `api`, `server`, `service` | api | 8000 |

### Port Extraction

Looks for patterns in documentation:
- "port 3000"
- "runs on port 8000"
- "PORT=3000"
- ":3000"

### Environment Variables

Extracts from:
1. Lines starting with `export VAR=value`
2. Lines with `ENV VAR=value`
3. `.env.example` file entries

Only captures UPPERCASE variable names (convention for env vars).

---

## Post-Generation Steps

After running `devup init`:

1. **Review devup.yaml**
   ```bash
   cat devup.yaml
   ```

2. **Update commands if needed**
   - Check service commands are correct
   - Verify ports
   - Adjust health check endpoints

3. **Test configuration**
   ```bash
   devup list
   ```

4. **Try starting services**
   ```bash
   devup start
   ```

5. **Customize as needed**
   - Add more modes
   - Configure environment variables
   - Set up additional hooks

---

## Tips

### 1. Prepare Your README

Make sure your README.md includes:
- Clear project description
- Setup instructions with environment variables
- Port numbers for services
- Run commands

### 2. Use .env.example

Create `.env.example` with all required environment variables:
```bash
DATABASE_URL=postgres://localhost/mydb
API_KEY=your_key_here
SECRET_KEY=change_this
PORT=3000
```

### 3. Organize Services

Use standard directory names:
- `frontend/` or `ui/` for frontend
- `backend/` or `api/` for backend
- Each with their own package manager files

### 4. Document Ports

Explicitly mention ports in README:
```markdown
## Ports

- Frontend: 3000
- Backend: 8000
- Database: 5432
```

### 5. Review Generated Config

Always review and customize the generated config:
- Verify commands are correct
- Check health check endpoints exist
- Adjust timeouts if needed
- Add custom environment variables

---

## Troubleshooting

### Issue: Wrong commands detected

**Solution**: Edit `devup.yaml` and update the `command` field for each service.

### Issue: Missing services

**Solution**: Either:
1. Add service directories with standard names (frontend, backend, etc.)
2. Manually add services to generated devup.yaml

### Issue: No environment variables detected

**Solution**: Add them to `.env.example` or document them in README with `export VAR=value` format.

### Issue: Wrong ports

**Solution**: Document ports clearly in README or edit them in devup.yaml after generation.

---

## Advanced Usage

### Custom Project Structure

If your project doesn't follow standard conventions, you can:

1. Generate basic config:
   ```bash
   devup init --scan=false
   ```

2. Manually edit devup.yaml to add services

### Docker Compose Integration

If you have `docker-compose.yml`, services will be detected automatically:

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:15
  redis:
    image: redis:7
```

These will be added as services you can integrate.

---

## See Also

- [TUTORIAL.md](TUTORIAL.md) - Complete DevUp tutorial
- [QUICKSTART.md](QUICKSTART.md) - 5-minute quick start
- [examples/](../examples/) - Example configurations
- [README.md](../README.md) - Main documentation

---

**The `devup init` command makes it easy to adopt DevUp for existing projects without manual configuration!**

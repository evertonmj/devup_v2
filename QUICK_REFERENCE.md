# Quick Reference - New Features in DevUp v1.0.0

## 🎯 Template Variables

**Syntax**: `${VARIABLE_NAME}`

### Built-in Variables
| Variable | Example | Use Case |
|----------|---------|----------|
| `${HOME}` | `/Users/everton.jesus` | User home directory |
| `${CWD}` | `/path/where/devup/run` | Current working directory |
| `${PROJECT_ROOT}` | App's workdir | App root for relative paths |
| `${WORKDIR}` | Same as PROJECT_ROOT | Alias for PROJECT_ROOT |
| `${APP_NAME}` | `myapp` | Application name |

### Examples

```yaml
# Example 1: Log paths always go to execution directory
logfile: "${CWD}/logs/service.log"

# Example 2: User config in home directory
environment:
  CONFIG_PATH: "${HOME}/.config/myapp"

# Example 3: Source code in app directory
workdir: "${PROJECT_ROOT}/backend"

# Example 4: Environment variable fallback
environment:
  DEBUG_MODE: "${DEBUG}"  # Falls back to process.env.DEBUG
```

---

## 🐳 Docker Services

**Type**: `docker`

### Minimal Example
```yaml
- name: "PostgreSQL"
  type: "docker"
  docker:
    image: "postgres:15"
    ports:
      - "5432:5432"
```

### Complete Example
```yaml
- name: "Web App"
  type: "docker"
  port: 3000
  docker:
    image: "node:18-alpine"
    container: "web-app"
    working_dir: "/app"
    volumes:
      - "${PROJECT_ROOT}/src:/app"
      - "/app/node_modules"
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: "development"
      API_URL: "http://localhost:8000"
    pull: true
    remove: true
    restart_policy: "unless-stopped"
    entrypoint: "npm"
    cmd: "start"
```

### Docker Options Reference

| Option | Type | Example | Notes |
|--------|------|---------|-------|
| `image` | string | `postgres:15` | **Required** |
| `container` | string | `my-db` | Defaults to service name |
| `ports` | array | `["8080:8080"]` | External:Internal |
| `volumes` | array | `["/data:/data"]` | Mount paths |
| `environment` | map | `{PORT: "8000"}` | Env vars |
| `networks` | array | `["backend"]` | Docker networks |
| `working_dir` | string | `/app` | Container workdir |
| `entrypoint` | string | `python` | Override entrypoint |
| `cmd` | string | `app.py --debug` | Override command |
| `user` | string | `1000:1000` | User:group |
| `pull` | bool | `true` | Pull latest image |
| `remove` | bool | `true` | Remove after stop |
| `restart_policy` | string | `unless-stopped` | Restart behavior |
| `privileged` | bool | `false` | Privileged mode |
| `labels` | map | `{app: "api"}` | Docker labels |

---

## 📝 Configuration Patterns

### Pattern 1: Mix Process + Docker
```yaml
services:
  - name: "Backend API"
    type: "process"
    command: "python app.py"
    port: 8000
  
  - name: "Database"
    type: "docker"
    docker:
      image: "postgres:15"
      ports: ["5432:5432"]
```

### Pattern 2: Template Variables + Docker
```yaml
services:
  - name: "Dev Container"
    type: "docker"
    docker:
      image: "ubuntu:22.04"
      volumes:
        - "${PROJECT_ROOT}:/workspace"
      working_dir: "/workspace"
      environment:
        HOME: "${HOME}"
        USER_ID: "${USER}"
```

### Pattern 3: Environment Paths
```yaml
environment:
  # Logs go to execution directory
  LOGS: "${CWD}/logs"
  
  # Config in user home
  CONFIG: "${HOME}/.config/app"
  
  # Data in project
  DATA: "${PROJECT_ROOT}/data"
```

---

## 🚀 Getting Started

### Step 1: Try the Example
```bash
cd examples
../build/devup -a docker-demo -c docker-demo.yaml list
```

### Step 2: Create Your Config
```yaml
version: "1.0"
apps:
  myapp:
    workdir: "."
    services:
      - name: "api"
        type: "process"
        command: "npm start"
        port: 3000
        logfile: "${CWD}/logs/api.log"
      
      - name: "db"
        type: "docker"
        docker:
          image: "postgres:15"
          ports: ["5432:5432"]
```

### Step 3: Run It
```bash
devup -a myapp -c my-config.yaml start
```

---

## 🔧 Common Tasks

### View Available Services
```bash
devup list -c config.yaml
```

### Start Specific Mode
```bash
devup -a myapp --mode development start
```

### View Service Logs
```bash
devup logs service-name
```

### Stop All Services
```bash
devup stop
```

---

## ⚠️ Common Issues

### "workdir does not exist"
- Solution: Use relative paths or `${PROJECT_ROOT}`
- Example: `workdir: "./backend"` or `workdir: "${PROJECT_ROOT}/backend"`

### Docker container not starting
- Check logs: `devup logs service-name`
- Verify image exists: `docker images | grep image-name`
- Check ports aren't in use: `lsof -i :port`

### Template variable not resolving
- Use correct syntax: `${VAR_NAME}` (not `$VAR_NAME`)
- Built-ins are case-sensitive: `HOME`, not `home`
- For env vars, ensure they're exported: `export DEBUG=true`

---

## 📚 Full Documentation

- **Detailed Guide**: See `DOCKER_AND_TEMPLATES.md`
- **Implementation Details**: See `IMPLEMENTATION_SUMMARY.md`
- **Release Status**: See `RELEASE_READY.md`
- **Example Config**: See `examples/docker-demo.yaml`

---

## 🎓 Examples

### Microservices Stack
```yaml
services:
  - name: "API"
    type: "process"
    command: "python api.py"
    environment:
      DATABASE_URL: "postgres://localhost/app"
  
  - name: "Frontend"
    type: "docker"
    docker:
      image: "nginx"
      volumes:
        - "${PROJECT_ROOT}/dist:/usr/share/nginx/html"
  
  - name: "Database"
    type: "docker"
    docker:
      image: "postgres:15"
      environment:
        POSTGRES_DB: "app"
```

### Development Environment
```yaml
services:
  - name: "Dev Tools"
    type: "docker"
    docker:
      image: "ubuntu:22.04"
      volumes:
        - "${HOME}/.config:/root/.config"
        - "${PROJECT_ROOT}:/workspace"
      working_dir: "/workspace"
```

### Logging Configuration
```yaml
environment:
  LOG_DIR: "${CWD}/logs"
  LOG_LEVEL: "debug"

services:
  - name: "service"
    logfile: "${CWD}/logs/service.log"
```

---

**Version**: DevUp v1.0.0+
**Last Updated**: Today
**Status**: Production Ready ✅

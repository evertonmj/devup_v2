# Template Variables and Docker Support

This document describes the new features added to DevUp for v1.0.0: template variable support and Docker service types.

## Template Variables

Template variables allow you to use dynamic values in your `devup.yaml` configuration. Variables are referenced using the `${VARIABLE_NAME}` syntax and are resolved at runtime.

### Built-in Variables

DevUp provides several built-in variables that are automatically available:

- **`${HOME}`** - User's home directory (e.g., `/Users/everton.jesus`)
- **`${PROJECT_ROOT}`** - The application's working directory (from `workdir` field)
- **`${WORKDIR}`** - Same as PROJECT_ROOT
- **`${APP_NAME}`** - The name of the application (from `name` field)
- **`${CWD}`** - Current working directory where devup command was executed

### Environment Variable Fallback

If a variable is not found in the built-in variables, DevUp will attempt to resolve it from environment variables. This allows you to use any environment variable in your configuration:

```yaml
services:
  api:
    command: "npm start"
    environment:
      DATABASE_URL: "${DATABASE_URL}"  # Falls back to process.env.DATABASE_URL
      API_PORT: "${PORT:-8080}"        # Note: default values not yet supported
```

### Usage Examples

#### Directory Paths

```yaml
apps:
  myapp:
    workdir: "${PROJECT_ROOT}/backend"
    services:
      server:
        command: "python app.py"
        logfile: "${CWD}/logs/server.log"  # Log to execution directory
        environment:
          DATA_DIR: "${PROJECT_ROOT}/data"
          CACHE_DIR: "${HOME}/.cache/myapp"
```

#### Docker Services

```yaml
services:
  web:
    type: "docker"
    docker:
      image: "node:18"
      volumes:
        - "${PROJECT_ROOT}/frontend:/app"
      working_dir: "/app"
      environment:
        NODE_ENV: "development"
        HOME_DIR: "${HOME}"
```

#### Conditional Configuration

```yaml
apps:
  app:
    environment:
      CONFIG_PATH: "${HOME}/.config/myapp/config.yml"
      LOGS_DIR: "${CWD}/logs"
      DATA_PATH: "${PROJECT_ROOT}/data"
    
    services:
      database:
        workdir: "${PROJECT_ROOT}/db"
        logfile: "${CWD}/logs/database.log"
```

## Docker Service Support

DevUp now supports running services as Docker containers. Use the `docker` service type to containerize your services.

### Basic Docker Service

```yaml
services:
  postgres:
    name: "PostgreSQL"
    type: "docker"
    docker:
      image: "postgres:15-alpine"
      container: "my-postgres"
      ports:
        - "5432:5432"
      environment:
        POSTGRES_PASSWORD: "secret"
```

### Docker Configuration Options

The `docker` section in a service supports the following options:

#### Required
- **`image`** - Docker image to use (e.g., `postgres:15`, `redis:7-alpine`)

#### Common Options
- **`container`** - Container name (defaults to service name if not specified)
- **`ports`** - Port mappings as strings (e.g., `["8080:8080", "9090"]`)
- **`volumes`** - Volume mounts (e.g., `["data:/data", "${PROJECT_ROOT}/src:/app/src"]`)
- **`environment`** - Environment variables to pass to container
- **`networks`** - Docker networks to connect to (e.g., `["backend", "frontend"]`)
- **`working_dir`** - Working directory inside container
- **`pull`** - Always pull image before running (default: false)
- **`remove`** - Remove container when stopped (default: true)

#### Advanced Options
- **`restart_policy`** - Restart policy: `no`, `always`, `on-failure`, `unless-stopped`
- **`entrypoint`** - Override container entrypoint
- **`cmd`** - Override container command
- **`user`** - User to run container as (e.g., `"1000:1000"`)
- **`privileged`** - Run container in privileged mode (default: false)
- **`labels`** - Docker labels as key-value pairs

### Complete Docker Example

```yaml
apps:
  fullstack:
    name: "Full Stack App"
    workdir: "${PROJECT_ROOT}"
    
    services:
      # Database with persistent volume
      db:
        name: "PostgreSQL"
        type: "docker"
        docker:
          image: "postgres:15-alpine"
          container: "app-postgres"
          volumes:
            - "postgres_data:/var/lib/postgresql/data"
          ports:
            - "5432:5432"
          environment:
            POSTGRES_USER: "app"
            POSTGRES_PASSWORD: "secret"
            POSTGRES_DB: "appdb"
          restart_policy: "unless-stopped"
          pull: true
      
      # Cache layer
      cache:
        name: "Redis"
        type: "docker"
        docker:
          image: "redis:7-alpine"
          container: "app-redis"
          ports:
            - "6379:6379"
          restart_policy: "unless-stopped"
      
      # Backend service (traditional process)
      api:
        name: "API Server"
        type: "process"
        command: "python -m uvicorn main:app --reload"
        workdir: "${PROJECT_ROOT}/backend"
        port: 8000
        environment:
          DATABASE_URL: "postgres://app:secret@localhost:5432/appdb"
          REDIS_URL: "redis://localhost:6379"
          LOG_DIR: "${CWD}/logs"
      
      # Frontend in Docker
      frontend:
        name: "React App"
        type: "docker"
        docker:
          image: "node:18-alpine"
          container: "app-frontend"
          working_dir: "/app"
          volumes:
            - "${PROJECT_ROOT}/frontend:/app"
            - "/app/node_modules"
          ports:
            - "3000:3000"
          environment:
            REACT_APP_API_URL: "http://localhost:8000"
          entrypoint: "sh"
          cmd: "-c 'npm install && npm start'"
    
    modes:
      development:
        services: ["db", "cache", "api", "frontend"]
        environment:
          DEBUG: "true"
      
      production:
        services: ["db", "cache", "api", "frontend"]
        environment:
          DEBUG: "false"
          NODE_ENV: "production"
```

### Docker Service Lifecycle

When you start a Docker service in DevUp:

1. **Pull Image** (if `pull: true`) - Latest image is pulled from registry
2. **Create Container** - Container is created with specified configuration
3. **Start Container** - Container starts in detached mode
4. **Health Checks** - Optional health checks can verify service readiness
5. **Stop** - Container is stopped gracefully with 10-second timeout
6. **Remove** (if `remove: true`) - Container is removed after stopping

### Logs for Docker Services

Docker services log to files in the `logs/` directory, similar to process-based services:

```bash
# View logs for a Docker service
devup logs service-name

# Logs are stored in:
# ${CWD}/logs/service-name.log
```

### Health Checks with Docker

Docker services support the same health check configuration as process services:

```yaml
services:
  api:
    type: "docker"
    docker:
      image: "myapp:latest"
    healthcheck:
      type: "http"
      endpoint: "http://localhost:8000/health"
      timeout: 5s
      interval: 10s
      retries: 3
```

## Combining Features: Template Variables + Docker

The most powerful use case combines both features:

```yaml
apps:
  microservices:
    workdir: "${PROJECT_ROOT}"
    
    services:
      # Database path uses template variable
      mongodb:
        type: "docker"
        docker:
          image: "mongo:6"
          volumes:
            - "${HOME}/.devup/data/mongo:/data/db"
          ports:
            - "27017:27017"
      
      # Service logs to execution directory
      api:
        type: "docker"
        docker:
          image: "mycompany/api:latest"
          volumes:
            - "${PROJECT_ROOT}/api:/app"
          working_dir: "/app"
          environment:
            LOG_DIR: "${CWD}/logs"
            MONGO_URL: "mongodb://localhost:27017"
    
    environment:
      # Global variables resolved for all services
      APP_NAME: "${APP_NAME}"
      WORKSPACE: "${PROJECT_ROOT}"
      LOCAL_LOGS: "${CWD}/logs"
```

## Migration Guide

### Updating Existing Configurations

If you have existing DevUp configurations, you can modernize them:

**Before:**
```yaml
services:
  api:
    command: "/home/user/projects/myapp/api/run.sh"
    logfile: "logs/api.log"
```

**After:**
```yaml
services:
  api:
    command: "${PROJECT_ROOT}/api/run.sh"
    logfile: "${CWD}/logs/api.log"  # Always relative to execution dir
```

**Container-based Before:**
```yaml
# No Docker support - had to use wrapper scripts
```

**Container-based After:**
```yaml
services:
  postgres:
    type: "docker"
    docker:
      image: "postgres:15"
      ports: ["5432:5432"]
```

## Best Practices

1. **Use `${CWD}` for logs** - Ensures logs are always saved where you run devup
2. **Use `${PROJECT_ROOT}` for source code** - Makes configs portable across systems
3. **Use `${HOME}` for user-specific data** - Keeps configs clean
4. **Set `pull: true` for Docker** - In CI/CD, always get latest images
5. **Use volume mounts for development** - Mount source code into containers for live reloading
6. **Set appropriate restart policies** - Use `unless-stopped` for critical services

## Troubleshooting

### Docker Container Not Starting

```bash
# Check logs
devup logs service-name

# Check container status
docker ps -a | grep service-name

# Check Docker errors
docker logs container-name
```

### Template Variable Not Resolving

- Verify variable syntax: `${VAR_NAME}` (not `$VAR_NAME` or `${VAR_NAME}name`)
- For environment variables, ensure they're exported in your shell
- Built-in variables are always available: HOME, PROJECT_ROOT, WORKDIR, APP_NAME, CWD

### Port Conflicts

```yaml
# Use different ports if needed
docker:
  ports:
    - "5433:5432"  # External:Internal
```

## Related Documentation

- [Configuration Guide](./ARCHITECTURE.md)
- [Quick Start](./QUICKSTART.md)
- [Full Feature Summary](./FEATURE_SUMMARY.md)

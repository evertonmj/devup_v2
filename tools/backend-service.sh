#!/bin/bash
# ============================================================================
# Backend Service Management Script
# ============================================================================
#
# DESCRIPTION:
#   Manages the backend service for the Engineering Supervisor Agent application. 
#   This script handles backend-specific process lifecycle, Python environment 
#   setup, and different backend modes (mock, python with mocks, python real).
#
# WHAT IT MANAGES:
#   - Python FastAPI backend (port 8000)
#   - Mock backend server (port 8001)
#   - Python virtual environment setup
#   - Environment variable generation
#   - Backend-specific logging
#   - Backend health checks
#
# USAGE:
#   ./backend-service.sh start [--py] [--no-mocks] [--debug]  # Start backend
#   ./backend-service.sh stop                                 # Stop backend
#   ./backend-service.sh status                               # Check status
#   ./backend-service.sh health                               # Health checks
#   ./backend-service.sh get-command [options]                # Get tmux command
#   ./backend-service.sh setup-env                            # Setup environment
#
# MODES:
#   Default (no --py): Mock backend server (port 8001)
#   --py:              Python backend with mocked external services (port 8000)
#   --py --no-mocks:   Python backend with real connections (port 8000)
#   --py --debug:      Python backend with debug mode (port 8000)
#
# PORTS:
#   - 8000: Python backend API (when using --py)
#   - 8001: Mock server (when not using --py)
#
# ============================================================================

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CONTAINER_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
PROJECT_ROOT="$( cd "$CONTAINER_DIR/.." && pwd )"

# Backend-specific ports
BACKEND_PORTS=(8000 8001)

# Parse command line arguments
ACTION=${1:-start}
USE_PYTHON=false
USE_REAL_CONNECTIONS=false
DEBUG_MODE=false

for arg in "$@"; do
  case $arg in
    --py)
      USE_PYTHON=true
      ;;
    --no-mocks)
      USE_REAL_CONNECTIONS=true
      ;;
    --debug)
      DEBUG_MODE=true
      USE_PYTHON=true  # Debug mode implies Python backend
      ;;
  esac
done

# ============================================================================
# Environment Setup Functions
# ============================================================================

# Setup Python virtual environment and generate environment variables
setup_backend_environment() {
  echo "Setting up backend environment..."
  
  # Check and setup Python virtual environment if needed
  VENV_PATH="$CONTAINER_DIR/service/venv"
  if [ ! -d "$VENV_PATH" ]; then
    echo "⚠️  Virtual environment not found at $VENV_PATH"
    echo "Creating virtual environment and installing dependencies..."
    echo "This may take a few minutes on first run..."
    cd "$CONTAINER_DIR/service"
    python3.12 -m venv venv
    source venv/bin/activate
    pip install -q -r requirements.txt -r requirements-test.txt
    cd "$CONTAINER_DIR"
    echo "✓ Virtual environment setup complete!"
  elif [ ! -f "$VENV_PATH/bin/activate" ]; then
    echo "❌ Virtual environment exists but is corrupted (missing activate script)"
    echo "Please run: cd $CONTAINER_DIR && make install-service"
    exit 1
  else
    echo "✓ Virtual environment found at $VENV_PATH"
  fi

  # Activate the virtual environment
  echo "Activating Python virtual environment..."
  source "$CONTAINER_DIR/service/venv/bin/activate"

  # Generate DATABRICKS_TOKEN_ENCRYPTION_KEY
  export DATABRICKS_TOKEN_ENCRYPTION_KEY=$(python3 -c "from cryptography.fernet import Fernet; key=Fernet.generate_key(); print(key.decode())")
  echo "DATABRICKS_TOKEN_ENCRYPTION_KEY generated: $DATABRICKS_TOKEN_ENCRYPTION_KEY"

  # Generate JWT_SECRET_KEY
  export JWT_SECRET_KEY=$(openssl rand -base64 32)
  echo "JWT_SECRET_KEY generated: $JWT_SECRET_KEY"

  # Generate SAMPLE_DATABRICKS_ACCESS_TOKEN for mock service (JSON format)
  export SAMPLE_DATABRICKS_ACCESS_TOKEN='{"access_token":"mock-databricks-token-'$(openssl rand -hex 16)'","issued_token_type":"urn:ietf:params:oauth:token-type:access_token","scope":"all-apis","token_type":"Bearer","expires_in":3600}'
  echo "SAMPLE_DATABRICKS_ACCESS_TOKEN generated: $SAMPLE_DATABRICKS_ACCESS_TOKEN"
}

# ============================================================================
# Backend Service Functions
# ============================================================================

# Kill processes on backend ports
kill_backend_processes() {
  echo "Stopping backend processes..."
  for port in "${BACKEND_PORTS[@]}"; do
    pids=$(lsof -ti tcp:$port 2>/dev/null)
    if [ -n "$pids" ]; then
      echo "Killing processes on port $port: $pids"
      echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
  done
}

# Check backend service status
check_backend_status() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Backend Service Status Check"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  for port in "${BACKEND_PORTS[@]}"; do
    if lsof -ti tcp:$port > /dev/null 2>&1; then
      pid=$(lsof -ti tcp:$port)
      echo "✓ Port $port is in use (PID: $pid)"
    else
      echo "❌ Port $port is available (no process running)"
    fi
  done

  # Check virtual environment
  VENV_PATH="$CONTAINER_DIR/service/venv"
  if [ -d "$VENV_PATH" ] && [ -f "$VENV_PATH/bin/activate" ]; then
    echo "✓ Python virtual environment is available"
  else
    echo "❌ Python virtual environment is missing or corrupted"
  fi
}

# Run backend health checks
run_backend_health_checks() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Backend Health Checks"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  local all_healthy=true
  
  # Check Python backend (if using Python mode)
  if lsof -ti tcp:8000 > /dev/null 2>&1; then
    if curl -sk --max-time 2 https://localhost:8000 > /dev/null 2>&1; then
      echo "✓ Python backend (FastAPI) is responding on port 8000"
    else
      echo "❌ Python backend (FastAPI) is running but NOT responding on port 8000!"
      all_healthy=false
    fi
  fi
  
  # Check Mock server (if using mock mode)
  if lsof -ti tcp:8001 > /dev/null 2>&1; then
    if curl -sk --max-time 2 https://localhost:8001 > /dev/null 2>&1; then
      echo "✓ Mock backend server is responding on port 8001"
    else
      echo "❌ Mock backend server is running but NOT responding on port 8001!"
      all_healthy=false
    fi
  fi
  
  # If no backend is running
  if ! lsof -ti tcp:8000 > /dev/null 2>&1 && ! lsof -ti tcp:8001 > /dev/null 2>&1; then
    echo "❌ No backend service is running (neither port 8000 nor 8001)!"
    all_healthy=false
  fi
  
  if $all_healthy; then
    echo "✓ Backend services are healthy!"
    return 0
  else
    echo "❌ Some backend services have issues!"
    return 1
  fi
}

# Get the tmux command for backend pane
get_backend_command() {
  # Create logs directory if it doesn't exist
  mkdir -p "$CONTAINER_DIR/logs"
  
  # Prepare backend command based on mode
  if $USE_PYTHON; then
    if $DEBUG_MODE; then
      # Debug Python backend with enhanced logging and debug features
      echo "cd service && source venv/bin/activate && source ../.env 2>/dev/null || true && export PYTHONPATH=src:tests && export LOG_LEVEL=DEBUG && export ENABLE_DEBUG_ENDPOINTS=true && export UVICORN_LOG_LEVEL=debug && python3.12 -u tests/testbotofbots/localmock/main.py 2>&1 | tee -a ../logs/service.log; exec bash"
    elif $USE_REAL_CONNECTIONS; then
      # Real Python backend with actual connections
      echo "cd service && source venv/bin/activate && source ../.env 2>/dev/null || true && export PYTHONPATH=src && python3.12 src/botofbots/app.py 2>&1 | tee -a ../logs/service.log; exec bash"
    else
      # Python backend with mock/local setup
      echo "cd service && source venv/bin/activate && source ../.env 2>/dev/null || true && export PYTHONPATH=src:tests && python3.12 tests/testbotofbots/localmock/main.py 2>&1 | tee -a ../logs/service.log; exec bash"
    fi
  else
    # Mock backend
    echo "cd ui && npm run mock 2>&1 | tee -a ../logs/service.log; exec bash"
  fi
}

# Start backend services (this is used when running backend in isolation)
start_backend_services() {
  echo "Starting backend services..."
  
  # Setup environment first
  setup_backend_environment
  
  # Kill any existing processes
  kill_backend_processes
  
  # Create logs directory
  mkdir -p "$CONTAINER_DIR/logs"
  
  echo "Logs will be saved to:"
  echo "  - $CONTAINER_DIR/logs/service.log"
  
  # Start appropriate backend service
  if $USE_PYTHON; then
    cd "$CONTAINER_DIR/service"
    source venv/bin/activate
    source ../.env 2>/dev/null || true
    
    if $DEBUG_MODE; then
      echo "Starting Python backend in DEBUG mode..."
      export PYTHONPATH=src:tests
      export LOG_LEVEL=DEBUG
      export ENABLE_DEBUG_ENDPOINTS=true
      export UVICORN_LOG_LEVEL=debug
      python3.12 -u tests/testbotofbots/localmock/main.py > ../logs/service.log 2>&1 &
    elif $USE_REAL_CONNECTIONS; then
      echo "Starting Python backend with real connections..."
      export PYTHONPATH=src
      python3.12 src/botofbots/app.py > ../logs/service.log 2>&1 &
    else
      echo "Starting Python backend with mocked services..."
      export PYTHONPATH=src:tests
      python3.12 tests/testbotofbots/localmock/main.py > ../logs/service.log 2>&1 &
    fi
    BACKEND_PID=$!
    echo "Started Python backend (PID: $BACKEND_PID)"
  else
    echo "Starting mock backend server..."
    cd "$CONTAINER_DIR/ui"
    npm run mock > ../logs/service.log 2>&1 &
    BACKEND_PID=$!
    echo "Started mock backend (PID: $BACKEND_PID)"
  fi
  
  # Wait a moment for service to initialize
  echo "Waiting for backend service to initialize..."
  sleep 10
  
  # Run health check
  run_backend_health_checks
}

# ============================================================================
# Main Action Handler
# ============================================================================

case $ACTION in
  start)
    start_backend_services
    ;;
  stop)
    kill_backend_processes
    ;;
  status)
    check_backend_status
    ;;
  health)
    run_backend_health_checks
    ;;
  get-command)
    get_backend_command
    ;;
  setup-env)
    setup_backend_environment
    ;;
  *)
    echo "Usage: $0 {start|stop|status|health|get-command|setup-env} [--py] [--no-mocks] [--debug]"
    echo ""
    echo "Commands:"
    echo "  start       Start backend service"
    echo "  stop        Stop backend service"
    echo "  status      Check backend service status"
    echo "  health      Run backend health checks"
    echo "  get-command Get tmux command for backend pane"
    echo "  setup-env   Setup backend environment (venv + env vars)"
    echo ""
    echo "Backend Mode Options:"
    echo "  (none)      Mock backend server (port 8001)"
    echo "  --py        Python backend with mocked external services (port 8000)"
    echo "  --py --no-mocks  Python backend with real connections (port 8000)"
    echo "  --py --debug     Python backend with debug mode (port 8000)"
    exit 1
    ;;
esac
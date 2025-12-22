#!/bin/bash
# ============================================================================
# Startup Script for Local Development - Service Orchestrator
# ============================================================================
#
# DESCRIPTION:
#   Orchestrator script that manages the local development environment for the 
#   Engineering Supervisor Agent application. This script coordinates UI and 
#   backend services using dedicated service management scripts, handles tmux 
#   session management, and provides integrated health checks.
#
# WHAT IT DOES:
#   1. Delegates UI management to ui-service.sh (React + Caddy)
#   2. Delegates backend management to backend-service.sh (Python/Mock)
#   3. Orchestrates multi-pane tmux session setup
#   4. Provides integrated health checks across all services
#   5. Manages iTerm2 integration for seamless development experience
#   6. Coordinates service lifecycle (start/stop/restart)
#
# ARCHITECTURE:
#   startup.sh (orchestrator)
#   ├── ui-service.sh     → React (port 3000) + Caddy (port 3006)
#   └── backend-service.sh → Python (port 8000) OR Mock (port 8001)
#
# REQUIREMENTS:
#   - macOS (uses osascript for iTerm2 automation)
#   - iTerm2 installed (https://iterm2.com/)
#   - tmux installed: brew install tmux
#   - Caddy installed: brew install caddy
#   - Python 3.12 or higher
#   - Node.js and npm
#   - pip, cryptography, openssl
#   - Virtual environment at Container/service/venv (create with make install-service)
#   - All environment variables configured (see README.md)
#
# USAGE:
#   # From project root or Container directory:
#   ./Container/tools/startup.sh              # Mock backend (no external dependencies)
#   ./Container/tools/startup.sh --py         # Python backend with mocks
#   ./Container/tools/startup.sh --py --no-mocks  # Python backend with real connections
#   ./Container/tools/startup.sh --py --debug # Python backend with debug mode (verbose logging)
#   ./Container/tools/startup.sh --stop       # Stop all services and cleanup
#   ./Container/tools/startup.sh --restart    # Restart all services
#
#   # Using Make (recommended):
#   cd Container
#   make start          # Mock backend
#   make start-py       # Python backend with mocks
#   make start-py-real  # Python backend with real connections
#   make start-debug    # Python backend with debug mode
#   make stop           # Stop all services
#
# PORTS:
#   - 3000: React UI (direct access, dev mode with hot reload)
#   - 3006: Caddy reverse proxy (HTTPS access to UI and API)
#   - 8000: Python backend API
#   - 8001: Mock server (when running without --py)
#
# TROUBLESHOOTING:
#   - If ports are in use: Run ./Container/tools/startup.sh --stop first
#   - If tmux session exists: Run tmux kill-session -t startup
#   - If services don't start: Check logs in each tmux pane
#   - Path errors: Ensure script is run from project root or Container directory
#
# NOTES:
#   - The script uses absolute paths and is location-aware
#   - All services run in a single tmux session for easy management
#   - Use Ctrl+B then arrow keys to navigate between tmux panes
#   - Use Ctrl+B then D to detach from tmux (services keep running)
#   - Use tmux attach -t startup to reattach to the session
# ============================================================================

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Container directory is the parent of tools
CONTAINER_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
PROJECT_ROOT="$( cd "$CONTAINER_DIR/.." && pwd )"

# Service management scripts
UI_SERVICE_SCRIPT="$SCRIPT_DIR/ui-service.sh"
BACKEND_SERVICE_SCRIPT="$SCRIPT_DIR/backend-service.sh"

# Parse flags first to check for actions
do_py=false
do_real=false
do_stop=false
do_restart=false
do_debug=false

# Build backend and UI arguments
backend_args=""
ui_args=""

for arg in "$@"; do
  case $arg in
    --py)
      do_py=true
      backend_args="$backend_args --py"
      ;;
    --no-mocks)
      do_real=true
      backend_args="$backend_args --no-mocks"
      ;;
    --stop)
      do_stop=true
      ;;
    --restart)
      do_restart=true
      ;;
    --debug)
      do_debug=true
      do_py=true  # Debug mode implies Python backend
      backend_args="$backend_args --debug"
      ui_args="$ui_args --debug"
      ;;
  esac
done

# If --stop flag is provided, stop all services and exit
if $do_stop; then
  echo "Stopping all services..."

  # Use service scripts to stop services
  if [ -f "$UI_SERVICE_SCRIPT" ]; then
    "$UI_SERVICE_SCRIPT" stop
  else
    echo "Warning: UI service script not found, falling back to manual cleanup"
    # Fallback: Kill UI processes manually
    for port in 3000 3006; do
      pids=$(lsof -ti tcp:$port 2>/dev/null)
      if [ -n "$pids" ]; then
        echo "Killing processes on port $port: $pids"
        echo "$pids" | xargs kill -9 2>/dev/null || true
      fi
    done
  fi

  if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
    "$BACKEND_SERVICE_SCRIPT" stop
  else
    echo "Warning: Backend service script not found, falling back to manual cleanup"
    # Fallback: Kill backend processes manually
    for port in 8000 8001; do
      pids=$(lsof -ti tcp:$port 2>/dev/null)
      if [ -n "$pids" ]; then
        echo "Killing processes on port $port: $pids"
        echo "$pids" | xargs kill -9 2>/dev/null || true
      fi
    done
  fi

  # Kill tmux session
  if tmux has-session -t startup 2>/dev/null; then
    echo "Killing tmux session 'startup'"
    tmux kill-session -t startup
  fi
  
  # Close iTerm2 windows/tabs running the startup tmux session
  if command -v osascript >/dev/null 2>&1; then
    echo "Closing iTerm2 windows/tabs with startup session..."
    # Close any iTerm2 tabs/windows that have "tmux attach -t startup" or "startup" session
    osascript 2>/dev/null <<'EOF' || true
tell application "iTerm"
  repeat with myWindow in windows
    try
      repeat with myTab in tabs of myWindow
        try
          set tabText to ""
          repeat with mySession in sessions of myTab
            try
              set tabText to (text of mySession)
              -- Check if this tab contains our tmux session
              if tabText contains "tmux" or tabText contains "startup" then
                close myTab
                exit repeat
              end if
            end try
          end repeat
        end try
      end repeat
    end try
  end repeat
end tell
EOF
  fi

  echo "All services stopped."
  exit 0
fi

# Setup backend environment (this includes venv setup and env variable generation)
echo "Setting up backend environment..."
if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
  "$BACKEND_SERVICE_SCRIPT" setup-env
else
  echo "Warning: Backend service script not found, skipping environment setup"
fi

# Normal startup: stop existing services and session
if [ -f "$UI_SERVICE_SCRIPT" ]; then
  "$UI_SERVICE_SCRIPT" stop 2>/dev/null || true
else
  # Fallback: Kill UI processes manually
  for port in 3000 3006; do
    lsof -ti tcp:$port | xargs kill 2>/dev/null || true
  done
fi

if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
  "$BACKEND_SERVICE_SCRIPT" stop 2>/dev/null || true
else
  # Fallback: Kill backend processes manually
  for port in 8000 8001; do
    lsof -ti tcp:$port | xargs kill 2>/dev/null || true
  done
fi

# Kill any existing tmux session named 'startup'
tmux kill-session -t startup 2>/dev/null || true

# Create logs directory if it doesn't exist
mkdir -p "$CONTAINER_DIR/logs"
# Do NOT clear previous log files, always append

echo "Logs will be saved to:"
echo "  - $CONTAINER_DIR/logs/ui.log"
echo "  - $CONTAINER_DIR/logs/service.log"
echo "  - $CONTAINER_DIR/logs/caddy.log"

# Get commands from service scripts
if [ -f "$UI_SERVICE_SCRIPT" ]; then
  ui_cmd=$("$UI_SERVICE_SCRIPT" get-command $ui_args)
  caddy_cmd=$("$UI_SERVICE_SCRIPT" get-caddy-command)
else
  # Fallback UI command
  if $do_debug; then
    ui_cmd="cd ui && export VITE_LOG_LEVEL=debug && export DEBUG=* && npm run start 2>&1 | tee ../logs/ui.log; exec bash"
  else
    ui_cmd="cd ui && npm run start 2>&1 | tee ../logs/ui.log; exec bash"
  fi
  caddy_cmd="cd service && ../tools/start_caddy.sh 2>&1 | tee ../logs/caddy.log; exec bash"
fi

if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
  backend_cmd=$("$BACKEND_SERVICE_SCRIPT" get-command $backend_args)
else
  # Fallback backend command
  if $do_py; then
    if $do_debug; then
      backend_cmd="cd service && source venv/bin/activate && source ../.env.local 2>/dev/null || true && export PYTHONPATH=src:tests && export LOG_LEVEL=DEBUG && export ENABLE_DEBUG_ENDPOINTS=true && export UVICORN_LOG_LEVEL=debug && python3.12 -u tests/testbotofbots/localmock/main.py 2>&1 | tee -a ../logs/service.log; exec bash"
    elif $do_real; then
      backend_cmd="cd service && source venv/bin/activate && source ../.env.local 2>/dev/null || true && export PYTHONPATH=src && python3.12 src/botofbots/app.py 2>&1 | tee -a ../logs/service.log; exec bash"
    else
      backend_cmd="cd service && source venv/bin/activate && source ../.env.local 2>/dev/null || true && export PYTHONPATH=src:tests && python3.12 tests/testbotofbots/localmock/main.py 2>&1 | tee -a ../logs/service.log; exec bash"
    fi
  else
    backend_cmd="cd ui && npm run mock 2>&1 | tee -a ../logs/service.log; exec bash"
  fi
fi

# Restart all services if --restart flag is provided
if ${do_restart:-false}; then
  echo "Restarting all services..."

  # Use service scripts to stop services
  if [ -f "$UI_SERVICE_SCRIPT" ]; then
    "$UI_SERVICE_SCRIPT" stop
  fi

  if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
    "$BACKEND_SERVICE_SCRIPT" stop
  fi

  # Kill tmux session
  if tmux has-session -t startup 2>/dev/null; then
    echo "Killing tmux session 'startup'"
    tmux kill-session -t startup
  fi

  echo "Waiting for cleanup..."
  sleep 3

  # Remove the --restart flag and re-run with remaining arguments
  new_args=()
  for arg in "$@"; do
    if [[ "$arg" != "--restart" ]]; then
      new_args+=("$arg")
    fi
  done

  # Prevent infinite loop by checking if we're already in a restart
  if [[ "${STARTUP_RESTARTING:-}" == "1" ]]; then
    echo "ERROR: Restart loop detected. Exiting."
    exit 1
  fi

  export STARTUP_RESTARTING=1
  exec $0 "${new_args[@]}"
fi

# Build the complete tmux command on a single line
# Don't use pane targeting - just use session name and tmux will split the last active pane
TMUX_CMD="cd $CONTAINER_DIR && tmux new-session -d -s startup '$ui_cmd' && tmux split-window -h -t startup '$backend_cmd' && tmux split-window -v -t startup '$caddy_cmd' && tmux select-layout -t startup tiled && tmux attach -t startup"

# Execute using osascript

osascript -e "tell application \"iTerm\"" \
          -e "  activate" \
          -e "  set myterm to (create window with default profile)" \
          -e "  tell myterm" \
          -e "    tell current session" \
          -e "      write text \"$TMUX_CMD\"" \
          -e "    end tell" \
          -e "  end tell" \
          -e "end tell"

# Health check before opening browser - now uses service scripts
health_check() {
  local attempt=$1
  local max_attempts=$2

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Running integrated health checks on all services (attempt $attempt/$max_attempts)..."
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  local all_healthy=true

  # Run UI health checks
  if [ -f "$UI_SERVICE_SCRIPT" ]; then
    if ! "$UI_SERVICE_SCRIPT" health; then
      all_healthy=false
    fi
  else
    echo "Warning: UI service script not found, skipping UI health checks"
    all_healthy=false
  fi

  echo ""

  # Run backend health checks
  if [ -f "$BACKEND_SERVICE_SCRIPT" ]; then
    if ! "$BACKEND_SERVICE_SCRIPT" health; then
      all_healthy=false
    fi
  else
    echo "Warning: Backend service script not found, skipping backend health checks"
    all_healthy=false
  fi

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  if $all_healthy; then
    echo "✓ All services are healthy!"
    return 0
  else
    echo "❌ Some services have issues!"
    return 1
  fi
}

# Wait for services to initialize before first health check
echo "Waiting for services to start up..."
sleep 30

# Health check with retry logic (try 3 times total)
max_attempts=3
for attempt in $(seq 1 $max_attempts); do
  if health_check $attempt $max_attempts; then
    # All services healthy, break out of retry loop
    break
  else
    # Health check failed
    if [ $attempt -lt $max_attempts ]; then
      echo "Some services not ready yet, waiting 5 seconds before retry..."
      sleep 5
    else
      echo "⚠️  Warning: Some services failed health checks after $max_attempts attempts."
      echo "   Check the tmux panes for error details."
    fi
  fi
done

# Open the main app and the direct React dev server in the default browser (macOS only)
sleep 5
open "https://localhost:3000/chat"

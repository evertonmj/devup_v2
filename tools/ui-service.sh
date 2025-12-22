#!/bin/bash
# ============================================================================
# UI Service Management Script
# ============================================================================
#
# DESCRIPTION:
#   Manages the UI frontend service (React + Caddy proxy) for the Engineering 
#   Supervisor Agent application. This script handles UI-specific process 
#   lifecycle, port management, and health checks.
#
# WHAT IT MANAGES:
#   - React UI development server (port 3000)
#   - Caddy HTTPS reverse proxy (port 3006)
#   - UI-specific logging
#   - UI health checks
#
# USAGE:
#   ./ui-service.sh start [--debug]     # Start UI services
#   ./ui-service.sh stop                # Stop UI services
#   ./ui-service.sh status              # Check UI service status
#   ./ui-service.sh health              # Run UI health checks
#   ./ui-service.sh get-command [--debug]  # Get tmux command for UI pane
#
# PORTS:
#   - 3000: React UI (direct access, dev mode with hot reload)
#   - 3006: Caddy reverse proxy (HTTPS access to UI and API)
#
# ============================================================================

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CONTAINER_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
PROJECT_ROOT="$( cd "$CONTAINER_DIR/.." && pwd )"

# UI-specific ports
UI_PORTS=(3000 3006)

# Parse command line arguments
ACTION=${1:-start}
DEBUG_MODE=false

for arg in "$@"; do
  case $arg in
    --debug)
      DEBUG_MODE=true
      ;;
  esac
done

# ============================================================================
# UI Service Functions
# ============================================================================

# Kill processes on UI ports
kill_ui_processes() {
  echo "Stopping UI processes..."
  for port in "${UI_PORTS[@]}"; do
    pids=$(lsof -ti tcp:$port 2>/dev/null)
    if [ -n "$pids" ]; then
      echo "Killing processes on port $port: $pids"
      echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
  done
}

# Check UI service status
check_ui_status() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "UI Service Status Check"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  for port in "${UI_PORTS[@]}"; do
    if lsof -ti tcp:$port > /dev/null 2>&1; then
      pid=$(lsof -ti tcp:$port)
      echo "✓ Port $port is in use (PID: $pid)"
    else
      echo "❌ Port $port is available (no process running)"
    fi
  done
}

# Run UI health checks
run_ui_health_checks() {
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "UI Health Checks"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  
  local all_healthy=true
  
  # Check direct React dev server
  if curl -sk --max-time 2 http://localhost:3000 > /dev/null 2>&1; then
    echo "✓ React dev server is responding on port 3000"
  else
    echo "❌ React dev server is NOT responding on port 3000!"
    all_healthy=false
  fi
  
  if $all_healthy; then
    echo "✓ All UI services are healthy!"
    return 0
  else
    echo "❌ Some UI services have issues!"
    return 1
  fi
}

# Get the tmux command for UI pane (React + Caddy)
get_ui_command() {
  # Create logs directory if it doesn't exist
  mkdir -p "$CONTAINER_DIR/logs"
  
  if $DEBUG_MODE; then
    echo "cd ui && export VITE_LOG_LEVEL=debug && export DEBUG=* && npm run start 2>&1 | tee ../logs/ui.log; exec bash"
  else
    echo "cd ui && npm run start 2>&1 | tee ../logs/ui.log; exec bash"
  fi
}

# Get the tmux command for Caddy pane
get_caddy_command() {
  # Create logs directory if it doesn't exist
  mkdir -p "$CONTAINER_DIR/logs"
  
  echo "cd service && ../tools/start_caddy.sh 2>&1 | tee ../logs/caddy.log; exec bash"
}

# Start UI services (this is used when running UI in isolation)
start_ui_services() {
  echo "Starting UI services..."
  
  # Kill any existing processes
  kill_ui_processes
  
  # Create logs directory
  mkdir -p "$CONTAINER_DIR/logs"
  
  echo "Logs will be saved to:"
  echo "  - $CONTAINER_DIR/logs/ui.log"
  echo "  - $CONTAINER_DIR/logs/caddy.log"
  
  # Start React dev server in background
  cd "$CONTAINER_DIR/ui"
  if $DEBUG_MODE; then
    export VITE_LOG_LEVEL=debug
    export DEBUG=*
  fi
  npm run start > ../logs/ui.log 2>&1 &
  UI_PID=$!
  echo "Started React dev server (PID: $UI_PID)"
  
  # Start Caddy in background
  cd "$CONTAINER_DIR/service"
  ../tools/start_caddy.sh > ../logs/caddy.log 2>&1 &
  CADDY_PID=$!
  echo "Started Caddy proxy (PID: $CADDY_PID)"
  
  # Wait a moment for services to initialize
  echo "Waiting for UI services to initialize..."
  sleep 10
  
  # Run health check
  run_ui_health_checks
}

# ============================================================================
# Main Action Handler
# ============================================================================

case $ACTION in
  start)
    start_ui_services
    ;;
  stop)
    kill_ui_processes
    ;;
  status)
    check_ui_status
    ;;
  health)
    run_ui_health_checks
    ;;
  get-command)
    get_ui_command
    ;;
  get-caddy-command)
    get_caddy_command
    ;;
  *)
    echo "Usage: $0 {start|stop|status|health|get-command|get-caddy-command} [--debug]"
    echo ""
    echo "Commands:"
    echo "  start              Start UI services (React + Caddy)"
    echo "  stop               Stop UI services"
    echo "  status             Check UI service status"
    echo "  health             Run UI health checks"
    echo "  get-command        Get tmux command for UI pane"
    echo "  get-caddy-command  Get tmux command for Caddy pane"
    echo ""
    echo "Options:"
    echo "  --debug            Enable debug mode (enhanced logging)"
    exit 1
    ;;
esac
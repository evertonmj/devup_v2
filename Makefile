# Recreate all services/containers (stop, remove, rebuild, and start)
recreate-all:
	@echo "[+] Recreating all services/containers..."
	$(MAKE) down || true
	$(MAKE) clean || true
	$(MAKE) build
	$(MAKE) up
# ============================================================================
# Makefile for Engineering Supervisor Agent - Local Development
# ============================================================================
#
# This Makefile provides convenient commands to manage the local development
# environment for the Engineering Supervisor Agent application.
#
# QUICK START:
#   1. make check-tools                 # Verify all required tools are installed
#   2. make install                     # Install all dependencies
#   3. make start                       # Start all services
#   4. make stop                        # Stop all services
#
# PREREQUISITES - Install these tools first:
#
#   macOS Command Line Tools:
#     xcode-select --install
#
#   Homebrew (if not installed):
#     /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
#
#   Required Tools:
#     brew install python@3.12          # Python 3.12 or higher
#     brew install node                 # Node.js 18+ and npm
#     brew install tmux                 # Terminal multiplexer
#     brew install caddy                # Web server for reverse proxy
#     brew install openssl              # SSL/TLS toolkit
#
#   iTerm2 (Terminal emulator):
#     Download from https://iterm2.com/ or: brew install --cask iterm2
#
#   Verify installations:
#     make check-tools
#
# ENVIRONMENT VARIABLES:
#   For Python backend modes (make start-py or make start-py-real),
#   configure environment variables as described in README.md
#
# ============================================================================

.PHONY: help help-examples check-tools check-pipeline create-pr create-pr-apply create-pr-force deploy-dev deploy-dev-apply deploy-dev-force list-deploys start start-ui-mck start-svc-mck start-no-mocks start-debug stop dump-logs clean obliterate install-ui install-service install test-ui test-service lint-ui lint-service test lint status health validate-all

# Internal flags for apply and skip validation modes
_APPLY ?= 0
_SKIP_VALIDATION ?= 0
_SKIP_LINT ?= 0

# Load environment variables from .env file if it exists
-include .env
export

# Notification service defaults
GOOGLE_CHAT_ENABLED ?= false
GOOGLE_CHAT_WEBHOOKS ?=
GOOGLE_CHAT_NOTIFY_PR ?= true
GOOGLE_CHAT_NOTIFY_DEPLOY ?= true

TEAMS_ENABLED ?= false
TEAMS_WEBHOOKS ?=
TEAMS_NOTIFY_PR ?= true
TEAMS_NOTIFY_DEPLOY ?= true

NOTIFICATION_PR_PREFIX ?= 🔔 Pull Request Created
NOTIFICATION_DEPLOY_PREFIX ?= 🚀 Deployment Triggered
NOTIFICATION_INCLUDE_BRANCH ?= true
NOTIFICATION_INCLUDE_AUTHOR ?= true

# Default target - show help
help:
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║         Engineering Supervisor Agent - Development Commands               ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Usage: make <command> [options]"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📋 HELP & INFORMATION"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make help           Show this comprehensive help message"
	@echo "  make help-examples  Show detailed usage examples for all commands"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔍 PREREQUISITES & CHECKS"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make check-tools"
	@echo "      Verify all required development tools are installed and meet version"
	@echo "      requirements. Checks: Python 3.12+, Node.js 18+, npm, tmux, Caddy,"
	@echo "      OpenSSL, iTerm2, and Homebrew."
	@echo ""
	@echo "  make check-pipeline"
	@echo "      Check GitHub Actions pipeline status for the current repository."
	@echo "      Shows latest 5 workflow runs with status, branch, and links."
	@echo "      Requires: GitHub CLI (gh) or GITHUB_TOKEN environment variable"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ VALIDATION (Runs before PR/Deploy in -apply mode)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make validate-all"
	@echo "      Run comprehensive validation checks in sequence:"
	@echo "      1. UI linting (ESLint/Prettier)"
	@echo "      2. Service linting (flake8)"
	@echo "      3. UI tests (Vitest)"
	@echo "      4. Service tests (pytest)"
	@echo "      5. UI build (production build)"
	@echo "      Exits with error if ANY check fails."
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🚀 GITHUB WORKFLOW (PR Creation & Deployment)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make create-pr [DESC='description']"
	@echo "      DRY-RUN: Preview PR creation without making changes."
	@echo "      Optional: DESC='Your PR description'"
	@echo "      Shows: branch name, target, commit message"
	@echo ""
	@echo "  make create-pr-apply [DESC='description']"
	@echo "      CREATE PR: Runs full validation, then creates PR with confirmation."
	@echo "      Steps: validate-all → check environment → confirm (yes) → push → create PR"
	@echo "      Uses GitHub Copilot for description if DESC not provided."
	@echo "      ⚠️  Blocks if validation fails"
	@echo ""
	@echo "  make create-pr-force [DESC='description']"
	@echo "      CREATE PR: Bypass validation (NOT RECOMMENDED - requires 2FA)"
	@echo "      Steps: 2FA challenge → confirm twice → create PR"
	@echo "      Use only in emergencies when validation cannot pass."
	@echo "      🔐 Requires: typing 'yes' + random challenge word (no copy-paste)"
	@echo ""
	@echo "  make deploy-dev"
	@echo "      DRY-RUN: Preview deployment trigger without making changes."
	@echo "      Shows: current branch, PR number, deployment comment"
	@echo ""
	@echo "  make deploy-dev-apply"
	@echo "      DEPLOY: Runs full validation, then triggers deployment with confirmation."
	@echo "      Steps: validate-all → find PR → check running deploys → confirm (yes) → deploy"
	@echo "      Adds 'deploy --target=Container' comment to PR."
	@echo "      ⚠️  Blocks if validation fails or no PR exists"
	@echo ""
	@echo "  make deploy-dev-force"
	@echo "      DEPLOY: Bypass validation (NOT RECOMMENDED - requires 2FA)"
	@echo "      Steps: 2FA challenge → confirm twice → deploy"
	@echo "      Use only in emergencies when validation cannot pass."
	@echo "      🔐 Requires: typing 'yes' + random challenge word (no copy-paste)"
	@echo ""
	@echo "  make list-deploys"
	@echo "      List current in-progress and recently completed deployments."
	@echo "      Shows: workflow name, branch, status, runtime, URL"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔔 NOTIFICATIONS (Google Chat & Microsoft Teams)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  Automatic notifications to Google Chat and/or Microsoft Teams when:"
	@echo "    • Pull Requests are created (make create-pr-apply)"
	@echo "    • Deployments are triggered (make deploy-dev-apply)"
	@echo ""
	@echo "  Setup:"
	@echo "    1. Copy template: cp .env.example .env"
	@echo "    2. Configure webhooks in .env file"
	@echo "    3. Enable services: GOOGLE_CHAT_ENABLED=true and/or TEAMS_ENABLED=true"
	@echo ""
	@echo "  Features:"
	@echo "    • Support for both Google Chat and Microsoft Teams"
	@echo "    • Multiple channels per service (comma-separated webhooks)"
	@echo "    • Independent control over PR and deployment notifications"
	@echo "    • Customizable message prefixes and content"
	@echo "    • Disabled by default (opt-in)"
	@echo ""
	@echo "  Documentation: See .env.example for detailed setup instructions"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 INSTALLATION & SETUP"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make install"
	@echo "      Install all dependencies (UI + Service). Run this first!"
	@echo "      Equivalent to: make install-ui && make install-service"
	@echo ""
	@echo "  make install-ui"
	@echo "      Install UI dependencies only."
	@echo "      Runs: cd ui && npm install"
	@echo ""
	@echo "  make install-service"
	@echo "      Install Python service dependencies only."
	@echo "      Creates venv, installs requirements.txt and requirements-test.txt"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "▶️  RUNNING THE APPLICATION"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make start"
	@echo "      Start all services with MOCK backend (fastest, no config needed)"
	@echo "      Services: React UI (port 3000) + Mock backend (port 8001) + Caddy (port 3006)"
	@echo "      Access: https://localhost:3006"
	@echo "      Best for: UI development, quick testing"
	@echo ""
	@echo "  make start-py"
	@echo "      Start with Python backend using MOCKED external services"
	@echo "      Services: React UI (port 3000) + Python backend (port 8000) + Caddy (port 3006)"
	@echo "      Mocked: Databricks, AWS connections"
	@echo "      Best for: Backend development without external dependencies"
	@echo ""
	@echo "  make start-py-real"
	@echo "      Start with Python backend using REAL connections"
	@echo "      Services: React UI (port 3000) + Python backend (port 8000) + Caddy (port 3006)"
	@echo "      Requires: All environment variables configured (see README.md)"
	@echo "      Best for: Integration testing, production-like testing"
	@echo ""
	@echo "  make start-debug"
	@echo "      Start with Python backend in DEBUG mode (enhanced logging + debug features)"
	@echo "      Services: React UI (port 3000) + Python backend (port 8000) + Caddy (port 3006)"
	@echo "      Features: Verbose logging, debug endpoints, enhanced error details, auto-reload"
	@echo "      Best for: Development debugging, troubleshooting, detailed inspection"
	@echo ""
	@echo "  make stop"
	@echo "      Stop all running services and cleanup"
	@echo "      Kills processes on ports: 3000, 3006, 8000, 8001"
	@echo "      Terminates tmux session: 'startup'"
	@echo ""
	@echo "  make status"
	@echo "      Check status of running services"
	@echo "      Shows: port usage (3000, 3006, 8000, 8001), tmux session status"
	@echo ""
	@echo "  make health"
	@echo "      Comprehensive health check of all services"
	@echo "      Checks: HTTP endpoints, response times, process status, overall health"
	@echo "      Tests actual service availability with curl requests"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🧪 TESTING"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make test-ui"
	@echo "      Run UI tests with Vitest"
	@echo "      Runs: cd ui && npm test"
	@echo ""
	@echo "  make test-service"
	@echo "      Run Python service tests with pytest"
	@echo "      Runs: cd service && source venv/bin/activate && pytest -vv"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🎨 CODE QUALITY"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make lint-ui"
	@echo "      Lint UI code (ESLint/Prettier)"
	@echo "      Runs: cd ui && npm run lint"
	@echo ""
	@echo "  make lint-service"
	@echo "      Lint Python service code (flake8)"
	@echo "      Runs: cd service && source venv/bin/activate && flake8 src tests"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🧹 MAINTENANCE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make clean"
	@echo "      Remove all build artifacts and dependencies"
	@echo "      Removes: node_modules, venv, __pycache__, .pytest_cache, build dirs, logs"
	@echo "      Requires: Confirmation (type 'yes')"
	@echo "      ⚠️  You'll need to run 'make install' again after this"
	@echo ""
	@echo "  make obliterate"
	@echo "      💥 NUCLEAR OPTION: Delete EVERYTHING (use with extreme caution!)"
	@echo "      Removes: ALL dependencies, caches, temp files, logs, environment files"
	@echo "      Includes: npm cache, pip cache, .DS_Store, IDE files, coverage, etc."
	@echo "      Requires: Two-factor confirmation (type 'OBLITERATE' + random word)"
	@echo "      ⚠️  ⚠️  EXTREMELY DESTRUCTIVE - Creates backup of .env.local"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                           SAFETY FEATURES                                  ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  🛡️  Dry-run by default: Commands preview changes without executing"
	@echo "  ✅ Automatic validation: Tests/lints/builds run before PRs/deployments"
	@echo "  🔒 Confirmation prompts: Must type 'yes' for destructive actions"
	@echo "  🔐 Two-factor auth: Required to bypass validation (random word challenge)"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         REQUIRED TOOLS                                     ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  If not installed, run these commands:"
	@echo ""
	@echo "  brew install python@3.12    # Python 3.12 or higher"
	@echo "  brew install node           # Node.js 18 or higher"
	@echo "  brew install tmux           # Terminal multiplexer"
	@echo "  brew install caddy          # Web server for reverse proxy"
	@echo "  brew install openssl        # SSL/TLS toolkit"
	@echo "  brew install --cask iterm2  # Terminal emulator (macOS only)"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                            PORT REFERENCE                                  ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  3000 - React UI development server (hot reload enabled)"
	@echo "  3006 - Caddy HTTPS reverse proxy (👉 USE THIS: https://localhost:3006)"
	@echo "  8000 - Python backend API"
	@echo "  8001 - Mock backend server (only with 'make start')"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                          QUICK START GUIDE                                 ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  First time setup:"
	@echo "    1. make check-tools     # Verify prerequisites"
	@echo "    2. make install         # Install dependencies"
	@echo "    3. make start           # Start with mock backend"
	@echo "    4. Open https://localhost:3006"
	@echo ""
	@echo "  Daily development:"
	@echo "    make start              # Start services"
	@echo "    # ... do your work ..."
	@echo "    make stop               # Stop services"
	@echo ""
	@echo "  Before creating a PR:"
	@echo "    make validate-all       # Run all checks"
	@echo "    make create-pr-apply    # Create PR with validation"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                           TMUX SHORTCUTS                                   ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  Ctrl+B then Arrow Keys  - Navigate between panes"
	@echo "  Ctrl+B then D           - Detach from tmux (services keep running)"
	@echo "  Ctrl+B then [           - Scroll mode (press q to exit)"
	@echo "  tmux attach -t startup  - Reattach to tmux session"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                      MORE INFORMATION & EXAMPLES                           ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  make help-examples      # Show detailed usage examples"
	@echo "  See README.md           # Complete setup instructions"
	@echo "  See MAKEFILE.md         # Comprehensive command documentation"
	@echo ""
	@echo "════════════════════════════════════════════════════════════════════════════"

# Show detailed usage examples
help-examples:
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║              Engineering Supervisor Agent - Usage Examples                ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📝 EXAMPLE 1: First Time Setup"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Step 1: Navigate to Container directory"
	@echo "  cd Container"
	@echo ""
	@echo "  # Step 2: Check if all required tools are installed"
	@echo "  make check-tools"
	@echo ""
	@echo "  # Step 3: Install all dependencies (UI + Service)"
	@echo "  make install"
	@echo ""
	@echo "  # Step 4: Start the application with mock backend"
	@echo "  make start"
	@echo ""
	@echo "  # Step 5: Open in browser"
	@echo "  # Visit: https://localhost:3006"
	@echo ""
	@echo "  # Step 6: When done, stop services"
	@echo "  make stop"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔄 EXAMPLE 2: Daily Development Workflow"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Morning: Start services"
	@echo "  make start"
	@echo ""
	@echo "  # Work on your features..."
	@echo "  # (Services run in tmux, hot reload enabled)"
	@echo ""
	@echo "  # Check service status anytime"
	@echo "  make status"
	@echo ""
	@echo "  # Detach from tmux (services keep running)"
	@echo "  # Press: Ctrl+B then D"
	@echo ""
	@echo "  # Reattach later"
	@echo "  tmux attach -t startup"
	@echo ""
	@echo "  # Evening: Stop services"
	@echo "  make stop"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🧪 EXAMPLE 3: Running Tests and Validation"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Run all validation checks (lints, tests, build)"
	@echo "  make validate-all"
	@echo ""
	@echo "  # Or run checks individually:"
	@echo ""
	@echo "  # Run UI tests only"
	@echo "  make test-ui"
	@echo ""
	@echo "  # Run service tests only"
	@echo "  make test-service"
	@echo ""
	@echo "  # Lint UI code"
	@echo "  make lint-ui"
	@echo ""
	@echo "  # Lint service code"
	@echo "  make lint-service"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 EXAMPLE 4: Creating a Pull Request (Safe Workflow)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Step 1: Preview what would happen (dry-run)"
	@echo "  make create-pr"
	@echo ""
	@echo "  # Step 2: Run validation checks manually (optional)"
	@echo "  make validate-all"
	@echo ""
	@echo "  # Step 3: Create PR with automatic validation"
	@echo "  make create-pr-apply"
	@echo "  # This will:"
	@echo "  #  - Run all tests, lints, and builds"
	@echo "  #  - Ask for confirmation (type 'yes')"
	@echo "  #  - Push branch and create PR"
	@echo "  #  - Generate description with GitHub Copilot"
	@echo ""
	@echo "  # Alternative: Provide custom description"
	@echo "  make create-pr-apply DESC=\"Fix authentication bug in login flow\""
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🚀 EXAMPLE 5: Deploying to Dev Environment"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Step 1: Check current deployments"
	@echo "  make list-deploys"
	@echo ""
	@echo "  # Step 2: Preview deployment (dry-run)"
	@echo "  make deploy-dev"
	@echo ""
	@echo "  # Step 3: Deploy with automatic validation"
	@echo "  make deploy-dev-apply"
	@echo "  # This will:"
	@echo "  #  - Run all tests, lints, and builds"
	@echo "  #  - Find PR for current branch"
	@echo "  #  - Check for running deployments"
	@echo "  #  - Ask for confirmation (type 'yes')"
	@echo "  #  - Trigger deployment"
	@echo ""
	@echo "  # Step 4: Monitor deployment"
	@echo "  make list-deploys"
	@echo "  # Or watch in real-time:"
	@echo "  gh run watch"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "⚠️  EXAMPLE 6: Emergency Bypass (NOT RECOMMENDED)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Only use when validation absolutely cannot pass"
	@echo "  # and you have a critical emergency"
	@echo ""
	@echo "  # Create PR bypassing validation (requires 2FA)"
	@echo "  make create-pr-force"
	@echo "  # You will be asked:"
	@echo "  #  1. 'Are you absolutely sure? (yes/no)' - Type: yes"
	@echo "  #  2. 'Type the word: elephant' - Type the exact word (no copy-paste)"
	@echo ""
	@echo "  # Deploy bypassing validation (requires 2FA)"
	@echo "  make deploy-dev-force"
	@echo "  # Same 2FA process as above"
	@echo ""
	@echo "  ⚠️  WARNING: Only use -force commands in true emergencies!"
	@echo "  ⚠️  They skip all tests and quality checks!"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔧 EXAMPLE 7: Different Backend Modes"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Mode 1: Mock backend (fastest, no config needed)"
	@echo "  make start"
	@echo "  # Best for: UI development, quick testing"
	@echo "  # Uses: Mock server on port 8001"
	@echo ""
	@echo "  # Mode 2: Python backend with mocked external services"
	@echo "  make start-py"
	@echo "  # Best for: Backend development"
	@echo "  # Uses: Real Python backend, mocked Databricks/AWS"
	@echo ""
	@echo "  # Mode 3: Python backend with real connections"
	@echo "  make start-py-real"
	@echo "  # Best for: Integration testing"
	@echo "  # Requires: All environment variables configured"
	@echo "  # Uses: Real connections to Databricks and AWS"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔍 EXAMPLE 8: Checking GitHub Pipeline Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Check latest workflow runs"
	@echo "  make check-pipeline"
	@echo ""
	@echo "  # Shows:"
	@echo "  #  - Latest 5 workflow runs"
	@echo "  #  - Status: ✅ success, ❌ failure, ⏳ in progress"
	@echo "  #  - Branch name"
	@echo "  #  - Time since run"
	@echo "  #  - Direct links to GitHub Actions"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🧹 EXAMPLE 9: Cleaning Up and Fresh Start"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Stop services first"
	@echo "  make stop"
	@echo ""
	@echo "  # Remove all build artifacts and dependencies"
	@echo "  make clean"
	@echo ""
	@echo "  # Reinstall everything"
	@echo "  make install"
	@echo ""
	@echo "  # Start fresh"
	@echo "  make start"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "💡 EXAMPLE 10: Common Troubleshooting"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # Problem: Ports already in use"
	@echo "  make stop"
	@echo "  # If still issues, manually kill processes:"
	@echo "  lsof -ti tcp:3000 | xargs kill"
	@echo "  lsof -ti tcp:3006 | xargs kill"
	@echo ""
	@echo "  # Problem: tmux session already exists"
	@echo "  tmux kill-session -t startup"
	@echo "  make start"
	@echo ""
	@echo "  # Problem: Dependencies out of sync"
	@echo "  make clean"
	@echo "  make install"
	@echo ""
	@echo "  # Problem: Services won't start"
	@echo "  make check-tools    # Verify prerequisites"
	@echo "  make status         # Check current status"
	@echo ""
	@echo "  # Problem: Tests failing"
	@echo "  make test-ui        # Run UI tests individually"
	@echo "  make test-service   # Run service tests individually"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📚 EXAMPLE 11: Complete Feature Development Flow"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  # 1. Create feature branch"
	@echo "  git checkout -b feat/new-feature"
	@echo ""
	@echo "  # 2. Start development environment"
	@echo "  make start"
	@echo ""
	@echo "  # 3. Develop your feature..."
	@echo "  # (Edit code, hot reload automatically updates)"
	@echo ""
	@echo "  # 4. Run validation periodically"
	@echo "  make validate-all"
	@echo ""
	@echo "  # 5. Commit your changes"
	@echo "  git add ."
	@echo "  git commit -m \"feat: add new feature\""
	@echo ""
	@echo "  # 6. Create PR with validation"
	@echo "  make create-pr-apply"
	@echo ""
	@echo "  # 7. PR created! Wait for review..."
	@echo "  make check-pipeline    # Check CI status"
	@echo ""
	@echo "  # 8. After approval, deploy to dev"
	@echo "  make deploy-dev-apply"
	@echo ""
	@echo "  # 9. Monitor deployment"
	@echo "  make list-deploys"
	@echo ""
	@echo "  # 10. Stop services when done"
	@echo "  make stop"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                          HELPFUL TIPS                                      ║"
	@echo "╚════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  💡 Always use dry-run first to preview changes:"
	@echo "     make create-pr          (preview)"
	@echo "     make create-pr-apply    (execute)"
	@echo ""
	@echo "  💡 Validation runs automatically with -apply commands"
	@echo "     No need to run 'make validate-all' separately"
	@echo ""
	@echo "  💡 Use tmux effectively:"
	@echo "     - Detach: Ctrl+B then D"
	@echo "     - Reattach: tmux attach -t startup"
	@echo "     - Navigate panes: Ctrl+B then Arrow Keys"
	@echo ""
	@echo "  💡 Check status anytime:"
	@echo "     make status             (local services)"
	@echo "     make check-pipeline     (GitHub Actions)"
	@echo "     make list-deploys       (deployments)"
	@echo ""
	@echo "  💡 Access the app via Caddy (avoid CORS issues):"
	@echo "     https://localhost:3006   (✅ recommended)"
	@echo "     http://localhost:3000    (direct React, may have CORS)"
	@echo ""
	@echo "════════════════════════════════════════════════════════════════════════════"
	@echo ""
	@echo "For more information:"
	@echo "  - Run 'make help' for command reference"
	@echo "  - See README.md for setup instructions"
	@echo "  - See MAKEFILE.md for comprehensive documentation"
	@echo ""

# ============================================================================
# Helper Functions
# ============================================================================

# Send notifications to configured services (Google Chat and/or Microsoft Teams)
# Usage: $(call send-notifications,TYPE,TITLE,MESSAGE,URL)
# TYPE: "pr" or "deploy"
# Example: $(call send-notifications,pr,"New PR","PR #123 created","https://github.com/...")
define send-notifications
	@TYPE="$(1)"; \
	TITLE="$(2)"; \
	MESSAGE="$(3)"; \
	URL="$(4)"; \
	AUTHOR=$$(git config user.name 2>/dev/null || echo "Unknown"); \
	BRANCH=$$(git branch --show-current 2>/dev/null || echo "unknown"); \
	REPO=$$(git config --get remote.origin.url 2>/dev/null | sed 's/.*github.com[:/]\(.*\)\.git/\1/' | sed 's/.*github.com[:/]\(.*\)/\1/' || echo "unknown"); \
	if [ "$$TYPE" = "pr" ]; then \
		PREFIX="$(NOTIFICATION_PR_PREFIX)"; \
	else \
		PREFIX="$(NOTIFICATION_DEPLOY_PREFIX)"; \
	fi; \
	SENT_ANY=false; \
	if [ "$(GOOGLE_CHAT_ENABLED)" = "true" ]; then \
		if [ "$$TYPE" = "pr" ] && [ "$(GOOGLE_CHAT_NOTIFY_PR)" = "true" ] || [ "$$TYPE" = "deploy" ] && [ "$(GOOGLE_CHAT_NOTIFY_DEPLOY)" = "true" ]; then \
			if [ -z "$(GOOGLE_CHAT_WEBHOOKS)" ]; then \
				echo "⚠️  Google Chat enabled but no webhooks configured"; \
			else \
				GCHAT_JSON=$$(cat <<EOF \
{ \
  "cards": [ \
    { \
      "header": { \
        "title": "$$PREFIX", \
        "subtitle": "$$REPO" \
      }, \
      "sections": [ \
        { \
          "widgets": [ \
            { \
              "keyValue": { \
                "topLabel": "Title", \
                "content": "$$TITLE" \
              } \
            }, \
            { \
              "keyValue": { \
                "topLabel": "Message", \
                "content": "$$MESSAGE" \
              } \
            } \
EOF
); \
				if [ "$(NOTIFICATION_INCLUDE_BRANCH)" = "true" ]; then \
					GCHAT_JSON="$$GCHAT_JSON,$$(cat <<EOF \
            { \
              "keyValue": { \
                "topLabel": "Branch", \
                "content": "$$BRANCH", \
                "icon": "DESCRIPTION" \
              } \
            } \
EOF
)"; \
				fi; \
				if [ "$(NOTIFICATION_INCLUDE_AUTHOR)" = "true" ]; then \
					GCHAT_JSON="$$GCHAT_JSON,$$(cat <<EOF \
            { \
              "keyValue": { \
                "topLabel": "Author", \
                "content": "$$AUTHOR", \
                "icon": "PERSON" \
              } \
            } \
EOF
)"; \
				fi; \
				if [ -n "$$URL" ]; then \
					GCHAT_JSON="$$GCHAT_JSON,$$(cat <<EOF \
            { \
              "buttons": [ \
                { \
                  "textButton": { \
                    "text": "VIEW", \
                    "onClick": { \
                      "openLink": { \
                        "url": "$$URL" \
                      } \
                    } \
                  } \
                } \
              ] \
            } \
EOF
)"; \
				fi; \
				GCHAT_JSON="$$GCHAT_JSON ] } ] } ] }"; \
				echo ""; \
				echo "📤 Sending notification to Google Chat..."; \
				IFS=','; \
				WEBHOOK_COUNT=0; \
				SUCCESS_COUNT=0; \
				for WEBHOOK in $(GOOGLE_CHAT_WEBHOOKS); do \
					WEBHOOK_COUNT=$$((WEBHOOK_COUNT + 1)); \
					if curl -s -X POST "$$WEBHOOK" \
						-H "Content-Type: application/json" \
						-d "$$GCHAT_JSON" > /dev/null 2>&1; then \
						SUCCESS_COUNT=$$((SUCCESS_COUNT + 1)); \
					fi; \
				done; \
				if [ $$SUCCESS_COUNT -eq $$WEBHOOK_COUNT ]; then \
					echo "✅ Google Chat: Sent to $$SUCCESS_COUNT channel(s)"; \
					SENT_ANY=true; \
				elif [ $$SUCCESS_COUNT -gt 0 ]; then \
					echo "⚠️  Google Chat: Sent to $$SUCCESS_COUNT/$$WEBHOOK_COUNT channel(s)"; \
					SENT_ANY=true; \
				else \
					echo "❌ Google Chat: Failed to send to any channels"; \
				fi; \
			fi; \
		fi; \
	fi; \
	if [ "$(TEAMS_ENABLED)" = "true" ]; then \
		if [ "$$TYPE" = "pr" ] && [ "$(TEAMS_NOTIFY_PR)" = "true" ] || [ "$$TYPE" = "deploy" ] && [ "$(TEAMS_NOTIFY_DEPLOY)" = "true" ]; then \
			if [ -z "$(TEAMS_WEBHOOKS)" ]; then \
				echo "⚠️  Microsoft Teams enabled but no webhooks configured"; \
			else \
				TEAMS_FACTS=""; \
				if [ "$(NOTIFICATION_INCLUDE_BRANCH)" = "true" ]; then \
					TEAMS_FACTS="$$TEAMS_FACTS{\"name\":\"Branch\",\"value\":\"$$BRANCH\"},"; \
				fi; \
				if [ "$(NOTIFICATION_INCLUDE_AUTHOR)" = "true" ]; then \
					TEAMS_FACTS="$$TEAMS_FACTS{\"name\":\"Author\",\"value\":\"$$AUTHOR\"},"; \
				fi; \
				TEAMS_FACTS=$${TEAMS_FACTS%,}; \
				TEAMS_JSON=$$(cat <<EOF \
{ \
  "@type": "MessageCard", \
  "@context": "https://schema.org/extensions", \
  "summary": "$$PREFIX", \
  "themeColor": "0078D7", \
  "title": "$$PREFIX", \
  "sections": [ \
    { \
      "activityTitle": "$$TITLE", \
      "activitySubtitle": "$$REPO", \
      "text": "$$MESSAGE", \
      "facts": [$$TEAMS_FACTS] \
    } \
  ] \
EOF
); \
				if [ -n "$$URL" ]; then \
					TEAMS_JSON="$$TEAMS_JSON,$$(cat <<EOF \
  "potentialAction": [ \
    { \
      "@type": "OpenUri", \
      "name": "View", \
      "targets": [ \
        { \
          "os": "default", \
          "uri": "$$URL" \
        } \
      ] \
    } \
  ] \
EOF
)"; \
				fi; \
				TEAMS_JSON="$$TEAMS_JSON }"; \
				echo "📤 Sending notification to Microsoft Teams..."; \
				IFS=','; \
				WEBHOOK_COUNT=0; \
				SUCCESS_COUNT=0; \
				for WEBHOOK in $(TEAMS_WEBHOOKS); do \
					WEBHOOK_COUNT=$$((WEBHOOK_COUNT + 1)); \
					if curl -s -X POST "$$WEBHOOK" \
						-H "Content-Type: application/json" \
						-d "$$TEAMS_JSON" > /dev/null 2>&1; then \
						SUCCESS_COUNT=$$((SUCCESS_COUNT + 1)); \
					fi; \
				done; \
				if [ $$SUCCESS_COUNT -eq $$WEBHOOK_COUNT ]; then \
					echo "✅ Microsoft Teams: Sent to $$SUCCESS_COUNT channel(s)"; \
					SENT_ANY=true; \
				elif [ $$SUCCESS_COUNT -gt 0 ]; then \
					echo "⚠️  Microsoft Teams: Sent to $$SUCCESS_COUNT/$$WEBHOOK_COUNT channel(s)"; \
					SENT_ANY=true; \
				else \
					echo "❌ Microsoft Teams: Failed to send to any channels"; \
				fi; \
			fi; \
		fi; \
	fi; \
	if [ "$$SENT_ANY" = "true" ]; then \
		echo ""; \
	fi
endef

# ============================================================================
# Prerequisites Check
# ============================================================================

# Check if all required tools are installed
check-tools:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Checking Required Tools Installation"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Checking Python..."
	@if command -v python3 >/dev/null 2>&1; then \
		PYTHON_VERSION=$$(python3 --version 2>&1 | awk '{print $$2}'); \
		MAJOR=$$(echo $$PYTHON_VERSION | cut -d. -f1); \
		MINOR=$$(echo $$PYTHON_VERSION | cut -d. -f2); \
		if [ $$MAJOR -ge 3 ] && [ $$MINOR -ge 12 ]; then \
			echo "  ✓ Python $$PYTHON_VERSION (OK)"; \
		else \
			echo "  ✗ Python $$PYTHON_VERSION (Need 3.12+)"; \
			echo "    Install: brew install python@3.12"; \
		fi \
	else \
		echo "  ✗ Python not found"; \
		echo "    Install: brew install python@3.12"; \
	fi
	@echo ""
	@echo "Checking Node.js..."
	@if command -v node >/dev/null 2>&1; then \
		NODE_VERSION=$$(node --version | sed 's/v//'); \
		MAJOR=$$(echo $$NODE_VERSION | cut -d. -f1); \
		if [ $$MAJOR -ge 18 ]; then \
			echo "  ✓ Node.js $$NODE_VERSION (OK)"; \
		else \
			echo "  ✗ Node.js $$NODE_VERSION (Need 18+)"; \
			echo "    Install: brew install node"; \
		fi \
	else \
		echo "  ✗ Node.js not found"; \
		echo "    Install: brew install node"; \
	fi
	@echo ""
	@echo "Checking npm..."
	@if command -v npm >/dev/null 2>&1; then \
		NPM_VERSION=$$(npm --version); \
		echo "  ✓ npm $$NPM_VERSION (OK)"; \
	else \
		echo "  ✗ npm not found (comes with Node.js)"; \
		echo "    Install: brew install node"; \
	fi
	@echo ""
	@echo "Checking tmux..."
	@if command -v tmux >/dev/null 2>&1; then \
		TMUX_VERSION=$$(tmux -V | awk '{print $$2}'); \
		echo "  ✓ tmux $$TMUX_VERSION (OK)"; \
	else \
		echo "  ✗ tmux not found"; \
		echo "    Install: brew install tmux"; \
	fi
	@echo ""
	@echo "Checking Caddy..."
	@if command -v caddy >/dev/null 2>&1; then \
		CADDY_VERSION=$$(caddy version 2>&1 | head -1 | awk '{print $$1}'); \
		echo "  ✓ Caddy $$CADDY_VERSION (OK)"; \
	else \
		echo "  ✗ Caddy not found"; \
		echo "    Install: brew install caddy"; \
	fi
	@echo ""
	@echo "Checking OpenSSL..."
	@if command -v openssl >/dev/null 2>&1; then \
		OPENSSL_VERSION=$$(openssl version | awk '{print $$2}'); \
		echo "  ✓ OpenSSL $$OPENSSL_VERSION (OK)"; \
	else \
		echo "  ✗ OpenSSL not found"; \
		echo "    Install: brew install openssl"; \
	fi
	@echo ""
	@echo "Checking iTerm2..."
	@if [ -d "/Applications/iTerm.app" ]; then \
		echo "  ✓ iTerm2 installed (OK)"; \
	else \
		echo "  ✗ iTerm2 not found"; \
		echo "    Install: brew install --cask iterm2"; \
		echo "    Or download from: https://iterm2.com/"; \
	fi
	@echo ""
	@echo "Checking Homebrew..."
	@if command -v brew >/dev/null 2>&1; then \
		BREW_VERSION=$$(brew --version | head -1 | awk '{print $$2}'); \
		echo "  ✓ Homebrew $$BREW_VERSION (OK)"; \
	else \
		echo "  ✗ Homebrew not found"; \
		echo "    Install: /bin/bash -c \"\$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""; \
	fi
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "For installation commands, see the header of this Makefile or run 'make help'"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check GitHub Actions pipeline status
check-pipeline:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "GitHub Actions Pipeline Status Check"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@# Check authentication - support both gh CLI and GITHUB_TOKEN environment variable
	@if [ -n "$$GITHUB_TOKEN" ]; then \
		echo "🔑 Using GITHUB_TOKEN from environment variable"; \
		echo ""; \
	elif command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then \
		echo "🔑 Using GitHub CLI authentication"; \
		echo ""; \
	elif command -v gh >/dev/null 2>&1; then \
		echo "❌ GitHub CLI is installed but not authenticated."; \
		echo ""; \
		echo "Choose one authentication method:"; \
		echo ""; \
		echo "Option 1: Authenticate with GitHub CLI (recommended):"; \
		echo "  gh auth login"; \
		echo ""; \
		echo "Option 2: Set GITHUB_TOKEN environment variable:"; \
		echo "  export GITHUB_TOKEN='your_github_personal_access_token'"; \
		echo ""; \
		echo "To create a token:"; \
		echo "  1. Go to https://github.com/settings/tokens"; \
		echo "  2. Click 'Generate new token (classic)'"; \
		echo "  3. Select scopes: 'repo' and 'workflow'"; \
		echo "  4. Copy the token and set it as GITHUB_TOKEN"; \
		echo ""; \
		exit 1; \
	else \
		echo "❌ GitHub CLI is not installed."; \
		echo ""; \
		echo "Install it with:"; \
		echo "  brew install gh"; \
		echo ""; \
		echo "Then authenticate:"; \
		echo "  gh auth login"; \
		echo ""; \
		echo "Or set GITHUB_TOKEN environment variable:"; \
		echo "  export GITHUB_TOKEN='your_github_personal_access_token'"; \
		echo ""; \
		exit 1; \
	fi
	@# Get repository information
	@if [ -d ../.git ]; then \
		echo "📦 Repository: $$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/' | sed 's/.*github.com[:/]\(.*\)/\1/')"; \
		echo "🌿 Current Branch: $$(git branch --show-current)"; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "Currently Running Workflows"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		RUNNING_WORKFLOWS=$$(if [ -n "$$GITHUB_TOKEN" ]; then \
			cd .. && GH_TOKEN=$$GITHUB_TOKEN gh run list --status in_progress --json databaseId,name,status,headBranch,createdAt,url \
				--template '{{range .}}🏃 {{.name}} ({{.headBranch}}){{"\n"}}   Status: {{.status}} | {{timeago .createdAt}}{{"\n"}}   {{.url}}{{"\n\n"}}{{end}}'; \
		else \
			cd .. && gh run list --status in_progress --json databaseId,name,status,headBranch,createdAt,url \
				--template '{{range .}}🏃 {{.name}} ({{.headBranch}}){{"\n"}}   Status: {{.status}} | {{timeago .createdAt}}{{"\n"}}   {{.url}}{{"\n\n"}}{{end}}'; \
		fi); \
		if [ -z "$$RUNNING_WORKFLOWS" ]; then \
			echo "   No workflows currently running"; \
			echo ""; \
		else \
			echo "$$RUNNING_WORKFLOWS"; \
		fi; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "Latest Workflow Runs"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		if [ -n "$$GITHUB_TOKEN" ]; then \
			cd .. && GH_TOKEN=$$GITHUB_TOKEN gh run list --limit 5 --json databaseId,name,status,conclusion,headBranch,createdAt,url \
				--template '{{range .}}{{if eq .status "completed"}}{{if eq .conclusion "success"}}✅{{else if eq .conclusion "failure"}}❌{{else if eq .conclusion "cancelled"}}⏸️{{else}}⚠️{{end}}{{else}}⏳{{end}} {{.name}} ({{.headBranch}}){{"\n"}}   Status: {{if eq .status "completed"}}{{.conclusion}}{{else}}{{.status}}{{end}} | {{timeago .createdAt}}{{"\n"}}   {{.url}}{{"\n\n"}}{{end}}'; \
		else \
			cd .. && gh run list --limit 5 --json databaseId,name,status,conclusion,headBranch,createdAt,url \
				--template '{{range .}}{{if eq .status "completed"}}{{if eq .conclusion "success"}}✅{{else if eq .conclusion "failure"}}❌{{else if eq .conclusion "cancelled"}}⏸️{{else}}⚠️{{end}}{{else}}⏳{{end}} {{.name}} ({{.headBranch}}){{"\n"}}   Status: {{if eq .status "completed"}}{{.conclusion}}{{else}}{{.status}}{{end}} | {{timeago .createdAt}}{{"\n"}}   {{.url}}{{"\n\n"}}{{end}}'; \
		fi; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "💡 Tips:"; \
		echo "   - View all runs: gh run list"; \
		echo "   - Watch a run: gh run watch"; \
		echo "   - View run details: gh run view <run-id>"; \
		echo "   - Re-run failed jobs: gh run rerun <run-id>"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	else \
		echo "❌ Not in a git repository."; \
		echo ""; \
		echo "Run this command from the project root or Container directory."; \
		exit 1; \
	fi

# ============================================================================
# Validation Commands
# ============================================================================

# Run all validation checks (tests, builds, lints)
validate: validate-all

validate-all:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running All Validation Checks"
	@if [ "$(_SKIP_LINT)" = "1" ]; then \
		echo "(Linting skipped)"; \
	fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@VALIDATION_FAILED=0; \
	STEP=1; \
	if [ "$(_SKIP_LINT)" != "1" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "1️⃣  Running UI Linting..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		if $(MAKE) lint-ui 2>&1; then \
			echo ""; \
			echo "✅ UI linting passed"; \
		else \
			echo ""; \
			echo "❌ UI linting failed"; \
			VALIDATION_FAILED=1; \
		fi; \
		echo ""; \
		STEP=2; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "2️⃣  Running Service Linting..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		if $(MAKE) lint-service 2>&1; then \
			echo ""; \
			echo "✅ Service linting passed"; \
		else \
			echo ""; \
			echo "❌ Service linting failed"; \
			VALIDATION_FAILED=1; \
		fi; \
		echo ""; \
		STEP=3; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ "$(_SKIP_LINT)" = "1" ]; then \
		echo "1️⃣  Running UI Tests..."; \
	else \
		echo "$${STEP}️⃣  Running UI Tests..."; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if $(MAKE) test-ui 2>&1; then \
		echo ""; \
		echo "✅ UI tests passed"; \
	else \
		echo ""; \
		echo "❌ UI tests failed"; \
		VALIDATION_FAILED=1; \
	fi; \
	echo ""; \
	STEP=$$((STEP + 1)); \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ "$(_SKIP_LINT)" = "1" ]; then \
		echo "2️⃣  Running Service Tests..."; \
	else \
		echo "$${STEP}️⃣  Running Service Tests..."; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if $(MAKE) test-service 2>&1; then \
		echo ""; \
		echo "✅ Service tests passed"; \
	else \
		echo ""; \
		echo "❌ Service tests failed"; \
		VALIDATION_FAILED=1; \
	fi; \
	echo ""; \
	STEP=$$((STEP + 1)); \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ "$(_SKIP_LINT)" = "1" ]; then \
		echo "3️⃣  Building UI..."; \
	else \
		echo "$${STEP}️⃣  Building UI..."; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if cd ui && npm run build 2>&1; then \
		echo ""; \
		echo "✅ UI build passed"; \
	else \
		echo ""; \
		echo "❌ UI build failed"; \
		VALIDATION_FAILED=1; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ $$VALIDATION_FAILED -eq 0 ]; then \
		echo "✅ ALL VALIDATION CHECKS PASSED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		exit 0; \
	else \
		echo "❌ VALIDATION FAILED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "Please fix the issues above before creating a PR or deploying."; \
		echo ""; \
		echo "To bypass validation (NOT RECOMMENDED), use:"; \
		echo "  make create-pr APPLY=1 SKIP_VALIDATION=1"; \
		echo "  make deploy-dev APPLY=1 SKIP_VALIDATION=1"; \
		echo ""; \
		exit 1; \
	fi

# ============================================================================
# GitHub Workflow Commands
# ============================================================================

# Wrapper targets for create-pr with different modes
create-pr-apply:
	@$(MAKE) create-pr _APPLY=1

create-pr-force:
	@$(MAKE) create-pr _APPLY=1 _SKIP_VALIDATION=1

# Create a Pull Request with optional description
# Usage:
#   make create-pr                  # Dry-run mode (default)
#   make create-pr-apply            # Actual execution with validation
#   make create-pr-force            # Skip validation (requires 2FA)
#   Optional: DESC="Your PR description"
# Without DESC, uses GitHub Copilot to generate description
create-pr:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ "$(_APPLY)" = "1" ]; then \
		echo "Creating Pull Request (APPLY MODE)"; \
	else \
		echo "Creating Pull Request (DRY-RUN MODE)"; \
	fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@# Check if gh CLI is installed and authenticated
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "❌ GitHub CLI (gh) is required."; \
		echo "Install: brew install gh"; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ] && ! gh auth status >/dev/null 2>&1; then \
		echo "❌ Not authenticated with GitHub."; \
		echo "Run: gh auth login"; \
		exit 1; \
	fi
	@# Check if in a git repository
	@if ! git rev-parse --git-dir >/dev/null 2>&1; then \
		echo "❌ Not in a git repository."; \
		exit 1; \
	fi
	@# Check if there are changes to commit
	@if ! git diff-index --quiet HEAD -- 2>/dev/null; then \
		echo "⚠️  Warning: You have uncommitted changes."; \
		echo ""; \
		read -p "Do you want to commit them? (y/n) " -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			git add -A; \
			read -p "Enter commit message: " commit_msg; \
			git commit -m "$$commit_msg"; \
		else \
			echo "❌ Please commit or stash your changes first."; \
			exit 1; \
		fi \
	fi
	@# Get current branch
	@CURRENT_BRANCH=$$(git branch --show-current); \
	echo "📍 Current branch: $$CURRENT_BRANCH"; \
	echo ""; \
	if [ "$$CURRENT_BRANCH" = "main" ] || [ "$$CURRENT_BRANCH" = "master" ]; then \
		echo "❌ Cannot create PR from main/master branch."; \
		echo "Create a feature branch first: git checkout -b feature/your-branch"; \
		exit 1; \
	fi; \
	if [ "$(_APPLY)" = "1" ] && [ "$(_SKIP_VALIDATION)" = "1" ]; then \
		echo "⚠️  WARNING: You are about to bypass validation checks!"; \
		echo ""; \
		echo "This is NOT recommended and should only be used in emergencies."; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "🔐 TWO-FACTOR AUTHENTICATION REQUIRED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		read -p "First confirmation - Are you absolutely sure? (yes/no) " -r CONFIRM1; \
		echo ""; \
		if [ "$$CONFIRM1" != "yes" ]; then \
			echo "❌ Operation cancelled."; \
			exit 1; \
		fi; \
		CHALLENGE_WORDS=("elephant" "volcano" "butterfly" "telescope" "saxophone" "submarine" "hurricane" "avalanche" "labyrinth" "astronaut" "flamingo" "cathedral" "pineapple" "octopus" "marshmallow"); \
		RANDOM_INDEX=$$(($$RANDOM % 15)); \
		CHALLENGE_WORD=$${CHALLENGE_WORDS[$$RANDOM_INDEX]}; \
		echo "🔐 Second confirmation - Type the following word EXACTLY (no copy-paste):"; \
		echo ""; \
		echo "   >>> $$CHALLENGE_WORD <<<"; \
		echo ""; \
		stty -echo 2>/dev/null || true; \
		trap 'stty echo 2>/dev/null || true' EXIT; \
		read -p "Type the word: " USER_INPUT; \
		stty echo 2>/dev/null || true; \
		trap - EXIT; \
		echo ""; \
		echo ""; \
		if [ "$$USER_INPUT" != "$$CHALLENGE_WORD" ]; then \
			echo "❌ Authentication failed. Word mismatch."; \
			echo "   Expected: $$CHALLENGE_WORD"; \
			echo "   Got: $$USER_INPUT"; \
			exit 1; \
		fi; \
		echo "✅ Two-factor authentication successful."; \
		echo ""; \
		echo "⚠️  Bypassing validation checks..."; \
		echo ""; \
	elif [ "$(_APPLY)" = "1" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "🔍 Running Validation Checks..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		if ! $(MAKE) validate-all; then \
			echo ""; \
			echo "❌ Validation failed. Cannot create PR."; \
			echo ""; \
			echo "Please fix the issues and try again, or use 'make create-pr-force' to bypass (not recommended)."; \
			exit 1; \
		fi; \
		echo ""; \
	fi; \
	if [ "$(_APPLY)" != "1" ]; then \
		echo "🔍 DRY-RUN: Would push branch to remote: $$CURRENT_BRANCH"; \
		echo ""; \
		if [ -n "$(DESC)" ]; then \
			echo "📝 DRY-RUN: Would use provided description"; \
		else \
			echo "🤖 DRY-RUN: Would generate PR description with GitHub Copilot"; \
		fi; \
		echo ""; \
		COMMIT_MSG=$$(git log -1 --pretty=%s); \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📋 DRY-RUN Summary:"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "  Branch:  $$CURRENT_BRANCH"; \
		echo "  Target:  main"; \
		echo "  Title:   $$COMMIT_MSG"; \
		echo ""; \
		echo "🔄 DRY-RUN: Would create Pull Request"; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "✅ DRY-RUN completed. No changes were made."; \
		echo ""; \
		echo "💡 To actually create the PR, run:"; \
		echo "   make create-pr-apply"; \
		if [ -n "$(DESC)" ]; then \
			echo "   make create-pr-apply DESC=\"$(DESC)\""; \
		fi; \
	else \
		echo "⚠️  APPLY MODE: This will actually create a Pull Request."; \
		echo ""; \
		read -p "Are you sure you want to continue? (yes/no) " -r CONFIRM; \
		echo ""; \
		if [ "$$CONFIRM" != "yes" ]; then \
			echo "❌ Operation cancelled."; \
			exit 1; \
		fi; \
		echo "🔄 Pushing branch to remote..."; \
		git push -u origin $$CURRENT_BRANCH 2>&1 || { \
			echo "❌ Failed to push branch."; \
			exit 1; \
		}; \
		echo ""; \
		if [ -n "$(DESC)" ]; then \
			echo "📝 Using provided description..."; \
			PR_BODY="$(DESC)"; \
		else \
			echo "🤖 Generating PR description with GitHub Copilot..."; \
			echo "   (This may take a few seconds)"; \
			echo ""; \
			if command -v gh >/dev/null 2>&1 && gh copilot --version >/dev/null 2>&1; then \
				PR_BODY=$$(cd .. && gh copilot suggest -t shell "generate a comprehensive PR description based on the git diff and commits in this branch" 2>/dev/null | tail -n +3 || echo ""); \
				if [ -z "$$PR_BODY" ]; then \
					echo "⚠️  Copilot unavailable, using default description."; \
					PR_BODY="## Changes\n\nAutomated PR created from branch $$CURRENT_BRANCH.\n\n## Testing\n- [ ] Manual testing completed\n- [ ] All tests passing"; \
				fi \
			else \
				echo "⚠️  GitHub Copilot CLI not available. Using default description."; \
				echo "   Install with: gh extension install github/gh-copilot"; \
				PR_BODY="## Changes\n\nAutomated PR created from branch $$CURRENT_BRANCH.\n\n## Testing\n- [ ] Manual testing completed\n- [ ] All tests passing"; \
			fi \
		fi; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "Creating PR..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		PR_URL=$$(cd .. && gh pr create --base main --head $$CURRENT_BRANCH --title "$$(git log -1 --pretty=%s)" --body "$$PR_BODY" 2>&1 | grep -o 'https://github.com[^[:space:]]*' || echo ""); \
		if [ -n "$$PR_URL" ]; then \
			echo "✅ Pull Request created successfully!"; \
			echo ""; \
			echo "🔗 PR URL: $$PR_URL"; \
			echo ""; \
			$(call send-notifications,pr,"$$(git log -1 --pretty=%s)","Pull request created successfully from branch $$CURRENT_BRANCH","$$PR_URL"); \
			echo "💡 Next steps:"; \
			echo "   - Review the PR: $$PR_URL"; \
			echo "   - Deploy to dev: make deploy-dev-apply"; \
			echo "   - Check deployments: make list-deploys"; \
		else \
			echo "❌ Failed to create PR."; \
			echo ""; \
			echo "This could mean:"; \
			echo "  - A PR already exists for this branch"; \
			echo "  - Check with: cd .. && gh pr list --head $$CURRENT_BRANCH"; \
			exit 1; \
		fi \
	fi

# Wrapper targets for deploy-dev with different modes
deploy-dev-apply:
	@$(MAKE) deploy-dev _APPLY=1

deploy-dev-force:
	@$(MAKE) deploy-dev _APPLY=1 _SKIP_VALIDATION=1

# Deploy to dev environment via PR comment
# Usage:
#   make deploy-dev          # Dry-run mode (default)
#   make deploy-dev-apply    # Actual execution with validation
#   make deploy-dev-force    # Skip validation (requires 2FA)
deploy-dev:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ "$(_APPLY)" = "1" ]; then \
		echo "Deploy to Dev Environment (APPLY MODE)"; \
	else \
		echo "Deploy to Dev Environment (DRY-RUN MODE)"; \
	fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@# Check if gh CLI is installed and authenticated
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "❌ GitHub CLI (gh) is required."; \
		echo "Install: brew install gh"; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ] && ! gh auth status >/dev/null 2>&1; then \
		echo "❌ Not authenticated with GitHub."; \
		echo "Run: gh auth login"; \
		exit 1; \
	fi
	@# Get current branch and find PR
	@CURRENT_BRANCH=$$(git branch --show-current 2>/dev/null); \
	if [ -z "$$CURRENT_BRANCH" ]; then \
		echo "❌ Not in a git repository or detached HEAD."; \
		exit 1; \
	fi; \
	echo "📍 Current branch: $$CURRENT_BRANCH"; \
	echo ""; \
	echo "🔍 Finding PR for this branch..."; \
	PR_NUMBER=$$(cd .. && gh pr list --head $$CURRENT_BRANCH --json number --jq '.[0].number' 2>/dev/null); \
	if [ -z "$$PR_NUMBER" ]; then \
		echo "❌ No PR found for branch $$CURRENT_BRANCH."; \
		echo ""; \
		echo "Create a PR first:"; \
		echo "  make create-pr-apply"; \
		exit 1; \
	fi; \
	echo "✅ Found PR #$$PR_NUMBER"; \
	echo ""; \
	echo "🔍 Checking for running deployments..."; \
	echo ""; \
	RUNNING_DEPLOYS=$$(cd .. && gh run list --workflow="deploy" --status="in_progress" --json databaseId,name --jq 'length' 2>/dev/null || echo "0"); \
	if [ "$$RUNNING_DEPLOYS" != "0" ]; then \
		echo "⚠️  Found $$RUNNING_DEPLOYS deployment(s) currently running."; \
		echo ""; \
		cd .. && gh run list --workflow="deploy" --status="in_progress" --json name,databaseId,headBranch,createdAt,url --template '{{range .}}⏳ {{.name}} ({{.headBranch}}){{"\n"}}   Started: {{timeago .createdAt}}{{"\n"}}   URL: {{.url}}{{"\n\n"}}{{end}}'; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
	else \
		echo "✅ No deployments currently running."; \
	fi; \
	echo ""; \
	if [ "$(_APPLY)" = "1" ] && [ "$(_SKIP_VALIDATION)" = "1" ]; then \
		echo "⚠️  WARNING: You are about to bypass validation checks!"; \
		echo ""; \
		echo "This is NOT recommended and should only be used in emergencies."; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "🔐 TWO-FACTOR AUTHENTICATION REQUIRED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		read -p "First confirmation - Are you absolutely sure? (yes/no) " -r CONFIRM1; \
		echo ""; \
		if [ "$$CONFIRM1" != "yes" ]; then \
			echo "❌ Operation cancelled."; \
			exit 1; \
		fi; \
		CHALLENGE_WORDS=("elephant" "volcano" "butterfly" "telescope" "saxophone" "submarine" "hurricane" "avalanche" "labyrinth" "astronaut" "flamingo" "cathedral" "pineapple" "octopus" "marshmallow"); \
		RANDOM_INDEX=$$(($$RANDOM % 15)); \
		CHALLENGE_WORD=$${CHALLENGE_WORDS[$$RANDOM_INDEX]}; \
		echo "🔐 Second confirmation - Type the following word EXACTLY (no copy-paste):"; \
		echo ""; \
		echo "   >>> $$CHALLENGE_WORD <<<"; \
		echo ""; \
		stty -echo 2>/dev/null || true; \
		trap 'stty echo 2>/dev/null || true' EXIT; \
		read -p "Type the word: " USER_INPUT; \
		stty echo 2>/dev/null || true; \
		trap - EXIT; \
		echo ""; \
		echo ""; \
		if [ "$$USER_INPUT" != "$$CHALLENGE_WORD" ]; then \
			echo "❌ Authentication failed. Word mismatch."; \
			echo "   Expected: $$CHALLENGE_WORD"; \
			echo "   Got: $$USER_INPUT"; \
			exit 1; \
		fi; \
		echo "✅ Two-factor authentication successful."; \
		echo ""; \
		echo "⚠️  Bypassing validation checks..."; \
		echo ""; \
	elif [ "$(_APPLY)" = "1" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "🔍 Running Validation Checks..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		if ! $(MAKE) validate-all; then \
			echo ""; \
			echo "❌ Validation failed. Cannot deploy."; \
			echo ""; \
			echo "Please fix the issues and try again, or use 'make deploy-dev-force' to bypass (not recommended)."; \
			exit 1; \
		fi; \
		echo ""; \
	fi; \
	if [ "$(_APPLY)" != "1" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📋 DRY-RUN Summary:"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "  Branch:  $$CURRENT_BRANCH"; \
		echo "  PR:      #$$PR_NUMBER"; \
		echo "  Comment: deploy --target=Container"; \
		echo ""; \
		echo "🔄 DRY-RUN: Would add deployment comment to PR"; \
		if [ "$$RUNNING_DEPLOYS" != "0" ]; then \
			echo "⚠️  DRY-RUN: Would prompt for confirmation due to running deployments"; \
		fi; \
		echo ""; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "✅ DRY-RUN completed. No changes were made."; \
		echo ""; \
		echo "💡 To actually deploy, run:"; \
		echo "   make deploy-dev-apply"; \
	else \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "🚀 Triggering deployment..."; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		if [ "$$RUNNING_DEPLOYS" != "0" ]; then \
			read -p "⚠️  Deploy anyway? This may cause conflicts. (y/n) " -n 1 -r; \
			echo ""; \
			if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
				echo "❌ Deployment cancelled."; \
				echo ""; \
				echo "💡 Wait for current deployment to finish or use:"; \
				echo "   make list-deploys"; \
				exit 1; \
			fi \
		fi; \
		echo "⚠️  APPLY MODE: This will trigger a deployment."; \
		echo ""; \
		read -p "Are you sure you want to continue? (yes/no) " -r CONFIRM; \
		echo ""; \
		if [ "$$CONFIRM" != "yes" ]; then \
			echo "❌ Operation cancelled."; \
			exit 1; \
		fi; \
		echo "💬 Adding comment to PR #$$PR_NUMBER: 'deploy --target=Container'"; \
		cd .. && gh pr comment $$PR_NUMBER --body "deploy --target=Container" 2>&1; \
		if [ $$? -eq 0 ]; then \
			echo ""; \
			echo "✅ Deploy comment added successfully!"; \
			echo ""; \
			echo "⏳ Waiting for workflow to start (checking for 30 seconds)..."; \
			echo ""; \
			for i in 1 2 3 4 5 6; do \
				sleep 5; \
				LATEST_RUN=$$(cd .. && gh run list --workflow="deploy" --limit 1 --json databaseId,status,conclusion,url,createdAt --jq '.[0]' 2>/dev/null); \
				if [ -n "$$LATEST_RUN" ]; then \
					RUN_URL=$$(echo "$$LATEST_RUN" | jq -r '.url'); \
					RUN_STATUS=$$(echo "$$LATEST_RUN" | jq -r '.status'); \
					RUN_CREATED=$$(echo "$$LATEST_RUN" | jq -r '.createdAt'); \
					SECONDS_AGO=$$(( ($$(date +%s) - $$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$$RUN_CREATED" +%s 2>/dev/null || echo 0)) )); \
					if [ $$SECONDS_AGO -lt 60 ]; then \
						echo "✅ Deployment workflow started!"; \
						echo ""; \
						echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
						echo "📊 Deployment Details:"; \
						echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
						echo ""; \
						echo "Status: $$RUN_STATUS"; \
						echo "🔗 Run URL: $$RUN_URL"; \
						echo ""; \
						$(call send-notifications,deploy,"Deployment to dev environment","Deployment triggered via PR #$$PR_NUMBER from branch $$CURRENT_BRANCH","$$RUN_URL"); \
						echo "💡 Monitor deployment:"; \
						echo "   - Watch in terminal: gh run watch"; \
						echo "   - Check status: make list-deploys"; \
						echo "   - Open in browser: $$RUN_URL"; \
						echo ""; \
						break; \
					fi \
				fi; \
				echo "   Still waiting... ($$i/6)"; \
			done; \
			if [ $$i -eq 6 ]; then \
				echo ""; \
				echo "⚠️  Workflow didn't start within 30 seconds."; \
				echo ""; \
				echo "This is normal if:"; \
				echo "  - Workflow trigger is configured differently"; \
				echo "  - There's a delay in GitHub Actions"; \
				echo ""; \
				echo "Check manually:"; \
				echo "  - View PR: cd .. && gh pr view $$PR_NUMBER"; \
				echo "  - List runs: make list-deploys"; \
			fi \
		else \
			echo ""; \
			echo "❌ Failed to add comment to PR."; \
			echo ""; \
			echo "Try manually:"; \
			echo "  cd .. && gh pr comment $$PR_NUMBER --body 'deploy --target=Container'"; \
			exit 1; \
		fi \
	fi

# List current running deployments
list-deploys:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Current Deployments Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@# Check if gh CLI is installed and authenticated
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "❌ GitHub CLI (gh) is required."; \
		echo "Install: brew install gh"; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ] && ! gh auth status >/dev/null 2>&1; then \
		echo "❌ Not authenticated with GitHub."; \
		echo "Run: gh auth login"; \
		exit 1; \
	fi
	@# Show in-progress deployments
	@echo "⏳ In-Progress Deployments:"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	IN_PROGRESS=$$(cd .. && gh run list --workflow="deploy" --status="in_progress" --limit 5 --json name,databaseId,headBranch,createdAt,url 2>/dev/null); \
	if [ "$$IN_PROGRESS" = "[]" ] || [ -z "$$IN_PROGRESS" ]; then \
		echo "   No deployments currently in progress."; \
	else \
		cd .. && gh run list --workflow="deploy" --status="in_progress" --limit 5 --json name,databaseId,headBranch,createdAt,url,workflowName \
			--template '{{range .}}⏳ {{.name}} ({{.headBranch}}){{"\n"}}   Workflow: {{.workflowName}}{{"\n"}}   Started: {{timeago .createdAt}}{{"\n"}}   URL: {{.url}}{{"\n\n"}}{{end}}'; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Recently Completed Deployments:"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	COMPLETED=$$(cd .. && gh run list --workflow="deploy" --status="completed" --limit 5 --json name,databaseId,conclusion,headBranch,createdAt,url 2>/dev/null); \
	if [ "$$COMPLETED" = "[]" ] || [ -z "$$COMPLETED" ]; then \
		echo "   No recent completed deployments."; \
	else \
		cd .. && gh run list --workflow="deploy" --status="completed" --limit 5 --json name,databaseId,conclusion,headBranch,createdAt,url,workflowName \
			--template '{{range .}}{{if eq .conclusion "success"}}✅{{else if eq .conclusion "failure"}}❌{{else if eq .conclusion "cancelled"}}⏸️{{else}}⚠️{{end}}{{else}}⏳{{end}} {{.name}} ({{.headBranch}}) - {{.conclusion}}{{"\n"}}   Workflow: {{.workflowName}}{{"\n"}}   Completed: {{timeago .createdAt}}{{"\n"}}   URL: {{.url}}{{"\n\n"}}{{end}}'; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "💡 Useful commands:"; \
	echo "   - Watch a deployment: gh run watch <run-id>"; \
	echo "   - View run details: gh run view <run-id>"; \
	echo "   - Cancel a run: gh run cancel <run-id>"; \
	echo "   - Trigger new deploy: make deploy-dev"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ============================================================================
# Service Management
# ============================================================================
#
# Start services with different configurations:
#
#   make start              Start with REAL Python backend (default, no mocks)
#   make start-ui-mck       Start with UI mocks only (no Python backend)
#   make start-svc-mck      Start with Python backend + mocked external services
#   make start-no-mocks     Start with REAL Python backend and real connections
#
# ============================================================================

# Start services (default: Python backend with mocked external services)
start: start-svc-mck

# Start with UI mocks only (no Python backend)
start-ui-mck:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Starting services with UI MOCK backend..."
	@echo "Python backend will NOT run. Only UI mock server."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh

# Start with Python backend + mocked external services (Databricks, AWS)
start-svc-mck:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Starting services with Python backend (MOCKED external services)..."
	@echo "External services (Databricks, AWS) will be mocked."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh --py

# Start with REAL Python backend and real connections to external services
start-no-mocks:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Starting services with Python backend (REAL connections)..."
	@echo "⚠️  WARNING: This requires all environment variables to be configured."	
	@echo "    See README.md for required environment variables."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh --py --no-mocks

# Start with Python backend in DEBUG mode (enhanced logging and debug features)
start-debug:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Starting services with Python backend in DEBUG mode..."
	@echo "🐛 DEBUG Features: Verbose logging, enhanced error details, debug endpoints"
	@echo "📊 Monitoring: Detailed request tracing, performance metrics, stack traces"
	@echo "🔄 Auto-reload: Code changes trigger automatic service restart"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh --py --debug

# Stop all services and cleanup
stop:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Stopping all services..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh --stop
	@echo "✓ All services stopped successfully"

# Dump service logs to timestamped file
dump-logs:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Dumping service logs..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	DUMP_FILE="logs/dump_$$TIMESTAMP.log"; \
	echo "Creating log dump: $$DUMP_FILE"; \
	echo "" > $$DUMP_FILE; \
	echo "═══════════════════════════════════════════════════════════════════════" >> $$DUMP_FILE; \
	echo "Service Logs Dump - $$TIMESTAMP" >> $$DUMP_FILE; \
	echo "═══════════════════════════════════════════════════════════════════════" >> $$DUMP_FILE; \
	echo "" >> $$DUMP_FILE; \
	if [ -f logs/ui.log ]; then \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		echo "UI Service Logs (logs/ui.log)" >> $$DUMP_FILE; \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		cat logs/ui.log >> $$DUMP_FILE; \
		echo "" >> $$DUMP_FILE; \
		echo "" >> $$DUMP_FILE; \
	fi; \
	if [ -f logs/service.log ]; then \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		echo "Backend Service Logs (logs/service.log)" >> $$DUMP_FILE; \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		cat logs/service.log >> $$DUMP_FILE; \
		echo "" >> $$DUMP_FILE; \
		echo "" >> $$DUMP_FILE; \
	fi; \
	if [ -f logs/caddy.log ]; then \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		echo "Caddy Proxy Logs (logs/caddy.log)" >> $$DUMP_FILE; \
		echo "───────────────────────────────────────────────────────────────────────" >> $$DUMP_FILE; \
		cat logs/caddy.log >> $$DUMP_FILE; \
		echo "" >> $$DUMP_FILE; \
	fi; \
	echo "═══════════════════════════════════════════════════════════════════════" >> $$DUMP_FILE; \
	echo "End of Log Dump" >> $$DUMP_FILE; \
	echo "═══════════════════════════════════════════════════════════════════════" >> $$DUMP_FILE; \
	echo "✓ Logs dumped to: $$DUMP_FILE"

# Restart all services
restart:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Restarting all services..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./tools/startup.sh --restart
	@echo "✓ All services restarted successfully"

# Check status of running services
status:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Service Status Check"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Checking ports..."
	@for port in 3000 3006 8000 8001; do \
		if lsof -ti tcp:$$port > /dev/null 2>&1; then \
			echo "✓ Port $$port: IN USE"; \
		else \
			echo "✗ Port $$port: FREE"; \
		fi; \
	done
	@echo ""
	@echo "Checking tmux session..."
	@if tmux has-session -t startup 2>/dev/null; then \
		echo "✓ tmux session 'startup': RUNNING"; \
		echo "  Use 'tmux attach -t startup' to connect"; \
	else \
		echo "✗ tmux session 'startup': NOT FOUND"; \
	fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check health of all running services with HTTP endpoints
health:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🏥 Health Check - All Services"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📋 Port Status:"
	@for port in 3000 3006 8000 8001; do \
		if lsof -ti tcp:$$port > /dev/null 2>&1; then \
			echo "  ✓ Port $$port: IN USE"; \
		else \
			echo "  ✗ Port $$port: FREE (service not running)"; \
		fi; \
	done
	@echo ""
	@echo "🌐 HTTP Health Checks:"
	@echo ""
	@echo "  🔹 UI Frontend (React - Port 3006):"
	@if lsof -ti tcp:3006 > /dev/null 2>&1; then \
		if curl -s --max-time 3 http://localhost:3006 > /dev/null 2>&1; then \
			RESPONSE_TIME=$$(curl -o /dev/null -s -w '%{time_total}' --max-time 3 http://localhost:3006 2>/dev/null); \
			echo "     ✅ HEALTHY (Response: $${RESPONSE_TIME}s)"; \
		else \
			echo "     ⚠️  PORT OPEN but NOT RESPONDING"; \
		fi; \
	else \
		echo "     ❌ NOT RUNNING"; \
	fi
	@echo ""
	@echo "  🔹 Caddy Reverse Proxy (HTTPS - Port 3000):"
	@if lsof -ti tcp:3000 > /dev/null 2>&1; then \
		if curl -sk --max-time 3 https://localhost:3000 > /dev/null 2>&1; then \
			RESPONSE_TIME=$$(curl -sk -o /dev/null -s -w '%{time_total}' --max-time 3 https://localhost:3000 2>/dev/null); \
			echo "     ✅ HEALTHY (Response: $${RESPONSE_TIME}s)"; \
			echo "     🌍 Access: https://localhost:3000"; \
		else \
			echo "     ⚠️  PORT OPEN but NOT RESPONDING"; \
		fi; \
	else \
		echo "     ❌ NOT RUNNING"; \
	fi
	@echo ""
	@echo "  🔹 Python Backend API (Port 8000):"
	@if lsof -ti tcp:8000 > /dev/null 2>&1; then \
		if curl -s --max-time 3 https://localhost:8000/health > /dev/null 2>&1 || curl -s --max-time 3 https://localhost:8000 > /dev/null 2>&1; then \
			RESPONSE_TIME=$$(curl -o /dev/null -s -w '%{time_total}' --max-time 3 https://localhost:8000 2>/dev/null); \
			echo "     ✅ HEALTHY (Response: $${RESPONSE_TIME}s)"; \
		else \
			echo "     ⚠️  PORT OPEN but NOT RESPONDING"; \
		fi; \
	else \
		echo "     ❌ NOT RUNNING"; \
	fi
	@echo ""
	@echo "  🔹 Mock Backend/Python Service (Port 8001):"
	@if lsof -ti tcp:8001 > /dev/null 2>&1; then \
		if curl -s --max-time 3 http://localhost:8001 > /dev/null 2>&1; then \
			RESPONSE_TIME=$$(curl -o /dev/null -s -w '%{time_total}' --max-time 3 http://localhost:8001 2>/dev/null); \
			echo "     ✅ HEALTHY (Response: $${RESPONSE_TIME}s)"; \
		else \
			echo "     ⚠️  PORT OPEN but NOT RESPONDING"; \
		fi; \
	else \
		echo "     ❌ NOT RUNNING"; \
	fi
	@echo ""
	@echo "🔧 Process Management:"
	@if tmux has-session -t startup 2>/dev/null; then \
		echo "  ✓ tmux session 'startup': RUNNING"; \
		echo "    - Use 'tmux attach -t startup' to view logs"; \
		echo "    - Use 'make dump-logs' to save logs to file"; \
	else \
		echo "  ✗ tmux session 'startup': NOT FOUND"; \
		echo "    - Run 'make start' to start services"; \
	fi
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@ALL_HEALTHY=true; \
	for port in 3000 3006 8000 8001; do \
		if ! lsof -ti tcp:$$port > /dev/null 2>&1; then \
			ALL_HEALTHY=false; \
			break; \
		fi; \
	done; \
	if [ "$$ALL_HEALTHY" = "true" ]; then \
		echo "✅ Overall Status: ALL SERVICES HEALTHY"; \
	else \
		echo "⚠️  Overall Status: SOME SERVICES DOWN"; \
		echo "   Run 'make start' to start all services"; \
	fi
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ============================================================================
# Installation & Setup
# ============================================================================

# Install all dependencies (UI + Service)
install: install-ui install-service
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ All dependencies installed successfully!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Install UI dependencies
install-ui:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Installing UI dependencies..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm install
	@echo "✓ UI dependencies installed"

# Install Python service dependencies
install-service:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Installing Python service dependencies..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd service && \
		python3.12 -m venv venv && \
		source venv/bin/activate && \
		pip install -r requirements.txt -r requirements-test.txt && \
		pip install -e .
	@echo "✓ Python service dependencies installed"

# ============================================================================
# Build Commands
# ============================================================================

# Build only the UI (React)
build-ui:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Building UI (React)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run build
	@echo "✓ UI build complete!"

# Build only the Python service
build-service:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Building Python service..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd service && source venv/bin/activate && python3.12 -m build
	@echo "✓ Python service build complete!"

# Build both UI and service
build: build-ui build-service
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ All components built successfully!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	$(MAKE) validate-all
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ Nuke-all complete! Environment fully reset, rebuilt, and validated."

# ============================================================================
# Testing & Linting
# ============================================================================

# Run UI tests (Vitest)
test-ui:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running UI Tests (Vitest)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm test

# Run E2E tests (Python Playwright - Default)
test-e2e:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running E2E Tests (Python Playwright)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(MAKE) test-e2e-py

# Run E2E tests with Python Playwright
test-e2e-py:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running Python E2E Tests (pytest-playwright)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ ! -d "service/venv" ]; then \
		echo "❌ Virtual environment not found. Run 'make install-service' first."; \
		exit 1; \
	fi
	@echo "Checking if application is running on port 3000..."
	@if ! lsof -ti tcp:3000 > /dev/null 2>&1; then \
		echo "⚠️  Application not running. Starting services without mocks..."; \
		$(MAKE) start-no-mocks > /dev/null 2>&1 & \
		echo "⏳ Waiting for application to start (checking port 3000)..."; \
		for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
			if lsof -ti tcp:3000 > /dev/null 2>&1; then \
				echo "✓ Application started successfully!"; \
				sleep 3; \
				break; \
			fi; \
			if [ $$i -eq 30 ]; then \
				echo "❌ Application failed to start within 30 seconds."; \
				exit 1; \
			fi; \
			sleep 1; \
		done; \
	else \
		echo "✓ Application is already running."; \
	fi
	@cd .. && source Container/service/venv/bin/activate && \
		python -m pytest tests/e2e/ -v

# Run E2E tests with Python Playwright in headed mode (visible browser for debugging)
test-e2e-py-headed:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running Python E2E Tests (Headed Mode - Visible Browser)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ ! -d "service/venv" ]; then \
		echo "❌ Virtual environment not found. Run 'make install-service' first."; \
		exit 1; \
	fi
	@echo "Checking if application is running on port 3000..."
	@if ! lsof -ti tcp:3000 > /dev/null 2>&1; then \
		echo "⚠️  Application not running. Starting services without mocks..."; \
		$(MAKE) start-no-mocks > /dev/null 2>&1 & \
		echo "⏳ Waiting for application to start (checking port 3000)..."; \
		for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
			if lsof -ti tcp:3000 > /dev/null 2>&1; then \
				echo "✓ Application started successfully!"; \
				sleep 3; \
				break; \
			fi; \
			if [ $$i -eq 30 ]; then \
				echo "❌ Application failed to start within 30 seconds."; \
				exit 1; \
			fi; \
			sleep 1; \
		done; \
	else \
		echo "✓ Application is already running."; \
	fi
	@cd .. && source Container/service/venv/bin/activate && \
		python -m pytest tests/e2e/ -v --headed --slowmo=100

# Run E2E tests with JavaScript/TypeScript Playwright (legacy)
test-e2e-js:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running JavaScript E2E Tests (Legacy Playwright)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run test:e2e

# Run E2E tests with UI mode
test-e2e-ui:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running E2E Tests with UI (Playwright)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run test:e2e:ui

# Run E2E tests in debug mode
test-e2e-debug:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running E2E Tests in Debug Mode (Playwright)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run test:e2e:debug

# Show E2E test report
test-e2e-report:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Opening E2E Test Report..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run test:e2e:report

# Run Playwright codegen for recording new tests
test-e2e-codegen:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Starting Playwright Codegen..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Note: Make sure the app is running (make start) before using codegen"
	@cd ui && npm run test:e2e:codegen

# Run Python service tests (pytest)
test-service:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running Service Tests (pytest)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd service && source venv/bin/activate && pytest

# Run UI linting (ESLint)
lint-ui:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running UI Linting (ESLint)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd ui && npm run lint

# Run Python service linting (flake8)
lint-service:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running Service Linting (flake8)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@cd service && source venv/bin/activate && flake8 src/ tests/

# Run all tests (UI + Service)
test: test-ui test-service
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ All tests completed!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Run all linting (UI + Service)
lint: lint-ui lint-service
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ All linting completed!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ============================================================================
# Maintenance & Cleanup
# ============================================================================

# Clean build artifacts and dependencies
clean:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Cleaning build artifacts and dependencies..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  This will remove:"
	@echo "  - UI: node_modules, dist, build, coverage"
	@echo "  - Service: venv, __pycache__, .pytest_cache, *.pyc, dist, build"
	@echo "  - Logs: All log files"
	@echo ""
	@read -p "Are you sure you want to continue? (yes/no) " -r CONFIRM; \
	echo ""; \
	if [ "$$CONFIRM" != "yes" ]; then \
		echo "❌ Operation cancelled."; \
		exit 1; \
	fi; \
	echo "🗑️  Removing UI artifacts..."; \
	rm -rf ui/node_modules ui/dist ui/build ui/coverage ui/.vite 2>/dev/null || true; \
	echo "✓ UI artifacts removed"; \
	echo ""; \
	echo "🗑️  Removing Python service artifacts..."; \
	rm -rf service/venv service/dist service/build service/*.egg-info 2>/dev/null || true; \
	find service -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true; \
	find service -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true; \
	find service -type f -name "*.pyc" -delete 2>/dev/null || true; \
	find service -type f -name "*.pyo" -delete 2>/dev/null || true; \
	echo "✓ Python service artifacts removed"; \
	echo ""; \
	echo "🗑️  Removing logs..."; \
	rm -rf logs/*.log 2>/dev/null || true; \
	echo "✓ Logs removed"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "✅ Clean complete!"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "💡 Next step: Run 'make install' to reinstall dependencies"

# OBLITERATE - Nuclear option: Remove EVERYTHING including caches, temp files, and environment files
obliterate:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "💥 OBLITERATE - Nuclear Cleanup"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "⚠️  ⚠️  ⚠️  WARNING: EXTREMELY DESTRUCTIVE OPERATION ⚠️  ⚠️  ⚠️"
	@echo ""
	@echo "This will PERMANENTLY DELETE:"
	@echo "  📦 Dependencies: node_modules, venv, package-lock.json"
	@echo "  🏗️  Build artifacts: dist, build, *.egg-info"
	@echo "  🗂️  Caches: __pycache__, .pytest_cache, .npm, .cache, coverage"
	@echo "  📝 Logs: All log files and dump files"
	@echo "  🔧 Temp files: *.pyc, *.pyo, .DS_Store, *.swp, *.swo"
	@echo "  📊 Coverage: .coverage, coverage/, htmlcov/"
	@echo "  🔐 Environment: .env.local (backup recommended!)"
	@echo "  🎯 IDE files: .vscode/.*, .idea/"
	@echo "  📦 npm cache, pip cache directories"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔐 TWO-FACTOR CONFIRMATION REQUIRED"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@read -p "First confirmation - Type 'OBLITERATE' (all caps): " -r CONFIRM1; \
	echo ""; \
	if [ "$$CONFIRM1" != "OBLITERATE" ]; then \
		echo "❌ Operation cancelled. You must type exactly: OBLITERATE"; \
		exit 1; \
	fi; \
	CHALLENGE_WORDS=("DESTROY" "ERASE" "PURGE" "DELETE" "NUKE" "REMOVE" "WIPE" "CLEAR"); \
	RANDOM_INDEX=$$(($$RANDOM % 8)); \
	CHALLENGE_WORD=$${CHALLENGE_WORDS[$$RANDOM_INDEX]}; \
	echo "🔐 Second confirmation - Type the following word EXACTLY:"; \
	echo ""; \
	echo "   >>> $$CHALLENGE_WORD <<<"; \
	echo ""; \
	read -p "Type the word: " USER_INPUT; \
	echo ""; \
	if [ "$$USER_INPUT" != "$$CHALLENGE_WORD" ]; then \
		echo "❌ Authentication failed. Word mismatch."; \
		exit 1; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "💥 OBLITERATING EVERYTHING..."; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "🛑 Stopping all services first..."; \
	$(MAKE) stop 2>/dev/null || true; \
	echo ""; \
	echo "🗑️  [1/12] Removing UI dependencies (node_modules)..."; \
	rm -rf ui/node_modules ui/package-lock.json 2>/dev/null || true; \
	echo "✓ UI dependencies obliterated"; \
	echo ""; \
	echo "🗑️  [2/12] Removing Python virtual environment..."; \
	rm -rf service/venv 2>/dev/null || true; \
	echo "✓ Virtual environment obliterated"; \
	echo ""; \
	echo "🗑️  [3/12] Removing build artifacts..."; \
	rm -rf ui/dist ui/build ui/.vite 2>/dev/null || true; \
	rm -rf service/dist service/build service/*.egg-info 2>/dev/null || true; \
	echo "✓ Build artifacts obliterated"; \
	echo ""; \
	echo "🗑️  [4/12] Removing Python cache files..."; \
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true; \
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true; \
	find . -type f -name "*.pyc" -delete 2>/dev/null || true; \
	find . -type f -name "*.pyo" -delete 2>/dev/null || true; \
	echo "✓ Python cache obliterated"; \
	echo ""; \
	echo "🗑️  [5/12] Removing coverage files..."; \
	rm -rf ui/coverage .coverage coverage/ htmlcov/ service/.coverage 2>/dev/null || true; \
	echo "✓ Coverage files obliterated"; \
	echo ""; \
	echo "🗑️  [6/12] Removing log files..."; \
	rm -rf logs/*.log logs/dump_*.log 2>/dev/null || true; \
	echo "✓ Logs obliterated"; \
	echo ""; \
	echo "🗑️  [7/12] Removing cache directories..."; \
	rm -rf .cache .npm service/.cache 2>/dev/null || true; \
	echo "✓ Cache directories obliterated"; \
	echo ""; \
	echo "🗑️  [8/12] Removing temporary files..."; \
	find . -type f -name ".DS_Store" -delete 2>/dev/null || true; \
	find . -type f -name "*.swp" -delete 2>/dev/null || true; \
	find . -type f -name "*.swo" -delete 2>/dev/null || true; \
	find . -type f -name "*.tmp" -delete 2>/dev/null || true; \
	find . -type f -name "*~" -delete 2>/dev/null || true; \
	echo "✓ Temporary files obliterated"; \
	echo ""; \
	echo "🗑️  [9/12] Removing IDE configuration caches..."; \
	rm -rf .vscode/.ropeproject .idea/ 2>/dev/null || true; \
	echo "✓ IDE caches obliterated"; \
	echo ""; \
	echo "🗑️  [10/12] Removing environment files..."; \
	if [ -f .env.local ]; then \
		echo "⚠️  Found .env.local - Creating backup at .env.local.backup"; \
		cp .env.local .env.local.backup; \
		rm .env.local; \
	fi; \
	echo "✓ Environment files cleaned"; \
	echo ""; \
	echo "🗑️  [11/12] Cleaning npm cache..."; \
	cd ui && npm cache clean --force 2>/dev/null || true; \
	echo "✓ npm cache obliterated"; \
	echo ""; \
	echo "🗑️  [12/12] Cleaning pip cache..."; \
	python3.12 -m pip cache purge 2>/dev/null || true; \
	echo "✓ pip cache obliterated"; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "💥 OBLITERATION COMPLETE!"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "📊 Summary:"; \
	echo "  ✅ All dependencies removed"; \
	echo "  ✅ All build artifacts removed"; \
	echo "  ✅ All cache files removed"; \
	echo "  ✅ All temporary files removed"; \
	echo "  ✅ All logs removed"; \
	echo "  ✅ npm & pip caches cleared"; \
	if [ -f .env.local.backup ]; then \
		echo "  ⚠️  .env.local backed up to .env.local.backup"; \
	fi; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo "🔄 Next Steps:"; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	echo ""; \
	echo "  1️⃣  make install      # Reinstall all dependencies"; \
	echo "  2️⃣  make start        # Start services"; \
	echo ""; \
	if [ -f .env.local.backup ]; then \
		echo "  💡 Restore your environment:"; \
		echo "     mv .env.local.backup .env.local"; \
		echo ""; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

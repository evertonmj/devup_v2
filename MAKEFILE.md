# Makefile Documentation - DevUp

## Overview

This document describes how to use the Makefile to simplify local development of DevUp. The Makefile provides convenient commands for managing dependencies, starting/stopping services, running tests, and maintaining the codebase.

## Prerequisites

Before using the Makefile commands, ensure you have:

- macOS with iTerm2 and tmux installed
- Python 3.12+
- Node.js 18+
- Caddy web server
- All environment variables configured (see [README.md](README.md) for details)

## Quick Start

```bash
# Navigate to Container directory
cd Container

# Check if all required tools are installed
make check-tools

# Install all dependencies (first time only)
make install

# Start the application with mock backend
make start

# Access the application at https://localhost:3000
```

## Available Commands

### Prerequisites Check

#### `make check-tools`
Verifies that all required development tools are installed on your system.

```bash
make check-tools
```

**What it checks:**
- Python 3.12+ (checks version and provides install command)
- Node.js 18+ (checks version and provides install command)
- npm (package manager for Node.js)
- tmux (terminal multiplexer)
- Caddy (web server for reverse proxy)
- OpenSSL (SSL/TLS toolkit)
- iTerm2 (terminal emulator)
- Homebrew (package manager for macOS)

**Example output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Checking Required Tools Installation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Checking Python...
  ✓ Python 3.13.7 (OK)

Checking Node.js...
  ✓ Node.js 24.3.0 (OK)

Checking npm...
  ✓ npm 11.5.2 (OK)

Checking tmux...
  ✓ tmux 3.5a (OK)

Checking Caddy...
  ✓ Caddy v2.10.0 (OK)

Checking OpenSSL...
  ✓ OpenSSL 3.5.4 (OK)

Checking iTerm2...
  ✓ iTerm2 installed (OK)

Checking Homebrew...
  ✓ Homebrew 4.6.18 (OK)
```

If any tool is missing or has an incorrect version, the command provides the exact installation command needed.

#### `make check-pipeline`
Checks the GitHub Actions pipeline status for the current repository.

```bash
make check-pipeline
```

**Prerequisites:**
- GitHub CLI (`gh`) must be installed: `brew install gh`
- Must be authenticated using one of these methods:
  - **Option 1 (Recommended)**: GitHub CLI authentication: `gh auth login`
  - **Option 2**: Set `GITHUB_TOKEN` environment variable with a GitHub Personal Access Token
- Must be run from within a git repository

**Authentication Options:**

**Option 1: GitHub CLI (Recommended)**
```bash
# Install GitHub CLI
brew install gh

# Authenticate interactively
gh auth login
```

**Option 2: GitHub Token**
```bash
# Create a token at: https://github.com/settings/tokens
# Required scopes: 'repo' and 'workflow'

# Set as environment variable (current session)
export GITHUB_TOKEN='ghp_your_token_here'

# Or add to your shell profile (~/.zshrc or ~/.bash_profile)
echo 'export GITHUB_TOKEN="ghp_your_token_here"' >> ~/.zshrc

# Then run the check
make check-pipeline
```

**Creating a GitHub Personal Access Token:**
1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Give it a descriptive name (e.g., "Local Dev Pipeline Check")
4. Select scopes: `repo` and `workflow`
5. Click "Generate token"
6. Copy the token (you won't be able to see it again!)
7. Set it as `GITHUB_TOKEN` environment variable

**What it shows:**
- Current repository name
- Current branch
- Latest 5 workflow runs with status (✅ success, ❌ failure, ⏳ in progress, ⏸️ cancelled)
- Workflow name and branch
- Time since run creation
- Direct links to workflow runs

**Example output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GitHub Actions Pipeline Status Check
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Repository: owner/repo-name
🌿 Current Branch: feat/my-feature

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Latest Workflow Runs:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ CI Tests (main)
   Status: success | 2h ago
   https://github.com/owner/repo/actions/runs/123456

⏳ Build and Deploy (feat/my-feature)
   Status: in_progress | 5m ago
   https://github.com/owner/repo/actions/runs/123457

❌ Lint (feat/my-feature)
   Status: failure | 10m ago
   https://github.com/owner/repo/actions/runs/123458
```

**Additional gh commands:**
- View all runs: `gh run list`
- Watch a run in real-time: `gh run watch`
- View run details: `gh run view <run-id>`
- Re-run failed jobs: `gh run rerun <run-id>`

---

### GitHub Workflow Management

#### Safety Features

All GitHub workflow commands that make changes (`create-pr`, `deploy-dev`) have multiple layers of protection:

##### 1. Dry-Run Mode (Default)

Commands show what would happen without making any actual changes.

##### 2. Automatic Validation

Before creating PRs or deploying, the system automatically runs:

- UI linting
- Service linting
- UI tests
- Service tests
- UI build

##### 3. Two-Factor Authentication

If you need to bypass validation (not recommended), you must:

- Confirm with "yes" twice
- Type a random word exactly (no copy-paste allowed)

##### Command Variants

```bash
# Safe preview - no changes, no validation
make create-pr
make deploy-dev

# Execute with validation - runs all tests/lints/builds first
make create-pr-apply
make deploy-dev-apply

# Force execution - skips validation, requires 2FA
make create-pr-force
make deploy-dev-force
```

##### Why these safeguards exist

- Prevents accidental PR creation or deployments
- Ensures code quality before deployment
- Catches bugs before they reach production
- Provides emergency bypass for urgent situations
- Makes it harder to skip important checks

#### Notification System

The Makefile can automatically send notifications to Google Chat and/or Microsoft Teams when Pull Requests are created or deployments are triggered. This keeps your team informed about important development activities.

##### Supported Services

- **Google Chat**: Send rich card notifications with PR/deployment details
- **Microsoft Teams**: Send MessageCard notifications with deployment information
- **Both services can be enabled independently** - use one, both, or neither

##### Setup Instructions

1. **Copy the environment template:**

   ```bash
   cp .env.example .env
   ```

2. **Configure Google Chat (optional):**
   - Go to your Google Chat space
   - Click space name → Apps & integrations → Add webhooks
   - Name it (e.g., "Engineering Supervisor Agent Notifications")
   - Copy the webhook URL
   - Edit `.env` and set:

     ```bash
     GOOGLE_CHAT_ENABLED=true
     GOOGLE_CHAT_WEBHOOKS=https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN
     ```

3. **Configure Microsoft Teams (optional):**
   - Go to your Teams channel
   - Click the three dots (•••) → Connectors or Workflows
   - Search for "Incoming Webhook" and configure
   - Copy the webhook URL
   - Edit `.env` and set:

     ```bash
     TEAMS_ENABLED=true
     TEAMS_WEBHOOKS=https://outlook.office.com/webhook/GUID@GUID/IncomingWebhook/GUID/GUID
     ```

##### Multiple Channels

You can send notifications to multiple channels by separating webhook URLs with commas:

```bash
GOOGLE_CHAT_WEBHOOKS=https://chat.googleapis.com/.../SPACE1/...,https://chat.googleapis.com/.../SPACE2/...
TEAMS_WEBHOOKS=https://outlook.office.com/.../WEBHOOK1/...,https://outlook.office.com/.../WEBHOOK2/...
```

##### What Gets Notified

Notifications are sent automatically when using the `-apply` commands:

- **PR Creation** (`make create-pr-apply`): Sends notification with PR title, branch, author, and link
- **Deployments** (`make deploy-dev-apply`): Sends notification with deployment target, PR number, branch, and workflow link

You can control what gets notified via `.env`:

```bash
# Control PR notifications
GOOGLE_CHAT_NOTIFY_PR=true
TEAMS_NOTIFY_PR=true

# Control deployment notifications
GOOGLE_CHAT_NOTIFY_DEPLOY=true
TEAMS_NOTIFY_DEPLOY=true
```

##### Message Customization

Optional environment variables for customizing notification messages:

```bash
# Custom prefixes
NOTIFICATION_PR_PREFIX="🔔 Pull Request Created"
NOTIFICATION_DEPLOY_PREFIX="🚀 Deployment Triggered"

# Include/exclude information
NOTIFICATION_INCLUDE_BRANCH=true
NOTIFICATION_INCLUDE_AUTHOR=true
```

##### Example `.env` Configuration

```bash
# Enable Google Chat for PR notifications only
GOOGLE_CHAT_ENABLED=true
GOOGLE_CHAT_WEBHOOKS=https://chat.googleapis.com/v1/spaces/ABC123/messages?key=KEY&token=TOKEN
GOOGLE_CHAT_NOTIFY_PR=true
GOOGLE_CHAT_NOTIFY_DEPLOY=false

# Enable Teams for all notifications
TEAMS_ENABLED=true
TEAMS_WEBHOOKS=https://outlook.office.com/webhook/GUID@GUID/IncomingWebhook/GUID/GUID
TEAMS_NOTIFY_PR=true
TEAMS_NOTIFY_DEPLOY=true
```

##### Disabling Notifications

To disable notifications, set the enabled flags to `false` or leave the webhook URLs empty:

```bash
GOOGLE_CHAT_ENABLED=false
TEAMS_ENABLED=false
```

Or simply don't create a `.env` file - notifications are disabled by default.

---

#### `make validate-all`

Runs all validation checks before creating PRs or deploying.

```bash
make validate-all
```

**What it does:**

1. Runs UI linting
2. Runs service linting
3. Runs UI tests
4. Runs service tests
5. Builds UI

If any check fails, the command exits with an error and shows which checks failed.

---

#### `make create-pr`

Creates a Pull Request from the current branch with an optional custom description or auto-generated description using GitHub Copilot.

**Dry-run mode (default - safe, no changes made):**

```bash
# Dry-run: Preview PR creation with auto-generated description
make create-pr

# Dry-run: Preview PR creation with custom description
make create-pr DESC="Your custom PR description"
```

**Apply mode (runs validation, requires confirmation):**

```bash
# Actually create PR with validation checks (requires "yes" confirmation)
make create-pr-apply

# With custom description
make create-pr-apply DESC="Your custom PR description"
```

**Force mode (skips validation, requires 2FA):**

```bash
# Bypass validation (NOT RECOMMENDED - requires 2FA)
make create-pr-force

# With custom description
make create-pr-force DESC="Your custom PR description"
```

**What it does in dry-run mode:**

- Shows what branch would be pushed
- Displays what PR title and target branch would be used
- No actual changes are made to GitHub
- No validation is run

**What it does in apply mode:**

1. **Runs validation:**
   - Executes all tests, lints, and builds
   - Fails if any validation check fails
   - Can be bypassed with force mode (requires 2FA)

2. **Validates environment:**
   - Checks for GitHub CLI installation
   - Verifies authentication
   - Ensures you're not on main/master branch

3. **Prompts for confirmation:**
   - Asks "Are you sure you want to continue? (yes/no)"
   - Must type "yes" (not just "y") to proceed
   - Operation cancelled if anything other than "yes" is entered

4. **Handles uncommitted changes:**
   - Prompts to commit if there are uncommitted changes
   - Allows you to commit interactively

5. **Pushes branch:**
   - Pushes current branch to remote
   - Sets upstream tracking

6. **Generates description:**
   - **With DESC parameter:** Uses your provided description
   - **Without DESC (default):** Attempts to use GitHub Copilot CLI to generate a comprehensive PR description
   - **Fallback:** Uses a default template if Copilot is unavailable

7. **Creates PR:**
   - Creates PR against `main` branch
   - Uses last commit message as title
   - Provides PR URL for review

**Example dry-run output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Creating Pull Request (DRY-RUN MODE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Current branch: feat/my-feature

🔍 DRY-RUN: Would push branch to remote: feat/my-feature

🤖 DRY-RUN: Would generate PR description with GitHub Copilot

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 DRY-RUN Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Branch:  feat/my-feature
  Target:  main
  Title:   Add new feature

🔄 DRY-RUN: Would create Pull Request

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ DRY-RUN completed. No changes were made.

💡 To actually create the PR, run:
   make create-pr APPLY=1
```

**Example apply mode output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Creating Pull Request (APPLY MODE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Current branch: feat/my-feature

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔍 Running Validation Checks...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Running All Validation Checks
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1️⃣  Running UI Linting...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ UI linting passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2️⃣  Running Service Linting...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Service linting passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3️⃣  Running UI Tests...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ UI tests passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4️⃣  Running Service Tests...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Service tests passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5️⃣  Building UI...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ UI build passed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ ALL VALIDATION CHECKS PASSED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  APPLY MODE: This will actually create a Pull Request.

Are you sure you want to continue? (yes/no) yes

🔄 Pushing branch to remote...

🤖 Generating PR description with GitHub Copilot...
   (This may take a few seconds)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Creating PR...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Pull Request created successfully!

🔗 PR URL: https://github.com/owner/repo/pull/123

💡 Next steps:
   - Review the PR: https://github.com/owner/repo/pull/123
   - Deploy to dev: make deploy-dev APPLY=1
   - Check deployments: make list-deploys
```

**Prerequisites:**
- GitHub Copilot CLI (optional, for auto-generation): `gh extension install github/gh-copilot`
- Must be on a feature branch (not main/master)
- All changes should be committed

#### `make deploy-dev`
Triggers a deployment to the dev environment by adding a comment to the PR associated with the current branch.

**Dry-run mode (default - safe, no changes made):**

```bash
# Dry-run: Preview deployment trigger
make deploy-dev
```

**Apply mode (requires confirmation):**

```bash
# Actually trigger deployment (requires "yes" confirmation)
make deploy-dev APPLY=1
```

**What it does in dry-run mode:**

- Finds the PR for your current branch
- Checks for running deployments
- Shows what comment would be added
- No actual changes are made to GitHub

**What it does in apply mode:**

1. **Finds associated PR:**
   - Automatically finds the PR for your current branch
   - Exits if no PR exists (suggests creating one)

2. **Checks for deployment collisions:**
   - Lists any currently running deployments
   - **If deployments are running:** Shows details and prompts for confirmation (y/n)
   - **If no deployments:** Proceeds to confirmation step

3. **Prompts for confirmation:**
   - Asks "Are you sure you want to continue? (yes/no)"
   - Must type "yes" (not just "y") to proceed
   - Operation cancelled if anything other than "yes" is entered

4. **Triggers deployment:**
   - Adds comment `deploy --target=Container` to the PR
   - This comment triggers your GitHub Actions workflow

5. **Waits for workflow:**
   - Monitors for 30 seconds for the workflow to start
   - **If started:** Displays run URL and status
   - **If not started:** Provides manual check instructions

6. **Returns deployment link:**
   - Provides direct link to the GitHub Actions run
   - Suggests monitoring commands

**Example dry-run output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deploy to Dev Environment (DRY-RUN MODE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Current branch: feat/my-feature

🔍 Finding PR for this branch...
✅ Found PR #123

🔍 Checking for running deployments...

✅ No deployments currently running.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 DRY-RUN Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Branch:  feat/my-feature
  PR:      #123
  Comment: deploy --target=Container

🔄 DRY-RUN: Would add deployment comment to PR

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ DRY-RUN completed. No changes were made.

💡 To actually deploy, run:
   make deploy-dev APPLY=1
```

**Example apply mode output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deploy to Dev Environment (APPLY MODE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Current branch: feat/my-feature

🔍 Finding PR for this branch...
✅ Found PR #123

🔍 Checking for running deployments...

✅ No deployments currently running.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 Triggering deployment...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  APPLY MODE: This will trigger a deployment.

Are you sure you want to continue? (yes/no) yes

💬 Adding comment to PR #123: 'deploy --target=Container'

✅ Deploy comment added successfully!

⏳ Waiting for workflow to start (checking for 30 seconds)...

   Still waiting... (1/6)
   Still waiting... (2/6)
✅ Deployment workflow started!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Deployment Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Status: in_progress
🔗 Run URL: https://github.com/owner/repo/actions/runs/123456

💡 Monitor deployment:
   - Watch in terminal: gh run watch
   - Check status: make list-deploys
   - Open in browser: https://github.com/owner/repo/actions/runs/123456
```

**Collision handling in apply mode:**

If a deployment is already running, you'll see an additional prompt:

```
⚠️  Found 1 deployment(s) currently running.

⏳ Deploy Container (main)
   Started: 5m ago
   URL: https://github.com/owner/repo/actions/runs/123455

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  Deploy anyway? This may cause conflicts. (y/n) y

⚠️  APPLY MODE: This will trigger a deployment.

Are you sure you want to continue? (yes/no) yes
```

**Prerequisites:**
- Must have an open PR for the current branch
- GitHub Actions workflow must be configured to respond to `deploy --target=Container` comment

#### `make list-deploys`
Lists all current and recent deployments with their status.

```bash
make list-deploys
```

**What it shows:**
1. **In-Progress Deployments:**
   - Currently running deployments
   - Workflow name and branch
   - Time since started
   - Direct link to run

2. **Recently Completed Deployments:**
   - Last 5 completed deployments
   - Status (✅ success, ❌ failure, ⏸️ cancelled)
   - Workflow name and branch
   - Time since completed
   - Direct link to run

**Example output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Current Deployments Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏳ In-Progress Deployments:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏳ Deploy Container (feat/my-feature)
   Workflow: Deploy to Dev
   Started: 2m ago
   URL: https://github.com/owner/repo/actions/runs/123456

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Recently Completed Deployments:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Deploy Container (main) - success
   Workflow: Deploy to Dev
   Completed: 1h ago
   URL: https://github.com/owner/repo/actions/runs/123455

❌ Deploy Container (feat/bugfix) - failure
   Workflow: Deploy to Dev
   Completed: 3h ago
   URL: https://github.com/owner/repo/actions/runs/123454

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 Useful commands:
   - Watch a deployment: gh run watch <run-id>
   - View run details: gh run view <run-id>
   - Cancel a run: gh run cancel <run-id>
   - Trigger new deploy: make deploy-dev
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Use cases:**
- Check deployment status before triggering new deploy
- Monitor ongoing deployments
- Review deployment history
- Identify failed deployments for debugging

---

### Installation & Setup

#### `make install`
Installs all dependencies for both UI and Python service.

```bash
make install
```

This is equivalent to running both `make install-ui` and `make install-service`.

#### `make install-ui`
Installs only the UI (Node.js) dependencies.

```bash
make install-ui
```

**What it does:**
- Runs `npm install` in the `ui` directory
- Installs all packages listed in `package.json`

#### `make install-service`
Installs only the Python service dependencies.

```bash
make install-service
```

**What it does:**
- Creates a Python virtual environment in `service/venv`
- Installs all packages from `requirements.txt` and `requirements-test.txt`

---

### Starting the Application

#### `make start`
Starts all services with a **mock backend** (recommended for UI development).

```bash
make start
```

**What it does:**
- Kills any processes on ports 3000, 3006, 8000, 8001
- Opens a new iTerm2 window with tmux session
- Starts three services in separate panes:
  - **Pane 1**: React UI (port 3000)
  - **Pane 2**: Mock backend (port 8001)
  - **Pane 3**: Caddy reverse proxy (port 3006)

**No configuration required!** This mode works without any environment variables.

#### `make start-py`
Starts all services with **Python backend (mocked external services)**.

```bash
make start-py
```

**What it does:**
- Same as `make start`, but runs the Python backend with mocked Databricks/AWS connections
- Uses `tests/testbotofbots/localmock/main.py`

**Requires:** Environment variables configured (see README.md)

#### `make start-py-real`
Starts all services with **Python backend (real connections)**.

```bash
make start-py-real
```

**What it does:**
- Same as `make start`, but runs the Python backend with real Databricks/AWS connections
- Uses `src/botofbots/app.py`

**Requires:** Full environment configuration with valid credentials

---

### Stopping Services

#### `make stop`
Stops all running services and cleans up resources.

```bash
make stop
```

**What it does:**
- Kills all processes on ports 3000, 3006, 8000, 8001
- Terminates the tmux session named 'startup'
- Displays confirmation message

---

### Service Management

#### `make status`
Checks the status of all services.

```bash
make status
```

**Output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Service Status Check
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Checking ports...
✓ Port 3000: IN USE
✓ Port 3006: IN USE
✓ Port 8000: IN USE
✗ Port 8001: FREE

Checking tmux session...
✓ tmux session 'startup': RUNNING
  Use 'tmux attach -t startup' to connect
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

### Testing

#### `make test-ui`
Runs all UI unit tests using Vitest.

```bash
make test-ui
```

**What it does:**
- Runs `npm test` in the `ui` directory
- Executes all test files matching `*.test.ts`, `*.test.tsx`, `*.spec.ts`, `*.spec.tsx`
- Uses jsdom environment for React Testing Library

#### `make test-e2e`
Runs end-to-end tests using Playwright.

```bash
make test-e2e
```

**What it does:**
- Automatically starts the application with mock backend
- Runs all E2E tests in `ui/tests/e2e/specs/`
- Tests the full user workflow in a real browser
- Generates HTML report in `ui/tests/e2e/playwright-report/`

**Prerequisites:**
- First time only: Install Playwright browsers
  ```bash
  cd Container/ui
  npx playwright install
  ```

**Example output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Running E2E Tests (Playwright)...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Running 10 tests using 1 worker

  ✓ auto-mode-default.spec.ts:25:3 › should enable auto mode by default
  ✓ auto-mode-default.spec.ts:37:3 › should display auto mode disclaimer
  ...

10 passed (45s)
```

#### `make test-e2e-ui`
Runs E2E tests with interactive UI mode for debugging.

```bash
make test-e2e-ui
```

**What it does:**
- Opens Playwright's interactive test runner
- Allows you to step through tests, inspect DOM, view timeline
- Helpful for debugging failing tests

**Use when:**
- Debugging test failures
- Understanding test flow
- Developing new tests

#### `make test-e2e-debug`
Runs E2E tests in debug mode with browser visible.

```bash
make test-e2e-debug
```

**What it does:**
- Runs tests with browser visible (headed mode)
- Opens Playwright Inspector
- Pauses on failures
- Allows step-by-step debugging

**Use when:**
- Tests are failing and you need to see what's happening
- Investigating browser behavior
- Debugging selectors

#### `make test-e2e-report`
Opens the HTML report from the last test run.

```bash
make test-e2e-report
```

**What it does:**
- Opens the Playwright HTML report in your default browser
- Shows detailed test results with screenshots and traces
- Allows you to inspect failures

#### `make test-e2e-codegen`
Opens Playwright Codegen for recording new tests.

```bash
# First, start the application
make start

# In a new terminal, run codegen
make test-e2e-codegen
```

**What it does:**
- Opens a browser and the Playwright Inspector
- Records your interactions and generates test code
- Helps you create new tests quickly

**How to use:**
1. Start the app: `make start`
2. In a new terminal: `make test-e2e-codegen`
3. Interact with your app in the opened browser
4. Copy the generated code from Playwright Inspector
5. Paste into a new test file in `ui/tests/e2e/specs/`

#### `make test-service`
Runs all Python service tests using pytest.

```bash
make test-service
```

**What it does:**
- Activates the Python virtual environment
- Sets `PYTHONPATH=src:tests`
- Runs `pytest -vv` in the `service` directory

#### `make test`
Runs all tests (UI unit tests + Service tests).

```bash
make test
```

**Note:** This does NOT include E2E tests. Run `make test-e2e` separately for E2E testing.

---

### Code Quality

#### `make lint-ui`
Lints the UI code.

```bash
make lint-ui
```

**What it does:**
- Runs `npm run lint` in the `ui` directory
- Checks TypeScript/JavaScript code for style and errors

#### `make lint-service`
Lints the Python service code.

```bash
make lint-service
```

**What it does:**
- Activates the Python virtual environment
- Runs `flake8 src tests` (if installed)

---

### Maintenance

#### `make clean`
Removes all build artifacts and dependencies.

```bash
make clean
```

**What it does:**
- Removes `node_modules`, `build`, `dist`, `.vite` from `ui`
- Removes `venv`, `__pycache__`, `.pytest_cache`, `.coverage`, `htmlcov` from `service`
- Finds and removes all `__pycache__` and `.pytest_cache` directories
- Removes all `*.egg-info` directories

**Warning:** After running this, you'll need to run `make install` again.

#### `make help`
Displays all available commands with descriptions.

```bash
make help
```

or simply:

```bash
make
```

---

## Using tmux

When you run `make start`, `make start-py`, or `make start-py-real`, services are launched in a tmux session. Here's how to control it:

### tmux Keyboard Shortcuts

| Command | Action |
|---------|--------|
| `Ctrl+B` then `Arrow Keys` | Navigate between panes |
| `Ctrl+B` then `D` | Detach from tmux (services keep running) |
| `Ctrl+B` then `[` | Enter scroll mode (use arrow keys, press `q` to exit) |
| `Ctrl+C` in a pane | Stop the service in that pane |

### Reattaching to tmux

If you detached from tmux or closed your terminal:

```bash
tmux attach -t startup
```

### Viewing Individual Pane Logs

1. Attach to the tmux session: `tmux attach -t startup`
2. Navigate to the desired pane: `Ctrl+B` then arrow keys
3. The pane will show real-time logs for that service

---

## Port Reference

| Port | Service | Direct Access | Used By |
|------|---------|---------------|---------|
| 3000 | React UI Dev Server | http://localhost:3000 | Development (hot reload) |
| 3006 | Caddy HTTPS Proxy | https://localhost:3006 | **Recommended access point** |
| 8000 | Python Backend API | http://localhost:8000 | Backend service |
| 8001 | Mock Backend | http://localhost:8001 | Mock mode only |

**Best Practice:** Access the application via Caddy at https://localhost:3006 to avoid CORS issues.

---

## Run Modes Comparison

| Mode | Command | Backend | External Services | Environment Config | Use Case |
|------|---------|---------|-------------------|-------------------|----------|
| **Mock** | `make start` | Node.js mock | None | Not required | Quick UI development |
| **Python (Mocked)** | `make start-py` | Python | Mocked | Required | Backend development & testing |
| **Python (Real)** | `make start-py-real` | Python | Real connections | Required + credentials | Integration testing |

---

## Common Workflows

### First Time Setup
```bash
cd Container
make install
make start
# Visit https://localhost:3006
```

### Daily Development (UI Work)
```bash
cd Container
make start  # Fast startup with mocks
# Make changes to UI code (hot reload enabled)
make stop   # When done
```

### Backend Development
```bash
cd Container
make start-py  # Python backend with mocks
# Make changes to backend code
# Restart backend pane with Ctrl+C then up arrow + enter
make stop      # When done
```

### Running Tests
```bash
cd Container
make test-ui
make test-service
```

### Code Quality Check
```bash
cd Container
make lint-ui
make lint-service
```

### Cleanup and Fresh Start
```bash
cd Container
make clean
make install
make start
```

---

## Troubleshooting

### Ports Already in Use

**Problem:** Error message about ports being in use.

**Solution:**
```bash
make stop
# or manually:
lsof -ti tcp:3000 | xargs kill
lsof -ti tcp:3006 | xargs kill
lsof -ti tcp:8000 | xargs kill
lsof -ti tcp:8001 | xargs kill
```

### tmux Session Already Exists

**Problem:** Error that tmux session 'startup' already exists.

**Solution:**
```bash
# Option 1: Attach to existing session
tmux attach -t startup

# Option 2: Kill and start fresh
tmux kill-session -t startup
make start
```

### Services Won't Start

**Problem:** Services fail to start.

**Solutions:**

1. **Check prerequisites:**
   ```bash
   python3 --version  # Should be 3.12+
   node --version     # Should be 18+
   tmux -V
   caddy version
   ```

2. **Reinstall dependencies:**
   ```bash
   make clean
   make install
   ```

3. **Check environment variables (for Python modes):**
   ```bash
   env | grep -E "(DATABRICKS|OKTA|JWT)"
   ```

### Make Command Not Found

**Problem:** `make: command not found`

**Solution:**
```bash
# Install make (usually pre-installed on macOS)
xcode-select --install
```

### iTerm2 Not Opening

**Problem:** Script doesn't open iTerm2 window.

**Solutions:**
- Ensure iTerm2 is installed: https://iterm2.com/
- Grant Terminal/iTerm2 automation permissions in System Preferences → Security & Privacy → Automation
- Try running the script directly: `./tools/startup.sh`

---

## Advanced Usage

### Custom Script Flags

You can also use the startup script directly with custom flags:

```bash
# From Container directory
./tools/startup.sh              # Mock backend
./tools/startup.sh --py         # Python backend (mocked)
./tools/startup.sh --py --no-mocks  # Python backend (real)
./tools/startup.sh --stop       # Stop services
```

### Modifying the Makefile

The Makefile is located at `Container/Makefile`. You can customize targets or add new ones as needed. Each target is well-commented for easy understanding.

### Environment-Specific Configuration

For different environments (dev, test, prod), you can:
1. Create separate `.env` files
2. Source them before running make commands
3. Or modify the Makefile to load environment-specific configurations

---

## Tips for Productive Development

1. **Use `make start` for UI work** - Fastest startup, no configuration needed
2. **Monitor all services in one window** - tmux panes show real-time logs
3. **Detach from tmux** (`Ctrl+B` then `D`) - Services keep running, work in other terminals
4. **Check status frequently** - Run `make status` before starting to avoid conflicts
5. **Use Caddy proxy** (https://localhost:3006) - Avoids CORS issues during development
6. **Run tests regularly** - `make test-ui` and `make test-service` catch issues early

---

## Related Documentation

- [README.md](README.md) - Main setup instructions and environment variables
- [startup.sh](tools/startup.sh) - Startup script with detailed comments
- [start_caddy.sh](tools/start_caddy.sh) - Caddy proxy startup script

---

**Last Updated:** January 2026

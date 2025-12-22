# DevUp Environment Configuration - Setup & Usage Guide

## 🚀 Quick Start (2 minutes)

### Step 1: Edit Your Shell Configuration

```bash
# Open your shell config file
nano ~/.zshrc    # For zsh (macOS default)
# OR
nano ~/.bashrc   # For bash
```

### Step 2: Add the Environment Variable

Add this line at the end of the file:

```bash
export DEVUP_DEFAULT_PROJECT="/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container"
```

### Step 3: Reload Shell

```bash
source ~/.zshrc  # For zsh
# OR
source ~/.bashrc # For bash
```

### Step 4: Verify Setup

```bash
# Check variable is set
echo $DEVUP_DEFAULT_PROJECT

# Output should be your project path
```

### Step 5: Build DevUp

```bash
cd /Users/everton.jesus/workspace/devup_v2
go build -o build/devup
```

### Step 6: Test It Works

```bash
# Run from anywhere
cd /tmp
devup start

# Should start your engineering supervisor project!
```

---

## 📋 Common Usage Patterns

### Pattern 1: Default Usage

Once environment variable is set:

```bash
# Run commands from anywhere
devup start      # Start services
devup install    # Install dependencies
devup setup      # Setup environment
devup status     # Check status
devup stop       # Stop services
```

### Pattern 2: Force Local Directory

```bash
# Use local devup.yaml instead of default project
devup start -l

# Or longer form:
devup start --local
```

### Pattern 3: Explicit Config Path

```bash
# Use specific config file
devup start -c /path/to/devup.yaml

# Or longer form:
devup start --config /path/to/devup.yaml
```

### Pattern 4: With Application Selection

```bash
# Start specific app from default project
devup start -a engineering-supervisor

# Or with local flag:
devup start -l -a my-app

# Or all together:
devup start -c /custom/devup.yaml -a myapp -v
```

---

## 🔄 Switching Between Projects

### Method 1: Change Environment Variable

```bash
# Project A
export DEVUP_DEFAULT_PROJECT=/path/to/project-a
devup start

# Project B
export DEVUP_DEFAULT_PROJECT=/path/to/project-b
devup start
```

### Method 2: Create Shell Functions

```bash
# Add to ~/.zshrc
project-eng-sup() {
    export DEVUP_DEFAULT_PROJECT="/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container"
    cd $DEVUP_DEFAULT_PROJECT
    echo "✅ Switched to Engineering Supervisor Agent"
}

project-other() {
    export DEVUP_DEFAULT_PROJECT="/path/to/other/project"
    cd $DEVUP_DEFAULT_PROJECT
    echo "✅ Switched to Other Project"
}

# Use it:
project-eng-sup
devup start

# Switch
project-other
devup start
```

### Method 3: Use Shell Aliases

```bash
# Add to ~/.zshrc
alias start-eng-sup="export DEVUP_DEFAULT_PROJECT='/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container' && devup start"
alias status-eng-sup="export DEVUP_DEFAULT_PROJECT='/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container' && devup status"

# Use it:
start-eng-sup
status-eng-sup
```

---

## 🐛 Troubleshooting

### Issue: "config file not found"

```bash
# Check if environment variable is set
echo $DEVUP_DEFAULT_PROJECT

# If empty, reload your shell config:
source ~/.zshrc

# If it shows a path, verify the file exists:
ls $DEVUP_DEFAULT_PROJECT/devup.yaml

# If file doesn't exist, update the environment variable to the correct path
nano ~/.zshrc  # Fix the path
source ~/.zshrc
```

### Issue: "still not working after editing ~/.zshrc"

```bash
# Option 1: Reload shell config
source ~/.zshrc

# Option 2: Close and reopen terminal

# Option 3: Restart your shell explicitly
exec zsh
```

### Issue: "want to ignore environment variable temporarily"

```bash
# Use the -l flag to force local directory
cd /path/to/other/project
devup start -l

# Or use explicit path
devup start -c /explicit/path/devup.yaml
```

### Issue: "environment variable set but not being used"

```bash
# Check if it's actually set
echo $DEVUP_DEFAULT_PROJECT

# Check if -l flag is being used (it would override env var)
# Remove -l flag from your command

# Verify devup.yaml exists in the project
ls $DEVUP_DEFAULT_PROJECT/devup.yaml

# Try explicit path for debugging
devup start -c $DEVUP_DEFAULT_PROJECT/devup.yaml -v
```

---

## 📚 Command Reference

### All Available Commands

```bash
devup start [FLAGS]     # Start services in specified mode
devup stop [FLAGS]      # Stop all services
devup install [FLAGS]   # Install dependencies
devup setup [FLAGS]     # Setup environment and variables
devup clean [FLAGS]     # Clean resources and directories
devup status [FLAGS]    # Show service status
devup env [FLAGS]       # Display environment variables
devup list [FLAGS]      # List available applications
```

### All Available Flags

```bash
-c, --config string    # Config file path (overrides everything)
-l, --local           # Use local directory, ignore DEVUP_DEFAULT_PROJECT
-a, --app string      # Application name to manage
-v, --verbose         # Verbose output
-m, --mode string     # Running mode (for start command)
-h, --help            # Show help
```

### Flag Priority

```
1. -c /path/config.yaml    ← Explicit path (highest priority)
2. -l                       ← Force local directory
3. $DEVUP_DEFAULT_PROJECT   ← Environment variable
4. ./devup.yaml             ← Default search (lowest priority)
```

---

## 🎯 Real World Example

### Scenario: Daily Development with Engineering Supervisor Agent

```bash
# 1. One-time setup in ~/.zshrc
export DEVUP_DEFAULT_PROJECT="/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container"

# 2. Daily usage - run from anywhere
devup start    # Start all services
devup status   # Check if running

# During development
devup status   # Check what's running
devup stop     # Stop services
devup start    # Restart

# Before committing
devup install  # Ensure dependencies fresh
devup status   # Verify all running
```

### Scenario: Multiple Projects

```bash
# Add project functions to ~/.zshrc
proj-a() {
    export DEVUP_DEFAULT_PROJECT=/path/to/project-a
    cd $DEVUP_DEFAULT_PROJECT
}

proj-b() {
    export DEVUP_DEFAULT_PROJECT=/path/to/project-b
    cd $DEVUP_DEFAULT_PROJECT
}

# In daily workflow:
proj-a
devup start     # Start project A

# Switch to project B
proj-b
devup start     # Start project B
```

---

## ✅ Verification Checklist

After setup, verify everything works:

```bash
# 1. Environment variable is set
[ -n "$DEVUP_DEFAULT_PROJECT" ] && echo "✅ Env var set" || echo "❌ Env var not set"

# 2. Project directory exists
[ -d "$DEVUP_DEFAULT_PROJECT" ] && echo "✅ Directory exists" || echo "❌ Directory not found"

# 3. devup.yaml exists
[ -f "$DEVUP_DEFAULT_PROJECT/devup.yaml" ] && echo "✅ Config file exists" || echo "❌ Config file not found"

# 4. devup command works
devup list && echo "✅ Devup works" || echo "❌ Devup failed"

# 5. Can start services
devup start && echo "✅ Services started" || echo "❌ Failed to start"
```

---

## 📞 Need Help?

Refer to:
- **Quick questions** → `DEVUP_ENV_QUICK_REF.md`
- **Detailed setup** → `DEVUP_ENV_CONFIG.md`
- **Technical details** → `CODE_CHANGES_DETAIL.md`
- **Verification** → `VERIFICATION_CHECKLIST.md`

---

## 🎉 You're All Set!

Once environment variable is configured:

✅ Run `devup start` from anywhere  
✅ No need to specify config file path  
✅ Switch projects easily  
✅ Override with `-l` or `-c` when needed  
✅ All devup commands work seamlessly  

**Happy developing!** 🚀

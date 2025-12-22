# DevUp Environment Configuration - Quick Reference

## TL;DR

### Setup (One-time)

```bash
# Add to ~/.zshrc
export DEVUP_DEFAULT_PROJECT="/path/to/your/engineering-supervisor-agent/Container"

# Reload
source ~/.zshrc
```

### Usage

```bash
# Run from anywhere - uses DEVUP_DEFAULT_PROJECT
devup start
devup install
devup setup
devup status

# Force local directory instead
devup start -l

# Use specific config
devup start -c /path/to/devup.yaml
```

## Flag Reference

| Flag | Short | Type | Purpose |
|------|-------|------|---------|
| `--config` | `-c` | string | Override config file path (highest priority) |
| `--local` | `-l` | boolean | Use local dir, ignore `DEVUP_DEFAULT_PROJECT` |
| `--app` | `-a` | string | Select specific app |
| `--verbose` | `-v` | boolean | Show detailed output |

## Environment Variable

```bash
DEVUP_DEFAULT_PROJECT=/path/to/project
```

When set, devup will automatically use `$DEVUP_DEFAULT_PROJECT/devup.yaml`

## Priority Order

1. `-c /path/config.yaml` (explicit, highest)
2. `-l` flag (force local)
3. `$DEVUP_DEFAULT_PROJECT` (environment var)
4. Default search paths (lowest)

## Examples

### Set Default Project for Engineering Supervisor

```bash
# Edit shell config
nano ~/.zshrc

# Add:
export DEVUP_DEFAULT_PROJECT="/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container"

# Save and reload
source ~/.zshrc

# Now use from anywhere:
devup start      # Works!
devup install    # Works!
devup setup      # Works!
```

### Multiple Projects

```bash
# Quick switch function
# Add to ~/.zshrc
project-eng-sup() {
    export DEVUP_DEFAULT_PROJECT="/Users/everton.jesus/workspace/projects/gilead/engineering-supervisor-agent/Container"
    cd $DEVUP_DEFAULT_PROJECT
}

# Use it:
project-eng-sup
devup start
```

### Override Default

```bash
# Default is set to Project A
export DEVUP_DEFAULT_PROJECT=/path/to/project-a

# Use Project B temporarily
devup start -c /path/to/project-b/devup.yaml

# Or force local
cd /path/to/project-c
devup start -l   # Uses ./devup.yaml
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "config file not found" | Check `echo $DEVUP_DEFAULT_PROJECT` and verify path exists |
| env var not working | Run `source ~/.zshrc` to reload shell config |
| Want to use local config | Use `-l` flag: `devup start -l` |
| Need specific config | Use `-c` flag: `devup start -c /path/devup.yaml` |

## All Commands Support This

```bash
devup start [FLAGS]     # Start services
devup stop [FLAGS]      # Stop services  
devup install [FLAGS]   # Install dependencies
devup setup [FLAGS]     # Setup environment
devup clean [FLAGS]     # Clean resources
devup status [FLAGS]    # Show status
devup env [FLAGS]       # Show environment variables
devup list [FLAGS]      # List applications
```

## Verification

```bash
# Check environment variable is set
echo $DEVUP_DEFAULT_PROJECT

# Verify file exists
ls $DEVUP_DEFAULT_PROJECT/devup.yaml

# Run with verbose to see config path
devup list -v
```

# DevUp Environment Configuration

## Overview

DevUp now supports configuring a default project directory via an environment variable, allowing you to easily manage a primary project without specifying the configuration file path every time.

## Features

### 1. `DEVUP_DEFAULT_PROJECT` Environment Variable

Set an environment variable to specify the default project directory:

```bash
export DEVUP_DEFAULT_PROJECT=/path/to/your/project
```

When this variable is set and no explicit config file is specified with `-c`, devup will automatically look for `devup.yaml` in that directory.

### 2. Priority Order

The configuration file resolution follows this priority:

1. **Explicit config path** (`-c` flag) - Highest priority
   ```bash
   devup start -c /custom/path/devup.yaml
   ```

2. **Local flag** (`-l` flag) - Forces local directory search, ignores environment variable
   ```bash
   devup start -l  # Uses devup.yaml in current directory
   ```

3. **Environment variable** (`DEVUP_DEFAULT_PROJECT`) - Used if no `-c` or `-l` specified
   ```bash
   export DEVUP_DEFAULT_PROJECT=/path/to/project
   devup start  # Uses /path/to/project/devup.yaml
   ```

4. **Local search paths** - Default fallback
   - `devup.yaml`
   - `.devup.yaml`
   - `config/devup.yaml`
   - `.config/devup.yaml`
   - `~/.devup.yaml`
   - `~/.config/devup/devup.yaml`

## Setup Instructions

### macOS/Linux

Add to your shell profile (`~/.zshrc`, `~/.bashrc`, or `~/.bash_profile`):

```bash
export DEVUP_DEFAULT_PROJECT="/Users/yourusername/workspace/projects/your-project"
```

Then reload your shell:
```bash
source ~/.zshrc
```

### Verify Setup

Check that the environment variable is set:
```bash
echo $DEVUP_DEFAULT_PROJECT
```

## Usage Examples

### Example 1: Using Default Project

```bash
# Set your default project
export DEVUP_DEFAULT_PROJECT=$HOME/workspace/my-app

# Navigate anywhere and run devup
cd /tmp
devup start        # Uses $DEVUP_DEFAULT_PROJECT/devup.yaml
devup status       # Uses $DEVUP_DEFAULT_PROJECT/devup.yaml
devup install      # Uses $DEVUP_DEFAULT_PROJECT/devup.yaml
```

### Example 2: Override with Local Config

```bash
export DEVUP_DEFAULT_PROJECT=$HOME/workspace/my-app

# Force use of local directory config
cd $HOME/workspace/other-app
devup start -l     # Uses ./devup.yaml (current directory)
```

### Example 3: Override with Explicit Path

```bash
export DEVUP_DEFAULT_PROJECT=$HOME/workspace/my-app

# Use specific config file
cd /tmp
devup start -c /custom/path/devup.yaml  # Uses specified file
```

### Example 4: Your Project

```bash
# Add to ~/.zshrc
export DEVUP_DEFAULT_PROJECT="/path/to/your/project"

# Now you can run from anywhere:
devup start                    # Uses DEVUP_DEFAULT_PROJECT
devup install                  # Uses DEVUP_DEFAULT_PROJECT
devup setup                     # Uses DEVUP_DEFAULT_PROJECT
devup status                    # Uses DEVUP_DEFAULT_PROJECT

# Or override when needed:
devup start -l                 # Use local dir instead
devup start -c ./devup.yaml    # Use specific file
```

## Command Reference

All devup commands support the configuration resolution:

```bash
devup start [flags]     # Start services
devup stop [flags]      # Stop services
devup install [flags]   # Install dependencies
devup setup [flags]     # Setup environment
devup status [flags]    # Show status
devup clean [flags]     # Clean resources
devup env [flags]       # Show environment variables
devup list [flags]      # List applications
```

### Available Flags

- `-c, --config string` - Specify config file path (overrides everything)
- `-l, --local` - Use local directory config, ignore `DEVUP_DEFAULT_PROJECT`
- `-a, --app string` - Select specific application
- `-v, --verbose` - Enable verbose output

## Tips & Best Practices

### Multiple Projects

If you work with multiple projects, you can switch between them:

```bash
# Project A
export DEVUP_DEFAULT_PROJECT=$HOME/workspace/project-a
devup start

# Switch to Project B
export DEVUP_DEFAULT_PROJECT=$HOME/workspace/project-b
devup start
```

Or use project-specific shell functions:

```bash
# Add to ~/.zshrc
project-a() {
    export DEVUP_DEFAULT_PROJECT=$HOME/workspace/project-a
    cd $DEVUP_DEFAULT_PROJECT
}

project-b() {
    export DEVUP_DEFAULT_PROJECT=$HOME/workspace/project-b
    cd $DEVUP_DEFAULT_PROJECT
}

# Use it:
project-a      # Switches to Project A with env var set
devup start     # Uses Project A's config
```

### Team Collaboration

For team projects, document the setup in your project's README:

```markdown
## Setup

1. Set your default project directory:
   ```bash
   export DEVUP_DEFAULT_PROJECT=/path/to/this/project
   ```

2. Run devup commands:
   ```bash
   devup install
   devup setup
   devup start
   ```
```

### Debugging Configuration Resolution

Use the verbose flag to see which config file is being used:

```bash
devup start -v   # Shows verbose output including config path
devup list -v    # Shows configuration source
```

## Environment Variable Persistence

### Permanent Setup (Recommended)

Edit your shell configuration file:

```bash
# For zsh (macOS default)
nano ~/.zshrc

# For bash
nano ~/.bashrc

# Add this line:
export DEVUP_DEFAULT_PROJECT="/path/to/your/project"
```

Then reload:
```bash
source ~/.zshrc  # or source ~/.bashrc
```

### Temporary Setup

Set for current session only:

```bash
export DEVUP_DEFAULT_PROJECT="/path/to/your/project"
```

This variable will be cleared when you close the terminal.

## Troubleshooting

### Configuration file not found

```bash
# Check if environment variable is set
echo $DEVUP_DEFAULT_PROJECT

# Verify the path exists
ls $DEVUP_DEFAULT_PROJECT/devup.yaml

# Use explicit path as workaround
devup start -c /correct/path/devup.yaml
```

### Environment variable not applying

```bash
# Reload shell configuration
source ~/.zshrc

# Verify variable is set
echo $DEVUP_DEFAULT_PROJECT

# Restart terminal
```

### Want to use local config instead of default project

```bash
# Use the -l (local) flag
devup start -l
```

## See Also

- [DevUp Documentation](README.md)
- [Makefile Documentation](MAKEFILE.md)

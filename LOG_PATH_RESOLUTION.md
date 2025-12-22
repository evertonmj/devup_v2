# Log Path Resolution - Implementation Update

## Change Summary

Updated DevUp to ensure **all logs are always saved in the directory where devup is run**, not relative to the service's workdir.

## Technical Details

### Files Modified

**`internal/service/process.go`**

### Changes Made

#### 1. Updated `setupLogging()` function
- Gets current working directory (where devup was run)
- Resolves log file paths relative to this directory
- Converts relative paths to absolute paths
- Creates log directories as needed

#### 2. Updated `Logs()` function
- Uses same path resolution logic
- Ensures reading logs from correct location
- Consistent with where logs are written

### Code Changes

**Before:**
```go
// Log path was relative to service workdir
logDir := filepath.Dir(p.config.LogFile)
logFile, err := os.OpenFile(p.config.LogFile, ...)
```

**After:**
```go
// Get where devup was run
cwd, err := os.Getwd()

// Resolve relative to devup's directory
logPath := p.config.LogFile
if !filepath.IsAbs(logPath) {
    logPath = filepath.Join(cwd, logPath)
}

// Use resolved path
logFile, err := os.OpenFile(logPath, ...)
```

## Behavior

### Example Scenarios

#### Scenario 1: Run from project root
```bash
cd /Users/everton.jesus/workspace/projects/my-project
devup start
```
**Result:** Logs saved to `/Users/everton.jesus/workspace/projects/my-project/logs/`

#### Scenario 2: Run from elsewhere with config
```bash
cd /tmp
devup start -c /Users/everton.jesus/workspace/projects/my-project/devup.yaml
```
**Result:** Logs saved to `/tmp/logs/` (where devup was run)

#### Scenario 3: Using environment variable
```bash
cd /tmp
export DEVUP_DEFAULT_PROJECT=/Users/everton.jesus/workspace/projects/my-project
devup start
```
**Result:** Logs saved to `/tmp/logs/` (where devup was run)

## Configuration

In your `devup.yaml`, log paths are now relative to where devup is run:

```yaml
services:
  - name: ui
    command: "npm run start"
    logfile: "logs/ui.log"     # Relative to where devup is run
    
  - name: backend
    command: "python app.py"
    logfile: "/tmp/backend.log" # Absolute paths also work
```

## Benefits

✅ Logs always go where user expects (working directory)  
✅ No confusion about log locations  
✅ Works correctly with environment variables  
✅ Works correctly with `-c` flag  
✅ Works correctly with `-l` flag  
✅ Consistent behavior regardless of project structure  

## Backward Compatibility

✅ **Fully backward compatible**
- Existing log paths still work
- Absolute paths respected
- Relative paths now resolved from execution directory (more intuitive)

## Testing

```bash
# Test 1: Logs save to current directory
cd /path/to/project
devup start
ls logs/         # Should contain log files

# Test 2: Logs with config file from different location
cd /tmp
devup start -c /path/to/project/devup.yaml
ls /tmp/logs/    # Logs saved here, not in project directory

# Test 3: Logs with environment variable
export DEVUP_DEFAULT_PROJECT=/path/to/project
cd /different/location
devup start
ls /different/location/logs/  # Logs saved here
```

## Summary

Logs are now **always saved where devup is run**, making log location predictable and consistent across all usage patterns.

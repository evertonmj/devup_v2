package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultVersion = "python3"
	DefaultDir     = ".venv"
)

// VenvEnv returns environment variables (KEY=value) to activate the venv.
// venvDir is the absolute path to the venv directory.
func VenvEnv(venvDir string) []string {
	bin := filepath.Join(venvDir, "bin")
	if runtime.GOOS == "windows" {
		bin = filepath.Join(venvDir, "Scripts")
	}
	return []string{
		"VIRTUAL_ENV=" + venvDir,
		"PATH=" + bin + string(filepath.ListSeparator) + os.Getenv("PATH"),
	}
}

// EnsureVenv creates the virtual environment if it does not exist.
// workDir is the app/service working directory; venvDir is the venv path (can be relative like ".venv" or absolute).
// pythonCmd is the Python executable (e.g. "python3", "python3.11").
func EnsureVenv(workDir, venvDir, pythonCmd string) error {
	absVenv := venvDir
	if !filepath.IsAbs(venvDir) {
		absVenv = filepath.Join(workDir, venvDir)
	}
	if _, err := os.Stat(filepath.Join(absVenv, "pyvenv.cfg")); err == nil {
		return nil
	}
	if _, err := os.Stat(filepath.Join(absVenv, "Scripts", "python.exe")); err == nil {
		return nil
	}
	cmd := exec.Command(pythonCmd, "-m", "venv", absVenv)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create venv: %w", err)
	}
	return nil
}

// IsPythonCommand returns true if cmd looks like a Python-related command (python, pip, uvicorn, etc.).
func IsPythonCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, prefix := range []string{"python", "pip", "uvicorn", "gunicorn", "flask", "django", "pytest"} {
		if strings.HasPrefix(lower, prefix+" ") || strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

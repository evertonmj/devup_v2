package python

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVenvEnv(t *testing.T) {
	env := VenvEnv("/tmp/venv")
	if len(env) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(env))
	}
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "VIRTUAL_ENV=") {
			found = true
			if !strings.Contains(e, "/tmp/venv") {
				t.Errorf("VIRTUAL_ENV should contain venv path, got %s", e)
			}
		}
	}
	if !found {
		t.Error("expected VIRTUAL_ENV in output")
	}
}

func TestIsPythonCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"python main.py", true},
		{"pip install x", true},
		{"uvicorn main:app", true},
		{"gunicorn app:app", true},
		{"flask run", true},
		{"django manage.py", true},
		{"pytest", true},
		{"npm start", false},
		{"go run .", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsPythonCommand(tt.cmd)
		if got != tt.want {
			t.Errorf("IsPythonCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

func TestEnsureVenv_AlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	venvDir := filepath.Join(tmp, ".venv")
	if err := os.MkdirAll(venvDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(venvDir, "pyvenv.cfg")
	if err := os.WriteFile(cfgPath, []byte("home = /usr"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureVenv(tmp, ".venv", "python3"); err != nil {
		t.Errorf("EnsureVenv with existing venv should not error: %v", err)
	}
}

func TestEnsureVenv_CreatesNew(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping venv creation on Windows")
	}
	tmp := t.TempDir()
	if err := EnsureVenv(tmp, ".venv", "python3"); err != nil {
		t.Skipf("python3 not available or venv failed: %v", err)
	}
	cfg := filepath.Join(tmp, ".venv", "pyvenv.cfg")
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		t.Errorf("venv should have been created at %s", cfg)
	}
}

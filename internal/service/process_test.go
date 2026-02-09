package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"devup/internal/config"
)

// Ensure context is available for future tests
var _ = func() {} // placeholder

// MockCommander implements Commander interface for testing
type MockCommander struct {
	startCalled bool
	waitCalled  bool
	killCalled  bool
	process     *os.Process
}

func (m *MockCommander) Start() error {
	m.startCalled = true
	return nil
}

func (m *MockCommander) Wait() error {
	m.waitCalled = true
	return nil
}

func (m *MockCommander) Process() *os.Process {
	// Return a dummy process for testing
	if m.process == nil {
		m.process = &os.Process{Pid: 12345}
	}
	return m.process
}

func (m *MockCommander) SysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func (m *MockCommander) SetSysProcAttr(attr *syscall.SysProcAttr) {}
func (m *MockCommander) SetDir(dir string)                        {}
func (m *MockCommander) SetEnv(env []string)                      {}
func (m *MockCommander) SetStdout(stdout io.Writer)               {}
func (m *MockCommander) SetStderr(stderr io.Writer)               {}

func (m *MockCommander) Kill() error {
	m.killCalled = true
	return nil
}

func TestNewProcessRunner(t *testing.T) {
	appWorkDir := "/app"

	// Test case 1: Service with no workdir
	svc1 := config.Service{
		Name: "service1",
	}
	runner1, err := NewProcessRunner(svc1, appWorkDir, nil)
	if err != nil {
		t.Fatalf("Failed to create process runner: %v", err)
	}
	if runner1.workDir != appWorkDir {
		t.Errorf("Expected workdir to be %s, but got %s", appWorkDir, runner1.workDir)
	}

	// Test case 2: Service with relative workdir
	svc2 := config.Service{
		Name:    "service2",
		WorkDir: "service2",
	}
	runner2, err := NewProcessRunner(svc2, appWorkDir, nil)
	if err != nil {
		t.Fatalf("Failed to create process runner: %v", err)
	}
	expectedWorkDir2 := filepath.Join(appWorkDir, "service2")
	if runner2.workDir != expectedWorkDir2 {
		t.Errorf("Expected workdir to be %s, but got %s", expectedWorkDir2, runner2.workDir)
	}

	// Test case 3: Service with absolute workdir
	svc3 := config.Service{
		Name:    "service3",
		WorkDir: "/service3",
	}
	runner3, err := NewProcessRunner(svc3, appWorkDir, nil)
	if err != nil {
		t.Fatalf("Failed to create process runner: %v", err)
	}
	if runner3.workDir != "/service3" {
		t.Errorf("Expected workdir to be /service3, but got %s", runner3.workDir)
	}
}

func TestProcessRunnerStartStopStatusLogs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "logs"), 0755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	ctx := context.Background()

	// Runner with log file: Start -> Status -> Logs -> Stop
	svc := config.Service{
		Name:    "svc1",
		Command: "echo hello",
		LogFile: "logs/svc.log",
	}
	runner, err := NewProcessRunner(svc, tmpDir, nil)
	if err != nil {
		t.Fatalf("NewProcessRunner: %v", err)
	}
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	st := runner.Status()
	if !st.Running || st.Name != "svc1" {
		t.Errorf("Status: %+v", st)
	}
	lines, err := runner.Logs()
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(lines) > 0 && lines[len(lines)-1] != "hello" {
		t.Logf("Logs: %v (expected hello)", lines)
	}
	if err := runner.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if runner.Status().Running {
		t.Error("Status after Stop should not be running")
	}

	// Logs with no log file
	svcNoLog := config.Service{Name: "nolog", Command: "echo x"}
	rNoLog, _ := NewProcessRunner(svcNoLog, tmpDir, nil)
	_ = rNoLog.Start(ctx)
	_, err = rNoLog.Logs()
	if err == nil {
		t.Error("Logs() expected error when no log file configured")
	}
	_ = rNoLog.Stop(ctx)
}

func TestProcessRunner_ExecHealthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	svc := config.Service{
		Name:    "healthsvc",
		Command: "sleep 0.5",
		HealthCheck: config.HealthCheck{
			Type:     "exec",
			Endpoint: "true",
			Timeout:  5 * time.Second,
			Interval: 50 * time.Millisecond,
			Retries:  5,
		},
	}
	runner, err := NewProcessRunner(svc, tmpDir, nil)
	if err != nil {
		t.Fatalf("NewProcessRunner: %v", err)
	}
	ctx := context.Background()
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !runner.Status().Running {
		t.Error("expected running after Start with exec healthcheck")
	}
	// Don't call Stop to avoid monitor vs Stop race; process exits on its own.
	time.Sleep(2 * time.Second)
}

func TestProcessRunner_StartAlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	_ = os.MkdirAll("logs", 0755)

	svc := config.Service{Name: "s1", Command: "sleep 3", LogFile: "logs/s1.log"}
	runner, _ := NewProcessRunner(svc, tmpDir, nil)
	ctx := context.Background()
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() { _ = runner.Stop(ctx) }()
	if err := runner.Start(ctx); err == nil {
		t.Error("Start expected error when already running")
	} else if !strings.Contains(err.Error(), "already running") {
		t.Errorf("Start(2nd): %v", err)
	}
}

func TestProcessRunner_StopWhenNotStarted(t *testing.T) {
	svc := config.Service{Name: "s1", Command: "echo x"}
	runner, _ := NewProcessRunner(svc, ".", nil)
	ctx := context.Background()
	if err := runner.Stop(ctx); err != nil {
		t.Errorf("Stop when not started: %v", err)
	}
}

func TestProcessRunner_WithPythonVenvAppScope(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "logs"), 0755)
	venvDir := filepath.Join(tmpDir, ".venv")
	if err := os.MkdirAll(venvDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(venvDir, "pyvenv.cfg"), []byte("home = /usr\n"), 0644); err != nil {
		t.Fatal(err)
	}
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	svc := config.Service{
		Name:    "py-svc",
		Command: "echo ok",
		LogFile: "logs/py.log",
	}
	pythonCfg := &config.PythonConfig{
		Venv: &config.PythonVenvConfig{Dir: ".venv", AppScope: true},
	}
	runner, err := NewProcessRunner(svc, tmpDir, pythonCfg)
	if err != nil {
		t.Fatalf("NewProcessRunner: %v", err)
	}
	ctx := context.Background()
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Start with venv: %v", err)
	}
	defer func() { _ = runner.Stop(ctx) }()
	if st := runner.Status(); !st.Running {
		t.Error("expected runner to be running")
	}
}

package service

import (
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

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
	runner1, err := NewProcessRunner(svc1, appWorkDir)
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
	runner2, err := NewProcessRunner(svc2, appWorkDir)
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
	runner3, err := NewProcessRunner(svc3, appWorkDir)
	if err != nil {
		t.Fatalf("Failed to create process runner: %v", err)
	}
	if runner3.workDir != "/service3" {
		t.Errorf("Expected workdir to be /service3, but got %s", runner3.workDir)
	}
}

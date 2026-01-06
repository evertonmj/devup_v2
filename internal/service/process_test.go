package service

import (
	"path/filepath"
	"testing"

	"devup/internal/config"
)

func TestProcessRunner_Stop(t *testing.T) {
	oldNewCommand := newCommand
	defer func() { newCommand = oldNewCommand }()

	mockCmd := &MockCommander{}
	newCommand = func(ctx context.Context, name string, arg ...string) Commander {
		return mockCmd
	}

	svc := config.Service{
		Name:    "test-service",
		Command: "echo 'hello'",
	}
	runner, err := NewProcessRunner(svc, "/tmp")
	if err != nil {
		t.Fatalf("Failed to create process runner: %v", err)
	}

	// Start the process first
	err = runner.Start(context.Background())
	if err != nil {
		t.Fatalf("Expected Start to succeed, but it failed: %v", err)
	}

	// Now stop it
	err = runner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Expected Stop to succeed, but it failed: %v", err)
	}

	if !mockCmd.killCalled {
		t.Errorf("Expected Kill to be called on the mock command, but it wasn't")
	}

	status := runner.Status()
	if status.Running {
		t.Errorf("Expected service to not be running after Stop, but it is")
	}
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

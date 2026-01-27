package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"devup/internal/config"
)

func TestNewManager(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "service1", Command: "echo test"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"service1"}},
		},
	}

	manager := NewManager(app, "default")
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	if manager.app.Name != "test-app" {
		t.Errorf("NewManager() app name = %v, want test-app", manager.app.Name)
	}
	if manager.mode != "default" {
		t.Errorf("NewManager() mode = %v, want default", manager.mode)
	}
}

func TestCalculateStartOrder(t *testing.T) {
	tests := []struct {
		name     string
		services []config.Service
		toStart  []string
		want     []string
		wantErr  bool
	}{
		{
			name: "no dependencies",
			services: []config.Service{
				{Name: "service1", Command: "echo 1"},
				{Name: "service2", Command: "echo 2"},
			},
			toStart: []string{"service1", "service2"},
			want:    []string{"service1", "service2"},
			wantErr: false,
		},
		{
			name: "simple dependency chain",
			services: []config.Service{
				{Name: "service1", Command: "echo 1", Dependencies: []string{"service2"}},
				{Name: "service2", Command: "echo 2"},
			},
			toStart: []string{"service1", "service2"},
			want:    []string{"service2", "service1"},
			wantErr: false,
		},
		{
			name: "complex dependency graph",
			services: []config.Service{
				{Name: "frontend", Command: "echo 1", Dependencies: []string{"api"}},
				{Name: "api", Command: "echo 2", Dependencies: []string{"database"}},
				{Name: "database", Command: "echo 3"},
			},
			toStart: []string{"frontend", "api", "database"},
			want:    []string{"database", "api", "frontend"},
			wantErr: false,
		},
		{
			name: "circular dependency",
			services: []config.Service{
				{Name: "service1", Command: "echo 1", Dependencies: []string{"service2"}},
				{Name: "service2", Command: "echo 2", Dependencies: []string{"service1"}},
			},
			toStart: []string{"service1", "service2"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing service",
			services: []config.Service{
				{Name: "service1", Command: "echo 1"},
			},
			toStart: []string{"service1", "nonexistent"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &config.AppSpec{
				Name:     "test-app",
				WorkDir:  ".",
				Services: tt.services,
			}
			manager := NewManager(app, "default")

			got, err := manager.calculateStartOrder(tt.toStart)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateStartOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("calculateStartOrder() length = %v, want %v", len(got), len(tt.want))
					return
				}
				for i, svc := range got {
					if svc != tt.want[i] {
						t.Errorf("calculateStartOrder()[%d] = %v, want %v", i, svc, tt.want[i])
					}
				}
			}
		})
	}
}

func TestApplyModeOverrides(t *testing.T) {
	tests := []struct {
		name       string
		service    config.Service
		modeConfig config.Mode
		wantPort   int
		wantCmd    string
		wantEnv    map[string]string
	}{
		{
			name: "no overrides",
			service: config.Service{
				Name:    "service1",
				Command: "echo test",
				Port:    8000,
			},
			modeConfig: config.Mode{
				Services: []string{"service1"},
			},
			wantPort: 8000,
			wantCmd:  "echo test",
		},
		{
			name: "port override",
			service: config.Service{
				Name:    "service1",
				Command: "echo test",
				Port:    8000,
			},
			modeConfig: config.Mode{
				Services: []string{"service1"},
				Overrides: []config.ServiceOverride{
					{
						ServiceName: "service1",
						Port:        9000,
					},
				},
			},
			wantPort: 9000,
			wantCmd:  "echo test",
		},
		{
			name: "command override",
			service: config.Service{
				Name:    "service1",
				Command: "echo test",
				Port:    8000,
			},
			modeConfig: config.Mode{
				Services: []string{"service1"},
				Overrides: []config.ServiceOverride{
					{
						ServiceName: "service1",
						Command:     "echo override",
					},
				},
			},
			wantPort: 8000,
			wantCmd:  "echo override",
		},
		{
			name: "environment override",
			service: config.Service{
				Name:    "service1",
				Command: "echo test",
				Port:    8000,
			},
			modeConfig: config.Mode{
				Services: []string{"service1"},
				Overrides: []config.ServiceOverride{
					{
						ServiceName: "service1",
						Environment: map[string]string{
							"TEST_VAR": "override_value",
						},
					},
				},
			},
			wantPort: 8000,
			wantCmd:  "echo test",
			wantEnv:  map[string]string{"TEST_VAR": "override_value"},
		},
		{
			name: "mode-level environment",
			service: config.Service{
				Name:    "service1",
				Command: "echo test",
				Port:    8000,
			},
			modeConfig: config.Mode{
				Services: []string{"service1"},
				Environment: map[string]string{
					"MODE_VAR": "mode_value",
				},
			},
			wantPort: 8000,
			wantCmd:  "echo test",
			wantEnv:  map[string]string{"MODE_VAR": "mode_value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &config.AppSpec{
				Name:     "test-app",
				WorkDir:  ".",
				Services: []config.Service{tt.service},
			}
			manager := NewManager(app, "test")

			got := manager.applyModeOverrides(tt.service, &tt.modeConfig)
			if got.Port != tt.wantPort {
				t.Errorf("applyModeOverrides() port = %v, want %v", got.Port, tt.wantPort)
			}
			if got.Command != tt.wantCmd {
				t.Errorf("applyModeOverrides() command = %v, want %v", got.Command, tt.wantCmd)
			}
			if tt.wantEnv != nil {
				for k, v := range tt.wantEnv {
					if got.Environment[k] != v {
						t.Errorf("applyModeOverrides() env[%s] = %v, want %v", k, got.Environment[k], v)
					}
				}
			}
		})
	}
}

func TestGetState(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "service1", Command: "echo test"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"service1"}},
		},
	}

	manager := NewManager(app, "default")
	state := manager.GetState()

	if state == nil {
		t.Fatal("GetState() returned nil")
	}
	if state.AppName != "test-app" {
		t.Errorf("GetState() app name = %v, want test-app", state.AppName)
	}
	if state.Mode != "default" {
		t.Errorf("GetState() mode = %v, want default", state.Mode)
	}
	if state.StartTime.IsZero() {
		t.Error("GetState() start time is zero")
	}
}

func TestStatus(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "service1", Command: "echo test"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"service1"}},
		},
	}

	manager := NewManager(app, "default")
	status := manager.Status()

	if status == nil {
		t.Fatal("Status() returned nil")
	}
	// Initially no services should be running
	if len(status) != 0 {
		t.Errorf("Status() returned %d services, want 0", len(status))
	}
}

func TestStartWithInvalidMode(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "service1", Command: "echo test"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"service1"}},
		},
	}

	manager := NewManager(app, "nonexistent")
	ctx := context.Background()

	err := manager.Start(ctx)
	if err == nil {
		t.Error("Start() should fail with nonexistent mode")
	}
}

func TestExecuteHooks(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Hooks: config.Hooks{
			PreStart:  []string{"echo pre-start"},
			PostStart: []string{"echo post-start"},
		},
	}

	manager := NewManager(app, "default")

	// Test that hooks don't cause errors
	err := manager.executeHooks(app.Hooks.PreStart)
	if err != nil {
		t.Errorf("executeHooks() error = %v, want nil", err)
	}

	err = manager.executeHooks(app.Hooks.PostStart)
	if err != nil {
		t.Errorf("executeHooks() error = %v, want nil", err)
	}

	// Test empty hooks
	err = manager.executeHooks([]string{})
	if err != nil {
		t.Errorf("executeHooks() with empty list error = %v, want nil", err)
	}
}

func TestGetServiceLogs(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "service1", Command: "echo test"},
		},
	}

	manager := NewManager(app, "default")

	// Test non-existent service
	_, err := manager.GetServiceLogs("nonexistent")
	if err == nil {
		t.Error("GetServiceLogs() should return error for nonexistent service")
	}
}

func TestManagerStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: tmpDir,
		Services: []config.Service{
			{Name: "s1", Command: "sleep 0.5", Type: "process"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
	}
	manager := NewManager(app, "default")
	ctx := context.Background()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	st := manager.Status()
	if len(st) != 1 || !st["s1"].Running {
		t.Errorf("Status() after Start: got %v", st)
	}
	if err := manager.Stop(ctx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	st = manager.Status()
	if len(st) != 0 {
		t.Errorf("Status() after Stop: got %v, want empty", st)
	}
}

func TestManagerStop_PostStopHookFails(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
		Hooks: config.Hooks{PostStop: []string{"exit 1"}},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Stop(ctx)
	if err == nil {
		t.Fatal("Stop expected error when post-stop hook fails")
	}
	if !strings.Contains(err.Error(), "post-stop") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStop_PreStopHookFails(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
		Hooks: config.Hooks{PreStop: []string{"exit 1"}},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Stop(ctx)
	if err == nil {
		t.Fatal("Stop expected error when pre-stop hook fails")
	}
	if !strings.Contains(err.Error(), "pre-stop") {
		t.Errorf("error = %v", err)
	}
}

func TestExecuteHooksFailure(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Hooks:   config.Hooks{PreStart: []string{"exit 1"}},
	}
	manager := NewManager(app, "default")
	err := manager.executeHooks(app.Hooks.PreStart)
	if err == nil {
		t.Error("executeHooks() expected error for failing hook")
	}
}

func TestManagerStart_PostStartHookFails(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Command: "sleep 2"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
		Hooks: config.Hooks{PostStart: []string{"exit 1"}},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start expected error when post-start hook fails")
	}
	if !strings.Contains(err.Error(), "post-start") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStart_PreStartHookFails(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
		Hooks: config.Hooks{PreStart: []string{"exit 1"}},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start expected error when pre-start hook fails")
	}
	if !strings.Contains(err.Error(), "pre-start") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStart_UnknownServiceType(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Type: "invalid", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start expected error for unknown service type")
	}
	if !strings.Contains(err.Error(), "unknown service type") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStart_DockerRequiresConfig(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Type: "docker", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start expected error when docker service has no docker config")
	}
	if !strings.Contains(err.Error(), "requires docker configuration") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStart_TmuxNotImplemented(t *testing.T) {
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: ".",
		Services: []config.Service{
			{Name: "s1", Type: "tmux", Command: "echo x"},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"s1"}},
		},
	}
	mgr := NewManager(app, "default")
	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start expected error for tmux")
	}
	if !strings.Contains(err.Error(), "tmux") {
		t.Errorf("error = %v", err)
	}
}

func TestManagerStartStop_DockerWhenAvailable(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not in PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	check := exec.CommandContext(ctx, "docker", "info")
	if err := check.Run(); err != nil {
		t.Skip("docker daemon not running")
	}

	tmpDir := t.TempDir()
	containerName := fmt.Sprintf("devup-mgr-test-%d", time.Now().UnixNano()%1000000)
	app := &config.AppSpec{
		Name:    "test-app",
		WorkDir: tmpDir,
		Services: []config.Service{
			{
				Name: "dc",
				Type: "docker",
				Port: 9998,
				Docker: &config.DockerConfig{
					Image:     "alpine:3.18",
					Container: containerName,
					Remove:    true,
					Cmd:       "sleep 2",
				},
			},
		},
		Modes: map[string]config.Mode{
			"default": {Services: []string{"dc"}},
		},
	}
	defer exec.CommandContext(context.Background(), "docker", "rm", "-f", containerName).Run()

	mgr := NewManager(app, "default")
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

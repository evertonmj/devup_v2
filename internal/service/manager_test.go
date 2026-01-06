package service

import (
	"context"
	"testing"

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

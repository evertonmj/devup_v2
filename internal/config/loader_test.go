package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoaderLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		wantApps    int
	}{
		{
			name: "valid basic config",
			yamlContent: `version: "1.0"
apps:
  test-app:
    name: "Test App"
    workdir: "."
    services:
      - name: service1
        command: "echo test"
        port: 8000
    modes:
      default:
        services:
          - service1
`,
			wantErr:  false,
			wantApps: 1,
		},
		{
			name:        "empty config",
			yamlContent: "",
			wantErr:     true,
			wantApps:    0,
		},
		{
			name: "invalid yaml",
			yamlContent: `version: "1.0"
apps:
  test-app:
    name: "Test App"
    invalid: [unclosed
`,
			wantErr:  true,
			wantApps: 0,
		},
		{
			name: "config with no apps",
			yamlContent: `version: "1.0"
apps: {}
`,
			wantErr:  true,
			wantApps: 0,
		},
	}

	// Load with explicit nonexistent path
	t.Run("load explicit nonexistent path", func(t *testing.T) {
		loader := NewLoader("/nonexistent/devup.yaml")
		_, err := loader.Load()
		if err == nil {
			t.Fatal("Load() expected error for nonexistent path")
		}
		if !strings.Contains(err.Error(), "config file not found") && !strings.Contains(err.Error(), "failed to resolve") {
			t.Errorf("Load() error = %v, expected config not found or resolve error", err)
		}
	})

	// Load with path that exists but is not a file (directory → read error)
	t.Run("load path is directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewLoader(tmpDir)
		_, err := loader.Load()
		if err == nil {
			t.Fatal("Load() expected error when path is directory")
		}
		if !strings.Contains(err.Error(), "read") && !strings.Contains(err.Error(), "parse") {
			t.Logf("Load() error = %v (read/parse expected)", err)
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "devup.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Create loader and load config
			loader := NewLoader(tmpFile)
			cfg, err := loader.Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Fatal("Load() returned nil config")
				}
				if len(cfg.Apps) != tt.wantApps {
					t.Errorf("Load() got %d apps, want %d", len(cfg.Apps), tt.wantApps)
				}
			}
		})
	}
}

func TestLoaderValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *AppConfig
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:    "service1",
								Command: "echo test",
								Port:    8000,
							},
						},
						Modes: map[string]Mode{
							"default": {
								Services: []string{"service1"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty apps",
			config: &AppConfig{
				Version: "1.0",
				Apps:    map[string]AppSpec{},
			},
			wantErr: true,
		},
		{
			name: "app with no services",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "service without name",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Command: "echo test",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "service without command",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name: "service1",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate service names",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:    "service1",
								Command: "echo test",
							},
							{
								Name:    "service1",
								Command: "echo test",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mode referencing nonexistent service",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:    "service1",
								Command: "echo test",
							},
						},
						Modes: map[string]Mode{
							"default": {
								Services: []string{"nonexistent"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid healthcheck type",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:    "service1",
								Command: "echo test",
								HealthCheck: HealthCheck{Type: "invalid", Endpoint: "http://x"},
							},
						},
						Modes: map[string]Mode{"default": {Services: []string{"service1"}}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "healthcheck http missing endpoint",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:        "service1",
								Command:     "echo test",
								HealthCheck: HealthCheck{Type: "http"},
							},
						},
						Modes: map[string]Mode{"default": {Services: []string{"service1"}}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unknown dependency",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{Name: "service1", Command: "echo test"},
							{Name: "service2", Command: "echo test", Dependencies: []string{"nonexistent"}},
						},
						Modes: map[string]Mode{"default": {Services: []string{"service1", "service2"}}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "app name empty uses key",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"my-key": {
						Name:    "",
						WorkDir: ".",
						Services: []Service{{Name: "s1", Command: "echo test"}},
						Modes:   map[string]Mode{"default": {Services: []string{"s1"}}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "service with docker config no command",
			config: &AppConfig{
				Version: "1.0",
				Apps: map[string]AppSpec{
					"test-app": {
						Name:    "Test App",
						WorkDir: ".",
						Services: []Service{
							{
								Name:    "db",
								Docker:  &DockerConfig{Image: "postgres:15"},
							},
						},
						Modes: map[string]Mode{"default": {Services: []string{"db"}}},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader("")
			err := loader.validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoaderValidate_WorkdirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{
		Version: "1.0",
		Apps: map[string]AppSpec{
			"test-app": {
				Name:    "Test App",
				WorkDir: tmpDir,
				Services: []Service{
					{Name: "s1", Command: "echo test", WorkDir: "nonexistent_subdir"},
				},
				Modes: map[string]Mode{"default": {Services: []string{"s1"}}},
			},
		},
	}
	loader := NewLoader("")
	err := loader.validate(cfg)
	if err == nil {
		t.Error("validate() expected error for missing workdir")
	}
}

func TestLoaderValidate_AppNameFromKey(t *testing.T) {
	cfg := &AppConfig{
		Version: "1.0",
		Apps: map[string]AppSpec{
			"my-key": {
				Name:    "",
				WorkDir: ".",
				Services: []Service{{Name: "s1", Command: "echo test"}},
				Modes:   map[string]Mode{"default": {Services: []string{"s1"}}},
			},
		},
	}
	loader := NewLoader("")
	err := loader.validate(cfg)
	if err != nil {
		t.Fatalf("validate() unexpected error: %v", err)
	}
	if cfg.Apps["my-key"].Name != "my-key" {
		t.Errorf("Apps[my-key].Name = %q, want my-key", cfg.Apps["my-key"].Name)
	}
}

func TestLoaderGetApp(t *testing.T) {
	config := &AppConfig{
		Version: "1.0",
		Apps: map[string]AppSpec{
			"app1": {Name: "App 1"},
			"app2": {Name: "App 2"},
		},
	}

	loader := NewLoader("")

	tests := []struct {
		name    string
		appName string
		want    string
		wantErr bool
	}{
		{
			name:    "existing app",
			appName: "app1",
			want:    "App 1",
			wantErr: false,
		},
		{
			name:    "nonexistent app",
			appName: "app3",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.GetApp(config, tt.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want {
				t.Errorf("GetApp() = %v, want %v", got.Name, tt.want)
			}
		})
	}
}

func TestListApps(t *testing.T) {
	config := &AppConfig{
		Version: "1.0",
		Apps: map[string]AppSpec{
			"app1": {Name: "App 1"},
			"app2": {Name: "App 2"},
			"app3": {Name: "App 3"},
		},
	}

	apps := ListApps(config)
	if len(apps) != 3 {
		t.Errorf("ListApps() returned %d apps, want 3", len(apps))
	}

	// Check that all apps are present (order doesn't matter)
	appMap := make(map[string]bool)
	for _, app := range apps {
		appMap[app] = true
	}

	for _, expected := range []string{"app1", "app2", "app3"} {
		if !appMap[expected] {
			t.Errorf("ListApps() missing app %s", expected)
		}
	}
}

func TestLoaderResolveConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		explicit  string
		useLocal  bool
		setup     func() (string, func())
		wantErr   bool
		checkPath func(t *testing.T, got string)
	}{
		{
			name:     "explicit path to existing file",
			explicit: "devup.yaml",
			setup: func() (string, func()) {
				tmpDir, _ := os.MkdirTemp("", "devup-test-*")
				tmpFile := filepath.Join(tmpDir, "devup.yaml")
				_ = os.WriteFile(tmpFile, []byte("version: '1.0'\napps: {}"), 0644)
				return tmpFile, func() { _ = os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name:     "explicit path to nonexistent file",
			explicit: "/nonexistent/path/devup.yaml",
			setup: func() (string, func()) {
				return "/nonexistent/path/devup.yaml", func() {}
			},
			wantErr: true,
		},
		{
			name:     "current directory search",
			explicit: "",
			setup: func() (string, func()) {
				tmpDir, _ := os.MkdirTemp("", "devup-test-*")
				origDir, _ := os.Getwd()
				_ = os.Chdir(tmpDir)
				_ = os.WriteFile("devup.yaml", []byte("version: '1.0'\napps: {}"), 0644)
				return "", func() {
					_ = os.Chdir(origDir)
					_ = os.RemoveAll(tmpDir)
				}
			},
			wantErr: false,
		},
		{
			name:     "DEVUP_DEFAULT_PROJECT when file exists",
			explicit: "",
			useLocal: false,
			setup: func() (string, func()) {
				tmpDir, _ := os.MkdirTemp("", "devup-test-*")
				f := filepath.Join(tmpDir, "devup.yaml")
				_ = os.WriteFile(f, []byte("version: '1.0'\napps: {}"), 0644)
				old := os.Getenv("DEVUP_DEFAULT_PROJECT")
				_ = os.Setenv("DEVUP_DEFAULT_PROJECT", tmpDir)
				return "", func() {
					_ = os.Setenv("DEVUP_DEFAULT_PROJECT", old)
					_ = os.RemoveAll(tmpDir)
				}
			},
			wantErr: false,
			checkPath: func(t *testing.T, got string) {
				if !strings.Contains(got, "devup.yaml") {
					t.Errorf("resolveConfigPath() = %s, expected path containing devup.yaml", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explicitPath, cleanup := tt.setup()
			defer cleanup()

			loader := NewLoaderWithLocal(explicitPath, tt.useLocal)
			got, err := loader.resolveConfigPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("resolveConfigPath() returned empty path")
			}
			if tt.checkPath != nil && !tt.wantErr {
				tt.checkPath(t, got)
			}
		})
	}
}

func TestNewLoaderWithLocal(t *testing.T) {
	loader := NewLoaderWithLocal("test.yaml", true)
	if loader == nil {
		t.Fatal("NewLoaderWithLocal() returned nil")
	}
	if loader.configPath != "test.yaml" {
		t.Errorf("NewLoaderWithLocal() configPath = %v, want test.yaml", loader.configPath)
	}
	if !loader.useLocal {
		t.Error("NewLoaderWithLocal() useLocal = false, want true")
	}
}

package config

import (
	"os"
	"path/filepath"
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
		name     string
		explicit string
		setup    func() (string, func())
		wantErr  bool
	}{
		{
			name:     "explicit path to existing file",
			explicit: "devup.yaml",
			setup: func() (string, func()) {
				tmpDir, _ := os.MkdirTemp("", "devup-test-*")
				tmpFile := filepath.Join(tmpDir, "devup.yaml")
				os.WriteFile(tmpFile, []byte("version: '1.0'\napps: {}"), 0644)
				return tmpFile, func() { os.RemoveAll(tmpDir) }
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
				os.Chdir(tmpDir)
				os.WriteFile("devup.yaml", []byte("version: '1.0'\napps: {}"), 0644)
				return "", func() {
					os.Chdir(origDir)
					os.RemoveAll(tmpDir)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explicitPath, cleanup := tt.setup()
			defer cleanup()

			loader := NewLoader(explicitPath)
			got, err := loader.resolveConfigPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("resolveConfigPath() returned empty path")
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

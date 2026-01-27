package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		wantPkgMgr  string
		wantType    string
		wantDevCmd  string
	}{
		{
			name:       "Node.js project",
			files:      []string{"package.json"},
			wantPkgMgr: "npm",
			wantType:   "nodejs",
			wantDevCmd: "npm run dev",
		},
		{
			name:       "Go project",
			files:      []string{"go.mod"},
			wantPkgMgr: "go",
			wantType:   "golang",
			wantDevCmd: "go run .",
		},
		{
			name:       "Python project with requirements",
			files:      []string{"requirements.txt"},
			wantPkgMgr: "pip",
			wantType:   "python",
			wantDevCmd: "python main.py",
		},
		{
			name:       "Rust project",
			files:      []string{"Cargo.toml"},
			wantPkgMgr: "cargo",
			wantType:   "rust",
			wantDevCmd: "cargo run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test files
			for _, file := range tt.files {
				f, err := os.Create(filepath.Join(tmpDir, file))
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				_ = f.Close()
			}

			// Test detection
			info := &ProjectInfo{
				Environment: make(map[string]string),
			}
			detectPackageManager(tmpDir, info)

			if info.PackageMgr != tt.wantPkgMgr {
				t.Errorf("PackageMgr = %v, want %v", info.PackageMgr, tt.wantPkgMgr)
			}
			if info.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", info.Type, tt.wantType)
			}
			if info.DevCmd != tt.wantDevCmd {
				t.Errorf("DevCmd = %v, want %v", info.DevCmd, tt.wantDevCmd)
			}
		})
	}
}

func TestAnalyzeDocumentation(t *testing.T) {
	tests := []struct {
		name            string
		readmeContent   string
		wantDescription string
		wantEnvVars     int
	}{
		{
			name: "README with title and env vars",
			readmeContent: `# My Awesome App

This is a great application.

## Setup

export DATABASE_URL=postgres://localhost/db
export API_KEY=secret
`,
			wantDescription: "My Awesome App",
			wantEnvVars:     2,
		},
		{
			name: "README with ports",
			readmeContent: `# Web App

Runs on port 3000

Backend API on port 8000
`,
			wantDescription: "Web App",
			wantEnvVars:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create README
			readmePath := filepath.Join(tmpDir, "README.md")
			if err := os.WriteFile(readmePath, []byte(tt.readmeContent), 0644); err != nil {
				t.Fatalf("Failed to create README: %v", err)
			}

			// Test analysis
			info := &ProjectInfo{
				Environment: make(map[string]string),
			}
			analyzeDocumentation(tmpDir, info)

			if info.Description != tt.wantDescription {
				t.Errorf("Description = %v, want %v", info.Description, tt.wantDescription)
			}
			if len(info.Environment) != tt.wantEnvVars {
				t.Errorf("Environment vars = %v, want %v", len(info.Environment), tt.wantEnvVars)
			}
		})
	}
}

func TestDetectProjectStructure(t *testing.T) {
	tests := []struct {
		name         string
		directories  []string
		wantServices int
		checkService string
	}{
		{
			name:         "Frontend/Backend structure",
			directories:  []string{"frontend", "backend"},
			wantServices: 2,
			checkService: "frontend",
		},
		{
			name:         "API structure",
			directories:  []string{"api", "web"},
			wantServices: 2,
			checkService: "api",
		},
		{
			name:         "No special structure",
			directories:  []string{"src", "tests"},
			wantServices: 1, // Should create default service
			checkService: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create directories
			for _, dir := range tt.directories {
				if err := os.Mkdir(filepath.Join(tmpDir, dir), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
			}

			// Test detection
			info := &ProjectInfo{
				Environment: make(map[string]string),
			}
			detectProjectStructure(tmpDir, info)

			if len(info.Services) != tt.wantServices {
				t.Errorf("Services count = %v, want %v", len(info.Services), tt.wantServices)
			}

			// Check if specific service exists
			found := false
			for _, svc := range info.Services {
				if svc.Name == tt.checkService {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected service %s not found", tt.checkService)
			}
		})
	}
}

func TestExtractEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantKey string
	}{
		{
			name:    "export statement",
			line:    "export DATABASE_URL=postgres://localhost/db",
			wantKey: "DATABASE_URL",
		},
		{
			name:    "ENV statement",
			line:    "ENV API_KEY=secret123",
			wantKey: "API_KEY",
		},
		{
			name:    "Simple assignment",
			line:    "PORT=3000",
			wantKey: "PORT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ProjectInfo{
				Environment: make(map[string]string),
			}

			extractEnvVars(tt.line, info)

			if _, exists := info.Environment[tt.wantKey]; !exists {
				t.Errorf("Expected environment variable %s not found", tt.wantKey)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name     string
		info     *ProjectInfo
		contains []string
	}{
		{
			name: "Simple single service",
			info: &ProjectInfo{
				Name:        "test-app",
				Description: "Test Application",
				Services: []ServiceInfo{
					{
						Name:    "app",
						Type:    "process",
						Command: "npm start",
						Port:    3000,
					},
				},
			},
			contains: []string{
				"version: \"1.0\"",
				"name: \"test-app\"",
				"description: \"Test Application\"",
				"- name: app",
				"command: \"npm start\"",
				"port: 3000",
			},
		},
		{
			name: "Multi-service with dependencies",
			info: &ProjectInfo{
				Name:        "fullstack",
				Description: "Fullstack App",
				Services: []ServiceInfo{
					{
						Name:    "backend",
						Type:    "api",
						Command: "npm run api",
						Port:    8000,
					},
					{
						Name:         "frontend",
						Type:         "web",
						Command:      "npm run dev",
						Port:         3000,
						Dependencies: []string{"backend"},
					},
				},
			},
			contains: []string{
				"- name: backend",
				"- name: frontend",
				"dependencies:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := generateConfig(tt.info)

			for _, want := range tt.contains {
				if !containsString(config, want) {
					t.Errorf("Generated config missing expected string: %s", want)
				}
			}
		})
	}
}

func TestScanProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a realistic project structure
	readme := `# Test Project

A web application with React frontend.

## Setup

export DATABASE_URL=postgres://localhost/test
export SECRET_KEY=mysecret

## Running

The app runs on port 3000
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	// Run scan
	info, err := scanProject(tmpDir)
	if err != nil {
		t.Fatalf("scanProject failed: %v", err)
	}

	// Verify results
	if info.PackageMgr != "npm" {
		t.Errorf("PackageMgr = %v, want npm", info.PackageMgr)
	}
	if info.Description != "Test Project" {
		t.Errorf("Description = %v, want 'Test Project'", info.Description)
	}
	if len(info.Environment) == 0 {
		t.Error("Expected environment variables to be detected")
	}
	if _, exists := info.Environment["DATABASE_URL"]; !exists {
		t.Error("Expected DATABASE_URL to be detected")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

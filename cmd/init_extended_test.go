
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDetectDatabaseServices(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		existingServices  []ServiceInfo
		expectedServices  []string
		unexpectedServices []string
	}{
		{
			name: "Detect postgres from go.mod",
			files: map[string]string{
				"go.mod": "require gorm.io/driver/postgres v1.3.9",
			},
			expectedServices: []string{"postgres"},
		},
		{
			name: "Detect mysql from package.json",
			files: map[string]string{
				"package.json": `{ "dependencies": { "mysql": "latest" } }`,
			},
			expectedServices: []string{"mysql"},
		},
		{
			name: "Detect mongo from requirements.txt",
			files: map[string]string{
				"requirements.txt": "pymongo==4.1.1",
			},
			expectedServices: []string{"mongo"},
		},
		{
			name: "Detect postgres from docker-compose.yml",
			files: map[string]string{
				"docker-compose.yml": "services:\n  postgres:\n",
			},
			expectedServices: []string{"postgres"},
		},
		{
			name: "Do not add duplicate database",
			files: map[string]string{
				"go.mod": "require gorm.io/driver/postgres v1.3.9",
			},
			existingServices: []ServiceInfo{
				{Name: "postgres"},
			},
			unexpectedServices: []string{"postgres"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for filename, content := range tt.files {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", filename, err)
				}
			}

			info := &ProjectInfo{
				Services:    tt.existingServices,
				Environment: make(map[string]string),
			}

			detectDatabaseServices(tmpDir, info)

			for _, expectedService := range tt.expectedServices {
				found := false
				for _, s := range info.Services {
					if s.Name == expectedService {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find service %s, but did not", expectedService)
				}
			}

			for _, unexpectedService := range tt.unexpectedServices {
                count := 0
				for _, s := range info.Services {
					if s.Name == unexpectedService {
						count++
					}
				}
				if count > 1 {
					t.Errorf("Did not expect to find duplicate service %s", unexpectedService)
				}
			}
		})
	}
}

func TestDetectServices(t *testing.T) {
	dockerComposeContent := `
services:
  web:
    build: .
  api:
    build: ./api
`
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(dockerComposeContent), 0644); err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	info := &ProjectInfo{}
	detectServices(tmpDir, info)

	if len(info.Services) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(info.Services))
	}

	expectedServices := []string{"web", "api"}
	for _, expected := range expectedServices {
		found := false
		for _, s := range info.Services {
			if s.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find service %s, but didn't", expected)
		}
	}
}

func TestScanEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()
	envContent := `
# comment
DATABASE_URL=postgres://user:pass@host:port/db
SECRET_KEY=my-secret
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env.example: %v", err)
	}

	info := &ProjectInfo{
		Environment: make(map[string]string),
	}
	scanEnvFiles(tmpDir, info)

	if val, ok := info.Environment["DATABASE_URL"]; !ok || val != "postgres://user:pass@host:port/db" {
		t.Errorf("Expected DATABASE_URL to be set correctly, got %s", val)
	}
	if val, ok := info.Environment["SECRET_KEY"]; !ok || val != "my-secret" {
		t.Errorf("Expected SECRET_KEY to be set correctly, got %s", val)
	}
}

func TestParseDotEnvContent(t *testing.T) {
	content := `
KEY1="value1"
KEY2='value2'
KEY3=value3
export KEY4=value4
`
	info := &ProjectInfo{
		Environment: make(map[string]string),
	}
	parseDotEnvContent(content, info)

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
		"KEY4": "value4",
	}

	for key, expectedVal := range expected {
		if val, ok := info.Environment[key]; !ok || val != expectedVal {
			t.Errorf("For key %s, expected '%s', got '%s'", key, expectedVal, val)
		}
	}
}

func TestExtractPorts(t *testing.T) {
	info := &ProjectInfo{Services: []ServiceInfo{{Name: "web"}}}
	extractPorts("The application runs on port 3000.", info)

	if info.Services[0].Port != 3000 {
		t.Errorf("Expected port 3000 to be extracted, got %d", info.Services[0].Port)
	}
}

func TestExtractCommands(t *testing.T) {
	info := &ProjectInfo{}
	extractCommands("You can start the server with `npm run dev`", info)

	if info.DevCmd != "npm run dev" {
		t.Errorf("Expected DevCmd to be 'npm run dev', got '%s'", info.DevCmd)
	}
}

func TestHasPythonServices(t *testing.T) {
	tests := []struct {
		name     string
		info     *ProjectInfo
		expected bool
	}{
		{
			name:     "Has pip package manager",
			info:     &ProjectInfo{PackageMgr: "pip"},
			expected: true,
		},
		{
			name: "Has python language service",
			info: &ProjectInfo{Services: []ServiceInfo{{Language: "python"}}},
			expected: true,
		},
		{
			name:     "No python services",
			info:     &ProjectInfo{PackageMgr: "npm"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasPythonServices(tt.info); got != tt.expected {
				t.Errorf("hasPythonServices() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPythonServiceDirs(t *testing.T) {
	tests := []struct {
		name     string
		info     *ProjectInfo
		expected []string
	}{
		{
			name:     "pip project in root",
			info:     &ProjectInfo{PackageMgr: "pip"},
			expected: []string{"."},
		},
		{
			name: "python service in subdirectory",
			info: &ProjectInfo{Services: []ServiceInfo{{Language: "python", Directory: "backend"}}},
			expected: []string{"backend"},
		},
		{
			name: "Multiple python services",
			info: &ProjectInfo{
				PackageMgr: "pip",
				Services: []ServiceInfo{
					{Language: "python", Directory: "backend"},
					{Language: "python", Directory: "worker"},
				},
			},
			expected: []string{".", "backend", "worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs := pythonServiceDirs(tt.info)
			if len(dirs) != len(tt.expected) {
				t.Fatalf("Expected %d directories, got %d", len(tt.expected), len(dirs))
			}
			for i, dir := range dirs {
				if dir != tt.expected[i] {
					t.Errorf("Expected dir %s, got %s", tt.expected[i], dir)
				}
			}
		})
	}
}

func TestRunInit(t *testing.T) {
    t.Run("devup.yaml already exists", func(t *testing.T) {
        tmpDir := t.TempDir()
        originalWd, _ := os.Getwd()
        os.Chdir(tmpDir)
        defer os.Chdir(originalWd)

        if _, err := os.Create("devup.yaml"); err != nil {
            t.Fatalf("Failed to create dummy devup.yaml: %v", err)
        }

        onlyInit = true
        defer func() { onlyInit = false }()

        cmd := &cobra.Command{}
        err := runInit(cmd, []string{})

        if err == nil || !strings.Contains(err.Error(), "devup.yaml already exists") {
            t.Errorf("Expected error about existing devup.yaml, got: %v", err)
        }
    })

    t.Run("force overwrite", func(t *testing.T) {
        tmpDir := t.TempDir()
        originalWd, _ := os.Getwd()
        os.Chdir(tmpDir)
        defer os.Chdir(originalWd)

        if _, err := os.Create("devup.yaml"); err != nil {
            t.Fatalf("Failed to create dummy devup.yaml: %v", err)
        }

        initForce = true
        defer func() { initForce = false }()

        // Mock the install and setup commands to prevent them from running
        originalInstallCmd := installCmd
        installCmd = &cobra.Command{Use: "install", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
        defer func() { installCmd = originalInstallCmd }()
        
        rootCmd.AddCommand(installCmd)
        
        originalSetupCmd := setupCmd
        setupCmd = &cobra.Command{Use: "setup", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
        defer func() { setupCmd = originalSetupCmd }()
        rootCmd.AddCommand(setupCmd)

        cmd := &cobra.Command{}
        err := runInit(cmd, []string{})

        if err != nil {
            t.Errorf("Expected no error with --force, got: %v", err)
        }
    })

    t.Run("only-init flag", func(t *testing.T) {
        tmpDir := t.TempDir()
        originalWd, _ := os.Getwd()
        os.Chdir(tmpDir)
        defer os.Chdir(originalWd)
    
        onlyInit = true
        defer func() { onlyInit = false }()
    
        // No need to mock install/setup as they shouldn't be called
    
        cmd := &cobra.Command{}
        err := runInit(cmd, []string{})
    
        if err != nil {
            t.Errorf("runInit with --only-init failed: %v", err)
        }
    
        if _, err := os.Stat("devup.yaml"); os.IsNotExist(err) {
            t.Error("devup.yaml was not created with --only-init")
        }
    })
}

// Mocks for cobra commands for testing purposes
var (
	mockInstallCmd = &cobra.Command{
		Use: "install",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("mock install called")
			return nil
		},
	}
	mockSetupCmd = &cobra.Command{
		Use: "setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("mock setup called")
			return nil
		},
	}
)

func TestMain(m *testing.M) {
	// Setup mock commands
	rootCmd.AddCommand(mockInstallCmd)
	rootCmd.AddCommand(mockSetupCmd)
	installCmd = mockInstallCmd
	setupCmd = mockSetupCmd

	// Run tests
	os.Exit(m.Run())
}

package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListApplications(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	const yaml = `version: "1.0"
apps:
  myapp:
    name: "My App"
    description: "Test"
    workdir: "."
    services:
      - name: s1
        command: "echo test"
    modes:
      default:
        services: [s1]
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	saveCfg, saveLocal, saveApp := cfgFile, local, appName
	defer func() {
		cfgFile, local, appName = saveCfg, saveLocal, saveApp
	}()

	cfgFile = cfgPath
	local = true
	appName = ""

	if err := listApplications(); err != nil {
		t.Fatalf("listApplications: %v", err)
	}
}

func TestListApplications_NoApps(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	const yaml = `version: "1.0"
apps: {}
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	origWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)
	os.WriteFile("devup.yaml", []byte(yaml), 0644)

	saveCfg, saveLocal := cfgFile, local
	defer func() { cfgFile, local = saveCfg, saveLocal }()

	cfgFile = ""
	local = true

	if err := listApplications(); err == nil {
		t.Error("listApplications expected error for config with no apps")
	}
}

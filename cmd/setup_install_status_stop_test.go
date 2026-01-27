package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

const minimalYaml = `version: "1.0"
apps:
  myapp:
    name: "My App"
    workdir: "."
    services:
      - name: s1
        command: "echo test"
    modes:
      default:
        services: [s1]
`

const setupWithDirsYaml = `version: "1.0"
apps:
  myapp:
    name: "My App"
    workdir: "."
    services:
      - name: s1
        command: "echo test"
    modes:
      default:
        services: [s1]
    setup:
      directories:
        - "logs"
        - "tmp"
`

const installWithHooksYaml = `version: "1.0"
apps:
  myapp:
    name: "My App"
    workdir: "."
    services:
      - name: s1
        command: "echo test"
    modes:
      default:
        services: [s1]
    hooks:
      pre_install:
        - "echo pre-install"
    install:
      steps: []
`

func setupMinimalConfig(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	if err := os.WriteFile(cfgPath, []byte(minimalYaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmpDir)
	return cfgPath, func() {
		os.Chdir(orig)
	}
}

func TestRunSetup_DryRun(t *testing.T) {
	cfgPath, cleanup := setupMinimalConfig(t)
	defer cleanup()

	saveCfg, saveLocal, saveApp, saveDry := cfgFile, local, appName, setupDryRun
	defer func() {
		cfgFile, local, appName, setupDryRun = saveCfg, saveLocal, saveApp, saveDry
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	setupDryRun = true

	if err := runSetup(); err != nil {
		t.Fatalf("runSetup: %v", err)
	}
}

func TestRunInstall_DryRun(t *testing.T) {
	cfgPath, cleanup := setupMinimalConfig(t)
	defer cleanup()

	saveCfg, saveLocal, saveApp, saveDry := cfgFile, local, appName, dryRun
	defer func() {
		cfgFile, local, appName, dryRun = saveCfg, saveLocal, saveApp, saveDry
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	dryRun = true

	if err := runInstall(); err != nil {
		t.Fatalf("runInstall: %v", err)
	}
}

func TestShowStatus(t *testing.T) {
	cfgPath, cleanup := setupMinimalConfig(t)
	defer cleanup()

	saveCfg, saveLocal, saveApp := cfgFile, local, appName
	defer func() { cfgFile, local, appName = saveCfg, saveLocal, saveApp }()

	cfgFile = cfgPath
	local = true
	appName = "myapp"

	if err := showStatus(); err != nil {
		t.Fatalf("showStatus: %v", err)
	}
}

func TestStopApplication(t *testing.T) {
	cfgPath, cleanup := setupMinimalConfig(t)
	defer cleanup()

	saveCfg, saveLocal, saveApp := cfgFile, local, appName
	defer func() { cfgFile, local, appName = saveCfg, saveLocal, saveApp }()

	cfgFile = cfgPath
	local = true
	appName = "myapp"

	if err := stopApplication(); err != nil {
		t.Fatalf("stopApplication: %v", err)
	}
}

func TestRunSetup_WithDirectoriesDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	if err := os.WriteFile(cfgPath, []byte(setupWithDirsYaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(orig)

	saveCfg, saveLocal, saveApp, saveDry := cfgFile, local, appName, setupDryRun
	defer func() {
		cfgFile, local, appName, setupDryRun = saveCfg, saveLocal, saveApp, saveDry
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	setupDryRun = true

	if err := runSetup(); err != nil {
		t.Fatalf("runSetup: %v", err)
	}
}

func TestRunInstall_WithHooksDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	if err := os.WriteFile(cfgPath, []byte(installWithHooksYaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	orig, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(orig)

	saveCfg, saveLocal, saveApp, saveDry := cfgFile, local, appName, dryRun
	defer func() {
		cfgFile, local, appName, dryRun = saveCfg, saveLocal, saveApp, saveDry
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	dryRun = true

	if err := runInstall(); err != nil {
		t.Fatalf("runInstall: %v", err)
	}
}

func TestStartApplication_InvalidMode(t *testing.T) {
	cfgPath, cleanup := setupMinimalConfig(t)
	defer cleanup()

	saveCfg, saveLocal, saveApp, saveMode := cfgFile, local, appName, mode
	defer func() {
		cfgFile, local, appName, mode = saveCfg, saveLocal, saveApp, saveMode
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	mode = "nonexistent"

	if err := startApplication(); err == nil {
		t.Error("startApplication expected error for invalid mode")
	}
}

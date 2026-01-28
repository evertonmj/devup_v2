package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

const envTestYaml = "version: \"1.0\"\napps:\n  myapp:\n    name: \"My App\"\n    workdir: \".\"\n    services:\n      - name: s1\n        command: \"echo test\"\n    modes:\n      default:\n        services: [s1]\n"

func TestRunEnv(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	cfgContent := "version: \"1.0\"\napps:\n  myapp:\n    name: \"My App\"\n    workdir: \"" + tmpDir + "\"\n    services:\n      - name: s1\n        command: \"echo test\"\n    modes:\n      default:\n        services: [s1]\n"
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	content := "KEY1=value1\nKEY2=value2\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	saveCfg, saveLocal, saveApp, saveFmt := cfgFile, local, appName, envFormat
	defer func() {
		cfgFile, local, appName, envFormat = saveCfg, saveLocal, saveApp, saveFmt
	}()

	cfgFile = cfgPath
	local = true
	appName = "myapp"
	envFormat = "export"

	if err := runEnv(); err != nil {
		t.Fatalf("runEnv: %v", err)
	}
}

func TestRunEnv_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	if err := os.WriteFile(cfgPath, []byte(envTestYaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.WriteFile("devup.yaml", []byte(envTestYaml), 0644); err != nil {
		t.Fatalf("write devup.yaml: %v", err)
	}
	// no .env

	saveCfg, saveLocal, saveApp := cfgFile, local, appName
	defer func() { cfgFile, local, appName = saveCfg, saveLocal, saveApp }()

	cfgFile = ""
	local = true
	appName = "myapp"
	envFormat = "export"

	if err := runEnv(); err == nil {
		t.Error("runEnv expected error when .env not found")
	}
}

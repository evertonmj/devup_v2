package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEnv(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "devup.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	const yaml = `version: "1.0"
apps:
  myapp:
    name: "My App"
    workdir: "%s"
    services:
      - name: s1
        command: "echo test"
    modes:
      default:
        services: [s1]
`
	yamlWithWorkdir := tmpDir
	if len(tmpDir) > 0 && tmpDir[0] != '/' {
		// use absolute path
		abs, _ := filepath.Abs(tmpDir)
		yamlWithWorkdir = abs
	}
	content := `KEY1=value1
KEY2=value2
`
	if err := os.WriteFile(cfgPath, []byte("version: \"1.0\"\napps:\n  myapp:\n    name: \"My App\"\n    workdir: \""+tmpDir+"\"\n    services:\n      - name: s1\n        command: \"echo test\"\n    modes:\n      default:\n        services: [s1]\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_ = yamlWithWorkdir
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
	const yaml = `version: "1.0"
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
	if err := os.WriteFile(cfgPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)
	os.WriteFile("devup.yaml", []byte(yaml), 0644)
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

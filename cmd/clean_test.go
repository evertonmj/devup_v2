package cmd

import (
	"os"
	"testing"
)

func TestRunClean_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
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
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()
	if err := os.WriteFile("devup.yaml", []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	saveCfg, saveLocal, saveApp := cfgFile, local, appName
	saveDry, saveForce := cleanDryRun, cleanForce
	defer func() {
		cfgFile, local, appName = saveCfg, saveLocal, saveApp
		cleanDryRun, cleanForce = saveDry, saveForce
	}()

	cfgFile = ""
	local = true
	appName = "myapp"
	cleanDryRun = true
	cleanForce = true

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("runClean: %v", err)
	}
}

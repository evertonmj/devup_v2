package cmd

import (
	"testing"

	"devup/internal/config"
)

func TestGetApp(t *testing.T) {
	save := appName
	defer func() { appName = save }()

	singleCfg := &config.AppConfig{
		Version: "1.0",
		Apps: map[string]config.AppSpec{
			"single": {
				Name:    "Single App",
				WorkDir: ".",
				Services: []config.Service{{Name: "s1", Command: "echo"}},
				Modes:   map[string]config.Mode{"default": {Services: []string{"s1"}}},
			},
		},
	}
	multiCfg := &config.AppConfig{
		Version: "1.0",
		Apps: map[string]config.AppSpec{
			"a": {
				Name:    "A",
				WorkDir: ".",
				Services: []config.Service{{Name: "s1", Command: "echo"}},
				Modes:   map[string]config.Mode{"default": {Services: []string{"s1"}}},
			},
			"b": {
				Name:    "B",
				WorkDir: ".",
				Services: []config.Service{{Name: "s1", Command: "echo"}},
				Modes:   map[string]config.Mode{"default": {Services: []string{"s1"}}},
			},
		},
	}

	t.Run("single app no flag", func(t *testing.T) {
		appName = ""
		got, err := getApp(singleCfg)
		if err != nil {
			t.Fatalf("getApp: %v", err)
		}
		if got.Name != "Single App" {
			t.Errorf("getApp single = %s", got.Name)
		}
	})

	t.Run("multi app no flag", func(t *testing.T) {
		appName = ""
		_, err := getApp(multiCfg)
		if err == nil {
			t.Error("getApp multi expected error")
		}
	})

	t.Run("app by name", func(t *testing.T) {
		appName = "single"
		got, err := getApp(singleCfg)
		if err != nil {
			t.Fatalf("getApp: %v", err)
		}
		if got.Name != "Single App" {
			t.Errorf("getApp = %s", got.Name)
		}
	})

	t.Run("app not found", func(t *testing.T) {
		appName = "nonexistent"
		_, err := getApp(singleCfg)
		if err == nil {
			t.Error("getApp expected error for nonexistent")
		}
	})
}

func TestGetApp_SingleApp(t *testing.T) {
	cfg := &config.AppConfig{
		Version: "1.0",
		Apps: map[string]config.AppSpec{
			"only": {
				Name:    "Only",
				WorkDir: ".",
				Services: []config.Service{{Name: "s1", Command: "echo"}},
				Modes:   map[string]config.Mode{"default": {Services: []string{"s1"}}},
			},
		},
	}
	appName = ""
	got, err := getApp(cfg)
	if err != nil {
		t.Fatalf("getApp: %v", err)
	}
	if got.Name != "Only" {
		t.Errorf("getApp = %s", got.Name)
	}
}

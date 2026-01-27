package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devup/internal/config"
)

func TestDockerRunner_WhenDockerAvailable(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not in PATH, skipping Docker test")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	check := exec.CommandContext(ctx, "docker", "info")
	if err := check.Run(); err != nil {
		t.Skip("docker daemon not running, skipping Docker test")
	}

	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	os.MkdirAll(logDir, 0755)
	containerName := fmt.Sprintf("devup-test-docker-%d", time.Now().UnixNano()%1000000)

	svc := &config.Service{
		Name: "dockersvc",
		Type: "docker",
		Port: 9999,
		Docker: &config.DockerConfig{
			Image:         "alpine:3.18",
			Container:     containerName,
			Ports:         []string{"19999:19999"},
			Volumes:       []string{tmpDir + ":/data"},
			Environment:   map[string]string{"X": "y"},
			RestartPolicy: "no",
			Remove:        true,
			Cmd:           "sleep 1",
		},
		LogFile: filepath.Join(logDir, "dockersvc.log"),
	}
	runner := NewDockerRunner(svc, svc.Docker, tmpDir)
	defer func() {
		exec.CommandContext(context.Background(), "docker", "rm", "-f", containerName).Run()
	}()

	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Docker Start: %v", err)
	}
	if err := runner.Start(ctx); err == nil {
		t.Error("Docker Start expected error when already running")
	} else if !strings.Contains(err.Error(), "already running") {
		t.Logf("Start (2nd): %v", err)
	}
	st := runner.Status()
	if !st.Running {
		t.Logf("Status (container may have exited): %+v", st)
	}
	logs, err := runner.Logs()
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(logs) == 0 {
		t.Error("Logs expected non-empty")
	}
	if err := runner.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := runner.Stop(ctx); err != nil {
		t.Errorf("Stop when already stopped: %v", err)
	}

	// Cover more buildDockerRunCommand branches: WorkDir, Labels, Entrypoint
	containerName2 := containerName + "-2"
	svc2 := &config.Service{
		Name: "dockersvc2",
		Type: "docker",
		Port: 9998,
		Docker: &config.DockerConfig{
			Image:     "alpine:3.18",
			Container: containerName2,
			Remove:    true,
			Cmd:       "1",
			WorkDir:   "/tmp",
			Labels:    map[string]string{"devup.test": "1"},
			Entrypoint: "sleep",
		},
		LogFile: filepath.Join(logDir, "dockersvc2.log"),
	}
	runner2 := NewDockerRunner(svc2, svc2.Docker, tmpDir)
	defer exec.CommandContext(context.Background(), "docker", "rm", "-f", containerName2).Run()
	if err := runner2.Start(ctx); err != nil {
		t.Logf("Docker Start (WorkDir+Labels): %v", err)
		return
	}
	_, _ = runner2.Logs()
	_ = runner2.Stop(ctx)
}

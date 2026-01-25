package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"devup/internal/config"
)

// DockerRunner implements ServiceRunner for Docker containers
type DockerRunner struct {
	name      string
	docker    *config.DockerConfig
	service   *config.Service
	container string
	logFile   string
	startTime time.Time
	isStopped bool
}

// NewDockerRunner creates a new Docker service runner
func NewDockerRunner(svc *config.Service, docker *config.DockerConfig, cwd string) *DockerRunner {
	container := docker.Container
	if container == "" {
		container = svc.Name
	}

	logFile := svc.LogFile
	if logFile == "" {
		logFile = filepath.Join(cwd, fmt.Sprintf("logs/%s.log", svc.Name))
	} else if !filepath.IsAbs(logFile) {
		logFile = filepath.Join(cwd, logFile)
	}

	// Ensure logs directory exists
	logDir := filepath.Dir(logFile)
	os.MkdirAll(logDir, 0755)

	return &DockerRunner{
		name:      svc.Name,
		docker:    docker,
		service:   svc,
		container: container,
		logFile:   logFile,
		isStopped: true,
	}
}

// Start starts the Docker container
func (dr *DockerRunner) Start(ctx context.Context) error {
	if !dr.isStopped {
		return fmt.Errorf("container %s is already running", dr.container)
	}

	// Check docker availability
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed or not in PATH. Please install Docker Desktop and ensure it's running")
	}
	// Check docker daemon status with a short timeout
	{
		checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		verCmd := exec.CommandContext(checkCtx, "docker", "info")
		if err := verCmd.Run(); err != nil {
			return fmt.Errorf("docker daemon is not running. Start Docker Desktop before running docker services")
		}
	}

	// Build docker run command
	args := dr.buildDockerRunCommand()

	// Execute docker run
	cmd := exec.CommandContext(ctx, "docker", args...)

	// Attach to current process group for clean shutdown
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Open log file
	logFilePtr, err := os.OpenFile(dr.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", dr.logFile, err)
	}
	defer logFilePtr.Close()

	cmd.Stdout = logFilePtr
	cmd.Stderr = logFilePtr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker container %s: %w", dr.container, err)
	}

	dr.startTime = time.Now()
	dr.isStopped = false

	// Log container start
	fmt.Fprintf(logFilePtr, "[devup] Container started at %s\n", dr.startTime.Format(time.RFC3339))

	return nil
}

// Stop stops and optionally removes the Docker container
func (dr *DockerRunner) Stop(ctx context.Context) error {
	if dr.isStopped {
		return nil
	}

	// Stop the container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", "-t", "10", dr.container)
	if err := stopCmd.Run(); err != nil {
		// Container might not exist, which is okay
		fmt.Printf("warning: failed to stop container %s: %v\n", dr.container, err)
	}

	// Remove container if configured
	if dr.docker.Remove {
		removeCmd := exec.CommandContext(ctx, "docker", "rm", "-f", dr.container)
		if err := removeCmd.Run(); err != nil {
			fmt.Printf("warning: failed to remove container %s: %v\n", dr.container, err)
		}
	}

	dr.isStopped = true

	// Log container stop
	if f, err := os.OpenFile(dr.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		defer f.Close()
		fmt.Fprintf(f, "[devup] Container stopped at %s (duration: %v)\n",
			time.Now().Format(time.RFC3339), time.Since(dr.startTime))
	}

	return nil
}

// Status returns the status of the Docker container
func (dr *DockerRunner) Status() ServiceStatus {
	// Check if container is running
	inspectCmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", dr.container)
	output, err := inspectCmd.CombinedOutput()

	running := false
	if err == nil && strings.TrimSpace(string(output)) == "true" {
		running = true
	}

	return ServiceStatus{
		Name:      dr.name,
		Running:   running,
		PID:       0, // Docker containers don't have direct PID mappings
		Port:      dr.service.Port,
		StartTime: dr.startTime,
		Error:     nil,
	}
}

// Logs returns the logs for the Docker container
func (dr *DockerRunner) Logs() ([]string, error) {
	// For Docker, we can optionally get container logs or return the log file
	return []string{dr.logFile}, nil
}

// buildDockerRunCommand builds the docker run command arguments
func (dr *DockerRunner) buildDockerRunCommand() []string {
	args := []string{"run", "-d"}

	// Container name
	args = append(args, "--name", dr.container)

	// Restart policy
	if dr.docker.RestartPolicy != "" {
		args = append(args, "--restart", dr.docker.RestartPolicy)
	}

	// Port mappings
	for _, port := range dr.docker.Ports {
		args = append(args, "-p", port)
	}

	// Volume mounts
	for _, volume := range dr.docker.Volumes {
		args = append(args, "-v", volume)
	}

	// Environment variables
	for key, value := range dr.docker.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Networks
	for _, network := range dr.docker.Networks {
		args = append(args, "--network", network)
	}

	// Working directory
	if dr.docker.WorkDir != "" {
		args = append(args, "-w", dr.docker.WorkDir)
	}

	// User
	if dr.docker.User != "" {
		args = append(args, "-u", dr.docker.User)
	}

	// Privileged mode
	if dr.docker.Privileged {
		args = append(args, "--privileged")
	}

	// Labels
	for key, value := range dr.docker.Labels {
		args = append(args, "-l", fmt.Sprintf("%s=%s", key, value))
	}

	// Entrypoint override
	if dr.docker.Entrypoint != "" {
		args = append(args, "--entrypoint", dr.docker.Entrypoint)
	}

	// Pull image before running
	if dr.docker.Pull {
		pullCmd := exec.Command("docker", "pull", dr.docker.Image)
		pullCmd.Run() // Ignore errors, docker run will fail if image unavailable
	}

	// Image
	args = append(args, dr.docker.Image)

	// Command/args override
	if dr.docker.Cmd != "" {
		// Split command into parts if it contains spaces
		cmdParts := strings.Fields(dr.docker.Cmd)
		args = append(args, cmdParts...)
	}

	return args
}

// Compile-time assertion that DockerRunner implements ServiceRunner
var _ ServiceRunner = (*DockerRunner)(nil)

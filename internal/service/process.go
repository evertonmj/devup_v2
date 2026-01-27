package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"devup/internal/config"
)

// Commander is an interface that abstracts the os/exec.Cmd behavior.
type Commander interface {
	Start() error
	Wait() error
	Process() *os.Process
	SysProcAttr() *syscall.SysProcAttr
	SetSysProcAttr(attr *syscall.SysProcAttr)
	SetDir(dir string)
	SetEnv(env []string)
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
	Kill() error
}

// CmdWrapper is a concrete implementation of Commander that wraps os/exec.Cmd.
type CmdWrapper struct {
	*exec.Cmd
}

func (c *CmdWrapper) Process() *os.Process {
	return c.Cmd.Process
}

func (c *CmdWrapper) SysProcAttr() *syscall.SysProcAttr {
	return c.Cmd.SysProcAttr
}

func (c *CmdWrapper) SetSysProcAttr(attr *syscall.SysProcAttr) {
	c.Cmd.SysProcAttr = attr
}

func (c *CmdWrapper) SetDir(dir string) {
	c.Dir = dir
}

func (c *CmdWrapper) SetEnv(env []string) {
	c.Env = env
}

func (c *CmdWrapper) SetStdout(stdout io.Writer) {
	c.Stdout = stdout
}

func (c *CmdWrapper) SetStderr(stderr io.Writer) {
	c.Stderr = stderr
}

func (c *CmdWrapper) Kill() error {
	return c.Cmd.Process.Kill()
}

// newCommand creates a new CmdWrapper.
// This function acts as a factory for our Commander interface.
var newCommand = func(ctx context.Context, name string, arg ...string) Commander {
	return &CmdWrapper{
		Cmd: exec.CommandContext(ctx, name, arg...),
	}
}

// ProcessRunner manages a single process-based service
type ProcessRunner struct {
	config      config.Service
	workDir     string
	cmd         Commander
	logFile     *os.File
	startTime   time.Time
	mu          sync.Mutex
	monitorDone chan struct{} // closed when monitor goroutine finishes
}

// NewProcessRunner creates a new process runner for a service
func NewProcessRunner(svc config.Service, appWorkDir string) (*ProcessRunner, error) {
	// Determine working directory
	workDir := appWorkDir
	if svc.WorkDir != "" {
		if filepath.IsAbs(svc.WorkDir) {
			workDir = svc.WorkDir
		} else {
			workDir = filepath.Join(appWorkDir, svc.WorkDir)
		}
	}

	return &ProcessRunner{
		config:  svc,
		workDir: workDir,
	}, nil
}

// Start starts the service process
func (p *ProcessRunner) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process() != nil {
		return fmt.Errorf("service '%s' is already running", p.config.Name)
	}

	// Create command
	cmd := newCommand(ctx, "bash", "-c", p.config.Command)
	cmd.SetDir(p.workDir)

	// Set environment variables
	cmd.SetEnv(os.Environ())
	for k, v := range p.config.Environment {
		cmd.SetEnv(append(os.Environ(), fmt.Sprintf("%s=%s", k, v)))
	}

	// Create a new process group so we can kill all child processes
	cmd.SetSysProcAttr(&syscall.SysProcAttr{
		Setpgid: true,
	})

	// Setup logging
	if err := p.setupLogging(cmd); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service '%s': %w", p.config.Name, err)
	}

	p.cmd = cmd
	p.startTime = time.Now()
	p.monitorDone = make(chan struct{})

	// Monitor process in background
	go p.monitor()

	// Wait for service to be ready
	if p.config.HealthCheck.Type != "" {
		if err := p.waitForReady(ctx); err != nil {
			_ = p.Stop(ctx)
			return err
		}
	}

	return nil
}

// Stop stops the service process
func (p *ProcessRunner) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process() == nil {
		return nil
	}

	// Get the process group ID (PGID)
	pgid, err := syscall.Getpgid(p.cmd.Process().Pid)
	if err != nil {
		// Process may already be dead, try killing it anyway
		_ = p.cmd.Kill()
		return nil
	}

	// Kill the entire process group (negative PID kills the group)
	// This ensures all child processes (npm, node, etc.) are killed
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		// If SIGTERM fails, force kill
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}

	// Wait for monitor (which owns Wait) to finish, with timeout
	select {
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-p.monitorDone
	case <-p.monitorDone:
	}

	// Close log file
	if p.logFile != nil {
		_ = p.logFile.Close()
		p.logFile = nil
	}

	p.cmd = nil
	return nil
}

// Status returns the current status of the service
func (p *ProcessRunner) Status() ServiceStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := ServiceStatus{
		Name:      p.config.Name,
		Port:      p.config.Port,
		StartTime: p.startTime,
	}

	if p.cmd != nil && p.cmd.Process() != nil {
		status.Running = true
		status.PID = p.cmd.Process().Pid
	}

	return status
}

// Logs returns recent log entries for the service
func (p *ProcessRunner) Logs() ([]string, error) {
	if p.config.LogFile == "" {
		return nil, fmt.Errorf("no log file configured for service '%s'", p.config.Name)
	}

	// Get the current working directory (where devup was run)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Resolve log file path relative to where devup was run
	logPath := p.config.LogFile
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(cwd, logPath)
	}

	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Return last 100 lines
	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}

	return lines, scanner.Err()
}

// setupLogging configures output redirection for the process
func (p *ProcessRunner) setupLogging(cmd Commander) error {
	if p.config.LogFile == "" {
		// No log file, use stdout/stderr
		cmd.SetStdout(os.Stdout)
		cmd.SetStderr(os.Stderr)
		return nil
	}

	// Get the current working directory (where devup was run)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Resolve log file path relative to where devup was run
	logPath := p.config.LogFile
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(cwd, logPath)
	}

	// Create log directory if needed
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	p.logFile = logFile

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	cmd.SetStdout(multiWriter)
	cmd.SetStderr(multiWriter)

	return nil
}

// monitor watches the process and handles unexpected exits.
// It is the only goroutine that calls Wait on the Cmd; Stop waits for monitorDone.
func (p *ProcessRunner) monitor() {
	defer close(p.monitorDone)
	if p.cmd == nil {
		return
	}
	err := p.cmd.Wait()
	if err != nil {
		fmt.Printf("Service '%s' exited with error: %v\n", p.config.Name, err)
	}
}

// waitForReady waits for the service to become healthy
func (p *ProcessRunner) waitForReady(ctx context.Context) error {
	timeout := 30 * time.Second
	if p.config.HealthCheck.Timeout > 0 {
		timeout = p.config.HealthCheck.Timeout
	}

	interval := 1 * time.Second
	if p.config.HealthCheck.Interval > 0 {
		interval = p.config.HealthCheck.Interval
	}

	retries := 30
	if p.config.HealthCheck.Retries > 0 {
		retries = p.config.HealthCheck.Retries
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := 0; i < retries; i++ {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("service '%s' did not become ready within timeout", p.config.Name)
		case <-ticker.C:
			if p.checkHealth() {
				return nil
			}
		}
	}

	return fmt.Errorf("service '%s' health check failed after %d retries", p.config.Name, retries)
}

// checkHealth performs a health check on the service
func (p *ProcessRunner) checkHealth() bool {
	switch p.config.HealthCheck.Type {
	case "http":
		return checkHTTPHealth(p.config.HealthCheck.Endpoint)
	case "tcp":
		return checkTCPHealth(p.config.Port)
	case "exec":
		return checkExecHealth(p.config.HealthCheck.Endpoint)
	default:
		// No health check configured, assume healthy after process starts
		return true
	}
}


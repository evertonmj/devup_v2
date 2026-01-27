package health

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"
)

// CheckHTTP performs an HTTP health check
func CheckHTTP(endpoint string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("HTTP health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("HTTP health check returned status %d", resp.StatusCode)
}

// CheckTCP performs a TCP port health check
func CheckTCP(host string, port int, timeout time.Duration) error {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("TCP health check failed: %w", err)
	}
	_ = conn.Close()
	return nil
}

// CheckExec performs a command-based health check
func CheckExec(command string, timeout time.Duration) error {
	cmd := exec.Command("bash", "-c", command)

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("exec health check failed to start: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return fmt.Errorf("exec health check timed out")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("exec health check failed: %w", err)
		}
		return nil
	}
}

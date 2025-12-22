package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CleanPort kills any process using the specified port
func CleanPort(port int) error {
	if port <= 0 {
		return nil
	}

	// Find processes using the port with lsof
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		// No process found on this port, which is fine
		return nil
	}

	// Parse PIDs from output
	pidsStr := strings.TrimSpace(string(output))
	if pidsStr == "" {
		return nil
	}

	pids := strings.Split(pidsStr, "\n")
	for _, pidStr := range pids {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Kill the process
		killCmd := exec.Command("kill", "-9", strconv.Itoa(pid))
		_ = killCmd.Run() // Ignore errors, process might already be dead
	}

	return nil
}

// CleanPorts cleans multiple ports
func CleanPorts(ports []int) error {
	for _, port := range ports {
		if err := CleanPort(port); err != nil {
			return fmt.Errorf("failed to clean port %d: %w", port, err)
		}
	}
	return nil
}

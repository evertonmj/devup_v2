package service

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"
)

// checkHTTPHealth performs an HTTP health check
func checkHTTPHealth(endpoint string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// checkTCPHealth performs a TCP port health check
func checkTCPHealth(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkExecHealth performs a command-based health check
func checkExecHealth(command string) bool {
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	return err == nil
}

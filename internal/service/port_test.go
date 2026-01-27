package service

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestCleanPort(t *testing.T) {
	err := CleanPort(0)
	if err != nil {
		t.Errorf("Expected CleanPort(0) to return nil, but got %v", err)
	}
	err = CleanPort(-1)
	if err != nil {
		t.Errorf("Expected CleanPort(-1) to return nil, but got %v", err)
	}
}

func TestCleanPorts(t *testing.T) {
	err := CleanPorts([]int{0, -1})
	if err != nil {
		t.Errorf("Expected CleanPorts([]int{0, -1}) to return nil, but got %v", err)
	}
}

func TestCleanPort_WithListener(t *testing.T) {
	helperPath := filepath.Join("testdata", "listen_port.go")
	if _, err := os.Stat(helperPath); err != nil {
		t.Skip("listen_port.go not found")
	}
	cmd := exec.Command("go", "run", "listen_port.go")
	cmd.Dir = filepath.Dir(helperPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Skipf("go run listener: %v", err)
	}
	defer cmd.Process.Kill()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatalf("could not read port")
	}
	portStr := scanner.Text()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid port %q: %v", portStr, err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := CleanPort(port); err != nil {
		t.Errorf("CleanPort(%d): %v", port, err)
	}
}

package service

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHTTPHealth(t *testing.T) {
	// Test case 1: Healthy server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if !checkHTTPHealth(server.URL) {
		t.Errorf("Expected HTTP health check to pass for a healthy server, but it failed")
	}

	// Test case 2: Unhealthy server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	if checkHTTPHealth(server.URL) {
		t.Errorf("Expected HTTP health check to fail for an unhealthy server, but it passed")
	}

	// Test case 3: Unreachable server
	if checkHTTPHealth("http://localhost:12345") {
		t.Errorf("Expected HTTP health check to fail for an unreachable server, but it passed")
	}
}

func TestCheckTCPHealth(t *testing.T) {
	// Test case 1: Port is open
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	if !checkTCPHealth(port) {
		t.Errorf("Expected TCP health check to pass for an open port, but it failed")
	}

	// Test case 2: Port is closed
	if checkTCPHealth(12345) {
		t.Errorf("Expected TCP health check to fail for a closed port, but it passed")
	}
}

func TestCheckExecHealth(t *testing.T) {
	// Test case 1: Command succeeds
	if !checkExecHealth("true") {
		t.Errorf("Expected exec health check to pass for a successful command, but it failed")
	}

	// Test case 2: Command fails
	if checkExecHealth("false") {
		t.Errorf("Expected exec health check to fail for a failing command, but it passed")
	}

	// Test case 3: Command does not exist
	if checkExecHealth("non-existent-command") {
		t.Errorf("Expected exec health check to fail for a non-existent command, but it passed")
	}
}

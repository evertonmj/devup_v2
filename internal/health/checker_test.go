package health

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestCheckHTTP(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful health check (200)",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "successful health check (204)",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "failed health check (500)",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "failed health check (404)",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "failed health check (400)",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Check health
			err := CheckHTTP(server.URL, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckHTTPTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Check should fail due to timeout
	err := CheckHTTP(server.URL, 100*time.Millisecond)
	if err == nil {
		t.Error("CheckHTTP() should fail on timeout")
	}
}

func TestCheckHTTPInvalidURL(t *testing.T) {
	err := CheckHTTP("http://localhost:59999", 1*time.Second)
	if err == nil {
		t.Error("CheckHTTP() should fail for unreachable server")
	}
}

func TestCheckTCP(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host and port from server
	// Note: httptest.Server.Listener.Addr() gives us the address
	addr := server.Listener.Addr().String()
	host := "localhost"
	port := 0
	_, err := fmt.Sscanf(addr, "%s:%d", &host, &port)
	if err != nil {
		// Try alternative format
		_, _ = fmt.Sscanf(addr, "[%s]:%d", &host, &port)
	}

	// If we can't parse, just test with the listening port
	if port == 0 {
		// Get port from the actual listener
		_, portStr, _ := net.SplitHostPort(server.Listener.Addr().String())
		port, _ = strconv.Atoi(portStr)
	}

	if port > 0 {
		err = CheckTCP("localhost", port, 5*time.Second)
		if err != nil {
			t.Errorf("CheckTCP() should succeed for listening port: %v", err)
		}
	}
}

func TestCheckTCPFailure(t *testing.T) {
	// Use a port that's definitely not listening
	err := CheckTCP("localhost", 59999, 1*time.Second)
	if err == nil {
		t.Error("CheckTCP() should fail for non-listening port")
	}
}

func TestCheckTCPInvalidHost(t *testing.T) {
	err := CheckTCP("invalid-host-that-does-not-exist.local", 80, 1*time.Second)
	if err == nil {
		t.Error("CheckTCP() should fail for invalid host")
	}
}

func TestCheckExec(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "successful command",
			command:  "true",
			expected: true,
		},
		{
			name:     "failing command",
			command:  "false",
			expected: false,
		},
		{
			name:     "echo command",
			command:  "echo 'test'",
			expected: true,
		},
		{
			name:     "exit 0",
			command:  "exit 0",
			expected: true,
		},
		{
			name:     "exit 1",
			command:  "exit 1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckExec(tt.command, 5*time.Second)
			if (err == nil) != tt.expected {
				t.Errorf("CheckExec() error = %v, want success = %v", err, tt.expected)
			}
		})
	}
}

func TestCheckExecTimeout(t *testing.T) {
	// Command that sleeps longer than timeout
	err := CheckExec("sleep 5", 100*time.Millisecond)
	if err == nil {
		t.Error("CheckExec() should timeout for long-running command")
	}
}

func TestCheckExecInvalidCommand(t *testing.T) {
	err := CheckExec("nonexistent-command-xyz-123", 1*time.Second)
	if err == nil {
		t.Error("CheckExec() should fail for nonexistent command")
	}
}

func TestCheckExecEmptyCommand(t *testing.T) {
	// Empty command should succeed (bash -c '' exits with 0)
	err := CheckExec("", 1*time.Second)
	if err != nil {
		t.Errorf("CheckExec() with empty command error = %v, want nil", err)
	}
}

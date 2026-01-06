package service

import (
	"testing"
)

func TestCleanPort(t *testing.T) {
	// We can't easily test the real CleanPort function as it requires a process to be running on a port.
	// However, we can test that it doesn't return an error for an invalid port.
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

package log

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestError_Nil(t *testing.T) {
	// Should not panic or write
	Error(nil)
}

func TestError_NonNil(t *testing.T) {
	stderr := os.Stderr
	defer func() { os.Stderr = stderr }()
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error(errors.New("test error"))

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "[devup] ERROR:") {
		t.Errorf("expected [devup] ERROR in output, got %q", out)
	}
	if !strings.Contains(out, "test error") {
		t.Errorf("expected error message in output, got %q", out)
	}
}

func TestErrorf(t *testing.T) {
	stderr := os.Stderr
	defer func() { os.Stderr = stderr }()
	r, w, _ := os.Pipe()
	os.Stderr = w

	Errorf("failed: %s", "xyz")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()
	if !strings.Contains(out, "[devup] ERROR:") {
		t.Errorf("expected [devup] ERROR in output, got %q", out)
	}
	if !strings.Contains(out, "failed: xyz") {
		t.Errorf("expected formatted message in output, got %q", out)
	}
}

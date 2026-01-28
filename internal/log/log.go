package log

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Error logs err and a full stack trace to stderr.
func Error(err error) {
	if err == nil {
		return
	}
	Errorf("%v", err)
}

// Errorf logs a formatted message and a full stack trace to stderr.
func Errorf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[devup] ERROR: %s\n", s)
	fmt.Fprintf(os.Stderr, "%s\n", debug.Stack())
}

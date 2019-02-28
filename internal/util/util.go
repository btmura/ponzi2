package util

import (
	"fmt"
	"runtime"
	"strings"
)

// Error wraps fmt.Error with file and line information.
func Error(a ...interface{}) error {
	return fmt.Errorf("%s: %s", fileLinePrefix(), fmt.Sprint(a...))
}

// Errorf wraps fmt.Errorf with file and line information.
func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s: %s", fileLinePrefix(), fmt.Sprintf(format, a...))
}

func fileLinePrefix() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

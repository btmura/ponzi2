package util

import (
	"fmt"
	"runtime"
	"strings"
)

// Errorf prepends file and line info to the error returned by fmt.Errorf.
func Errorf(format string, a ...interface{}) error {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	return fmt.Errorf("%s:%d: %s", file, line, fmt.Sprintf(format, a...))
}

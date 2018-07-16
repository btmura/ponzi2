// Package logger provides logging functions.
package logger

import (
	"fmt"

	"github.com/golang/glog"
)

// Info logs an Info message.
func Info(args ...interface{}) {
	if glog.V(2) {
		glog.InfoDepth(1, args...)
	}
}

// Infof logs an Info message
func Infof(format string, args ...interface{}) {
	if glog.V(2) {
		glog.InfoDepth(1, fmt.Sprintf(format, args...))
	}
}

// Fatal logs and exits.
func Fatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

// Fatalf logs and exits.
func Fatalf(format string, args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintf(format, args...))
}

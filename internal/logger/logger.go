// Package logger provides logging functions.
package logger

import (
	"github.com/golang/glog"
)

// Info logs an Info message.
func Info(args ...interface{}) {
	glog.V(2).Info(args...)
}

// Infof logs an Info message
func Infof(format string, args ...interface{}) {
	glog.V(2).Infof(format, args...)
}

// Fatal logs and exits.
func Fatal(args ...interface{}) {
	glog.Fatal(args...)
}

// Fatalf logs and exits.
func Fatalf(format string, args ...interface{}) {
	glog.Fatalf(format, args...)
}

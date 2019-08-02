package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

// Info logs an info message.
func Info(a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("I", file, line))
	fmt.Print(a...)
	fmt.Println()
}

// Infof logs an info message.
func Infof(format string, a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("I", file, line))
	fmt.Printf(format, a...)
	fmt.Println()
}

// Debug logs a debug message.
func Debug(a ...interface{}) {
	file, line := fileLine(2)
	if debugEnabled(file) {
		fmt.Print(linePrefix("D", file, line))
		fmt.Print(a...)
		fmt.Println()
	}
}

// Debugf logs a debug message.
func Debugf(format string, a ...interface{}) {
	file, line := fileLine(2)
	if debugEnabled(file) {
		fmt.Print(linePrefix("D", file, line))
		fmt.Printf(format, a...)
		fmt.Println()
	}
}

// Error logs an error message.
func Error(format string, a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("E", file, line))
	fmt.Print(a...)
	fmt.Println()
}

// Errorf logs an error message.
func Errorf(format string, a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("E", file, line))
	fmt.Printf(format, a...)
	fmt.Println()
}

// Fatal logs a fatal message and exits.
func Fatal(a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("F", file, line))
	fmt.Print(a...)
	fmt.Println()
	os.Exit(1)
}

// Fatalf logs a fatal message and calls exit.
func Fatalf(format string, a ...interface{}) {
	file, line := fileLine(2)
	fmt.Print(linePrefix("F", file, line))
	fmt.Printf(format, a...)
	fmt.Println()
	os.Exit(1)
}

func fileLine(skip int) (file string, line int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
	}

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	return file, line
}

func linePrefix(level, file string, line int) string {
	return fmt.Sprintf("%s %s %s:%d ", level, time.Now().Format("15:04:05"), file, line)
}

func debugEnabled(file string) bool {
	base := strings.TrimSuffix(file, path.Ext(file))
	return strings.Contains(os.Getenv("PZDEBUG"), base)
}

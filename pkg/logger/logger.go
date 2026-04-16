// Package logger provides structured logging for ODIN
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var levelNames = []string{"DEBUG", "INFO", "WARN", "ERROR"}

func (l Level) String() string {
	if l < DebugLevel || l > ErrorLevel {
		return "UNKNOWN"
	}
	return levelNames[l]
}

// Logger is a thread-safe structured logger
type Logger struct {
	mu     sync.Mutex
	output io.Writer
	level  Level
}

// global logger instance
var (
	defaultLogger = &Logger{output: os.Stdout, level: InfoLevel}
	mu            sync.Mutex
)

// SetOutput sets the output writer
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger.output = w
}

// SetLevel sets the log level
func SetLevel(level Level) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger.level = level
}

// Debug logs a debug message
func Debug(msg string, keyvals ...interface{}) {
	log(DebugLevel, msg, keyvals...)
}

// Info logs an info message
func Info(msg string, keyvals ...interface{}) {
	log(InfoLevel, msg, keyvals...)
}

// Warn logs a warning message
func Warn(msg string, keyvals ...interface{}) {
	log(WarnLevel, msg, keyvals...)
}

// Error logs an error message
func Error(msg string, keyvals ...interface{}) {
	log(ErrorLevel, msg, keyvals...)
}

func log(level Level, msg string, keyvals ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if level < defaultLogger.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()

	fmt.Fprintf(defaultLogger.output, "%s [%s] %s", timestamp, levelStr, msg)

	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			fmt.Fprintf(defaultLogger.output, " %v=%v", keyvals[i], keyvals[i+1])
		}
	}

	fmt.Fprintln(defaultLogger.output)
}

// With returns a new logger with additional context
func With(keyvals ...interface{}) *ContextLogger {
	return &ContextLogger{keyvals: keyvals}
}

// ContextLogger is a logger with additional context
type ContextLogger struct {
	keyvals []interface{}
}

// Info logs with context
func (c *ContextLogger) Info(msg string) {
	Info(msg, c.keyvals...)
}

// Error logs with context
func (c *ContextLogger) Error(msg string) {
	Error(msg, c.keyvals...)
}

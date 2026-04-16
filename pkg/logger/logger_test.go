package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(100), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"error", ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			// We can't directly check internal state, but we can verify
			// that SetLevel doesn't panic
		})
	}
}

func TestSetOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel) // Ensure info level is captured

	Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", output)
	}
}

func TestDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	Debug("debug test")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected DEBUG level in output, got: %s", output)
	}
	if !strings.Contains(output, "debug test") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(InfoLevel)

	Info("info test")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO level in output, got: %s", output)
	}
	if !strings.Contains(output, "info test") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(WarnLevel)

	Warn("warn test")

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected WARN level in output, got: %s", output)
	}
}

func TestError(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(ErrorLevel)

	Error("error test")

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected ERROR level in output, got: %s", output)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(ErrorLevel) // Only show errors

	Debug("should not appear")
	Info("should not appear")
	Warn("should not appear")
	Error("should appear")

	output := buf.String()

	if strings.Contains(output, "should not appear") {
		t.Error("expected filtered messages NOT to appear")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("expected error message to appear")
	}
}

func TestLogKeyvals(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	Info("test message", "key1", "value1", "key2", "value2")

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("expected key1=value1 in output, got: %s", output)
	}
	if !strings.Contains(output, "key2=value2") {
		t.Errorf("expected key2=value2 in output, got: %s", output)
	}
}

func TestLogOddKeyvals(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	// Odd number of keyvals - last one should be ignored
	Info("test", "key1", "value1", "key2")

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("expected key1=value1 in output, got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	ctx := With("service", "test", "request_id", "123")
	ctx.Info("request processed")

	output := buf.String()
	// Logger outputs: "service=test request_id=123"
	if !strings.Contains(output, "service=") {
		t.Errorf("expected 'service=' in output, got: %s", output)
	}
	if !strings.Contains(output, "request_id=") {
		t.Errorf("expected 'request_id=' in output, got: %s", output)
	}
	if !strings.Contains(output, "request processed") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestTimestampFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	Info("test")

	output := buf.String()
	// Timestamp format: 2006-01-02 15:04:05
	if len(output) < 19 {
		t.Errorf("expected timestamp in output, got: %s", output)
	}
}

func TestLogTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	Info("test message", "key", "value")

	output := buf.String()

	// Check that timestamp follows format YYYY-MM-DD HH:MM:SS
	parts := strings.SplitN(output, " ", 2)
	if len(parts) < 2 {
		t.Errorf("expected timestamp at start, got: %s", output)
	}
	timestamp := parts[0]
	if len(timestamp) != 10 || timestamp[4] != '-' || timestamp[7] != '-' {
		t.Errorf("expected timestamp format YYYY-MM-DD, got: %s", timestamp)
	}
}

func TestConcurrentLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(DebugLevel)

	// Run multiple goroutines logging concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Info("concurrent log", "id", id, "msg", j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock/data races, the test passed
	// (data races would be caught by -race flag)
	if buf.Len() == 0 {
		t.Error("expected some output from concurrent logging")
	}
}

// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"testing"
	"time"
)

func TestFlakyTracker_RecordRun(t *testing.T) {
	tracker := NewFlakyTracker(5)

	// Record some runs
	tracker.RecordRun("TestOne", true, 10*time.Millisecond)
	tracker.RecordRun("TestOne", true, 10*time.Millisecond)
	tracker.RecordRun("TestOne", false, 10*time.Millisecond) // Fail
	tracker.RecordRun("TestOne", true, 10*time.Millisecond)
	tracker.RecordRun("TestOne", true, 10*time.Millisecond)

	results := tracker.GetFlakyResults()
	if len(results) != 1 {
		t.Errorf("expected 1 test result, got %d", len(results))
	}

	if results[0].TestName != "TestOne" {
		t.Errorf("expected TestOne, got %s", results[0].TestName)
	}

	if results[0].RunCount != 5 {
		t.Errorf("expected 5 runs, got %d", results[0].RunCount)
	}

	if results[0].PassCount != 4 {
		t.Errorf("expected 4 passes, got %d", results[0].PassCount)
	}

	if results[0].IsFlaky {
		t.Error("test should not be marked flaky - it passed threshold after failing")
	}
}

func TestFlakyTracker_DetectsFlaky(t *testing.T) {
	tracker := NewFlakyTracker(3)

	// Intermittent failures
	tracker.RecordRun("TestIntermittent", true, 10*time.Millisecond)
	tracker.RecordRun("TestIntermittent", false, 10*time.Millisecond)
	tracker.RecordRun("TestIntermittent", true, 10*time.Millisecond)
	tracker.RecordRun("TestIntermittent", false, 10*time.Millisecond)
	tracker.RecordRun("TestIntermittent", true, 10*time.Millisecond)
	tracker.RecordRun("TestIntermittent", true, 10*time.Millisecond)

	results := tracker.GetFlakyResults()
	if len(results) != 1 {
		t.Errorf("expected 1 test result, got %d", len(results))
	}

	if !results[0].IsFlaky {
		t.Error("test with intermittent failures should be marked flaky")
	}

	if results[0].Consistency < 1.0 && results[0].Consistency > 0 {
		// Expected - test has consistency between 0 and 1
		t.Logf("Consistency: %.2f", results[0].Consistency)
	}
}

func TestFlakyTracker_Clear(t *testing.T) {
	tracker := NewFlakyTracker(3)

	tracker.RecordRun("TestClear", true, 10*time.Millisecond)
	tracker.RecordRun("TestClear", true, 10*time.Millisecond)

	tracker.Clear()

	results := tracker.GetFlakyResults()
	if len(results) != 0 {
		t.Errorf("expected 0 results after clear, got %d", len(results))
	}
}

func TestFlakyDetector_AnalyzeResult(t *testing.T) {
	detector := NewFlakyDetector(3)

	tests := []struct {
		name     string
		runs     []bool
		expected bool
		pattern  string
	}{
		{
			name:     "always passes",
			runs:     []bool{true, true, true, true},
			expected: false,
			pattern:  "intermittent",
		},
		{
			name:     "always fails",
			runs:     []bool{false, false, false, false},
			expected: false,
			pattern:  "consistent",
		},
		{
			name:     "intermittent failures",
			runs:     []bool{true, false, true, false, true, true},
			expected: true,
			pattern:  "intermittent",
		},
		{
			name:     "first fail then pass",
			runs:     []bool{false, true, true, true, true},
			expected: true,
			pattern:  "first-fail",
		},
		{
			name:     "last fail",
			runs:     []bool{true, true, true, true, false},
			expected: true,
			pattern:  "last-fail",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.AnalyzeResult(tc.runs)
			if result.IsFlaky != tc.expected {
				t.Errorf("expected IsFlaky=%v, got %v", tc.expected, result.IsFlaky)
			}
			if result.FailurePattern != tc.pattern {
				t.Errorf("expected pattern=%s, got %s", tc.pattern, result.FailurePattern)
			}
		})
	}
}

func TestBenchMarker_BenchmarkPhase(t *testing.T) {
	marker := NewBenchMarker()
	marker.SetIterations(10)

	result := marker.BenchmarkPhase("test_phase", func() error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	if result.Phase != "test_phase" {
		t.Errorf("expected phase test_phase, got %s", result.Phase)
	}

	if result.Iterations != 10 {
		t.Errorf("expected 10 iterations, got %d", result.Iterations)
	}

	if result.LatencyMs <= 0 {
		t.Error("expected latency > 0")
	}

	if result.Passed && result.LatencyMs > result.ThresholdMs {
		t.Error("if passed, latency should be under threshold")
	}
}

func TestBenchMarker_SetThreshold(t *testing.T) {
	marker := NewBenchMarker()

	marker.SetThreshold("custom_phase", 200.0)

	if marker.Thresholds["custom_phase"] != 200.0 {
		t.Errorf("expected threshold 200.0, got %f", marker.Thresholds["custom_phase"])
	}
}

func TestVerifyConfig_Defaults(t *testing.T) {
	cfg := DefaultVerifyConfig()

	if cfg.Timeout != 15*time.Minute {
		t.Errorf("expected 15m timeout, got %v", cfg.Timeout)
	}

	if cfg.FlakyThreshold != 3 {
		t.Errorf("expected flaky threshold 3, got %d", cfg.FlakyThreshold)
	}

	if len(cfg.MatrixTargets) != 3 {
		t.Errorf("expected 3 matrix targets, got %d", len(cfg.MatrixTargets))
	}
}

func TestNornir_New(t *testing.T) {
	n := New(nil)

	if n == nil {
		t.Error("New() should not return nil")
	}

	if n.config == nil {
		t.Error("config should not be nil")
	}

	if n.flakyTrack == nil {
		t.Error("flakyTrack should not be nil")
	}

	if n.benchMarker == nil {
		t.Error("benchMarker should not be nil")
	}
}

func TestMatrixRunner_GetTargets(t *testing.T) {
	runner := NewMatrixRunner()
	targets := runner.GetTargets()

	if len(targets) == 0 {
		t.Error("expected at least one target")
	}
}

func TestMatrixRunner_CurrentPlatform(t *testing.T) {
	runner := NewMatrixRunner()
	current := runner.CurrentPlatform()

	if current.OS == "" {
		t.Error("expected non-empty OS")
	}

	if current.Arch == "" {
		t.Error("expected non-empty Arch")
	}
}

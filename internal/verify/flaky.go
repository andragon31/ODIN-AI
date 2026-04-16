// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"fmt"
	"sync"
	"time"
)

// FlakyTracker tracks test flakiness
type FlakyTracker struct {
	mu        sync.RWMutex
	threshold int
	testRuns  map[string]*testRunHistory
}

// testRunHistory tracks the history of test runs
type testRunHistory struct {
	mu        sync.RWMutex
	PassCount int
	FailCount int
	RunCount  int
	Failures  []int // Run indices where test failed
	Durations []time.Duration
}

// NewFlakyTracker creates a new flaky test tracker
func NewFlakyTracker(threshold int) *FlakyTracker {
	if threshold < 1 {
		threshold = 3
	}
	return &FlakyTracker{
		threshold: threshold,
		testRuns:  make(map[string]*testRunHistory),
	}
}

// RecordRun records a test run result
func (ft *FlakyTracker) RecordRun(testName string, passed bool, duration time.Duration) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	history, exists := ft.testRuns[testName]
	if !exists {
		history = &testRunHistory{
			Failures:  make([]int, 0),
			Durations: make([]time.Duration, 0),
		}
		ft.testRuns[testName] = history
	}

	history.mu.Lock()
	defer history.mu.Unlock()

	history.RunCount++
	history.Durations = append(history.Durations, duration)

	if passed {
		history.PassCount++
	} else {
		history.FailCount++
		history.Failures = append(history.Failures, history.RunCount)
	}
}

// GetFlakyResults returns the list of flaky test results
func (ft *FlakyTracker) GetFlakyResults() []FlakyResult {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	results := make([]FlakyResult, 0, len(ft.testRuns))

	for testName, history := range ft.testRuns {
		history.mu.Lock()

		result := FlakyResult{
			TestName:    testName,
			RunCount:    history.RunCount,
			PassCount:   history.PassCount,
			FailedRuns:  make([]int, len(history.Failures)),
			AvgDuration: calculateAvgDuration(history.Durations),
		}
		copy(result.FailedRuns, history.Failures)

		// Calculate consistency (ratio of passes to total runs)
		if history.RunCount > 0 {
			result.Consistency = float64(history.PassCount) / float64(history.RunCount)
		}

		// A test is considered flaky if:
		// 1. It has failures AND
		// 2. It eventually passes the threshold number of times
		// 3. Consistency is below 100% but above 0
		result.IsFlaky = history.FailCount > 0 &&
			result.Consistency > 0 &&
			result.Consistency < 1.0 &&
			history.PassCount >= ft.threshold

		history.mu.Unlock()
		results = append(results, result)
	}

	return results
}

// calculateAvgDuration calculates the average duration
func calculateAvgDuration(durations []time.Duration) float64 {
	if len(durations) == 0 {
		return 0
	}

	var total float64
	for _, d := range durations {
		total += d.Seconds() * 1000 // Convert to ms
	}

	return total / float64(len(durations))
}

// GetTestHistory returns the run history for a specific test
func (ft *FlakyTracker) GetTestHistory(testName string) *testRunHistory {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	return ft.testRuns[testName]
}

// Clear clears all tracking data
func (ft *FlakyTracker) Clear() {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	ft.testRuns = make(map[string]*testRunHistory)
}

// FlakyDetector implements algorithmic flaky test detection
type FlakyDetector struct {
	mu           sync.RWMutex
	runsRequired int
}

// NewFlakyDetector creates a new flaky test detector
func NewFlakyDetector(runsRequired int) *FlakyDetector {
	if runsRequired < 3 {
		runsRequired = 3
	}
	return &FlakyDetector{
		runsRequired: runsRequired,
	}
}

// DetectionResult holds the result of flaky detection
type DetectionResult struct {
	TestName       string
	IsFlaky        bool
	FailurePattern string // "intermittent", "first-fail", "last-fail", "consistent"
	Confidence     float64
	Message        string
}

// AnalyzeResult analyzes a test result for flakiness patterns
func (fd *FlakyDetector) AnalyzeResult(runs []bool) DetectionResult {
	if len(runs) < fd.runsRequired {
		return DetectionResult{
			IsFlaky:    false,
			Confidence: 0,
			Message:    fmt.Sprintf("insufficient data: need at least %d runs", fd.runsRequired),
		}
	}

	failCount := 0
	firstFail := -1
	lastFail := -1

	for i, passed := range runs {
		if !passed {
			failCount++
			if firstFail == -1 {
				firstFail = i
			}
			lastFail = i
		}
	}

	result := DetectionResult{
		TestName:       "", // Will be set by caller
		FailurePattern: "intermittent",
		Confidence:     0,
	}

	if failCount == 0 {
		result.IsFlaky = false
		result.Message = "test always passes"
		result.Confidence = 1.0
		return result
	}

	if failCount == len(runs) {
		result.IsFlaky = false // Not flaky, just broken
		result.FailurePattern = "consistent"
		result.Message = "test always fails (not flaky, broken)"
		result.Confidence = 1.0
		return result
	}

	// Detect pattern
	if firstFail == 0 && lastFail == len(runs)-1 {
		result.FailurePattern = "first-fail"
		result.Message = "test fails on first run then passes"
		result.Confidence = 0.9
	} else if lastFail == len(runs)-1 && firstFail > 0 {
		result.FailurePattern = "last-fail"
		result.Message = "test passes then fails on last run"
		result.Confidence = 0.9
	} else {
		// Check if failures are sporadic
		failRatio := float64(failCount) / float64(len(runs))
		if failRatio > 0.5 {
			result.FailurePattern = "high-failure"
			result.Message = "test fails more than 50% of runs"
			result.Confidence = 0.95
		} else {
			result.FailurePattern = "intermittent"
			result.Message = "test has intermittent failures"
			result.Confidence = 0.8
		}
	}

	result.IsFlaky = true
	return result
}

// Package verify provides Nornir - the verification suite for ODIN
// Includes flaky test detection, benchmarking, and multi-OS matrix testing
package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// VerifyConfig holds Nornir configuration
type VerifyConfig struct {
	Timeout        time.Duration `json:"timeout"`
	FlakyThreshold int           `json:"flaky_threshold"`
	MatrixTargets  []string      `json:"matrix_targets"`
}

// Nornir is the verification suite
type Nornir struct {
	config      *VerifyConfig
	flakyTrack  *FlakyTracker
	benchMarker *BenchMarker
}

// FlakyResult represents the result of flaky test detection
type FlakyResult struct {
	TestName    string  `json:"test_name"`
	RunCount    int     `json:"run_count"`
	PassCount   int     `json:"pass_count"`
	IsFlaky     bool    `json:"is_flaky"`
	Consistency float64 `json:"consistency"`
	AvgDuration float64 `json:"avg_duration_ms"`
	FailedRuns  []int   `json:"failed_runs,omitempty"`
}

// BenchmarkResult represents latency benchmark results
type BenchmarkResult struct {
	Phase        string  `json:"phase"`
	LatencyMs    float64 `json:"latency_ms"`
	ThresholdMs  float64 `json:"threshold_ms"`
	Passed       bool    `json:"passed"`
	Iterations   int     `json:"iterations"`
	MinLatencyMs float64 `json:"min_latency_ms"`
	MaxLatencyMs float64 `json:"max_latency_ms"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P50LatencyMs float64 `json:"p50_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`
}

// MatrixResult holds multi-OS test matrix results
type MatrixResult struct {
	Target   string        `json:"target"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	Tests    []TestResult  `json:"tests"`
	Errors   []string      `json:"errors,omitempty"`
}

// TestResult represents a single test result
type TestResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	IsFlaky  bool          `json:"is_flaky,omitempty"`
	Retries  int           `json:"retries,omitempty"`
	Output   string        `json:"output,omitempty"`
	ErrorMsg string        `json:"error_msg,omitempty"`
}

// Report represents the full verification report
type Report struct {
	GeneratedAt   time.Time         `json:"generated_at"`
	Platform      string            `json:"platform"`
	NornirVersion string            `json:"nornir_version"`
	Duration      time.Duration     `json:"duration"`
	Summary       ReportSummary     `json:"summary"`
	FlakyTests    []FlakyResult     `json:"flaky_tests,omitempty"`
	Benchmarks    []BenchmarkResult `json:"benchmarks,omitempty"`
	Matrix        []MatrixResult    `json:"matrix,omitempty"`
}

// ReportSummary holds summary statistics
type ReportSummary struct {
	TotalTests        int     `json:"total_tests"`
	PassedTests       int     `json:"passed_tests"`
	FailedTests       int     `json:"failed_tests"`
	FlakyTests        int     `json:"flaky_tests"`
	PassRate          float64 `json:"pass_rate"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	AllBenchmarksPass bool    `json:"all_benchmarks_pass"`
}

// DefaultVerifyConfig returns the default verify configuration
func DefaultVerifyConfig() *VerifyConfig {
	return &VerifyConfig{
		Timeout:        15 * time.Minute,
		FlakyThreshold: 3,
		MatrixTargets:  []string{"ubuntu", "macos", "windows"},
	}
}

// New creates a new Nornir verification suite
func New(cfg *VerifyConfig) *Nornir {
	if cfg == nil {
		cfg = DefaultVerifyConfig()
	}
	return &Nornir{
		config:      cfg,
		flakyTrack:  NewFlakyTracker(cfg.FlakyThreshold),
		benchMarker: NewBenchMarker(),
	}
}

// RunVerification runs the full verification suite
func (n *Nornir) RunVerification() (*Report, error) {
	start := time.Now()

	report := &Report{
		GeneratedAt:   time.Now(),
		Platform:      runtime.GOOS,
		NornirVersion: "1.0",
		Summary:       ReportSummary{},
	}

	// Run flaky test detection
	flakyResults := n.DetectFlakyTests()
	report.FlakyTests = flakyResults

	// Run benchmarks
	benchResults := n.RunBenchmarks()
	report.Benchmarks = benchResults

	report.Duration = time.Since(start)

	// Calculate summary
	n.calculateSummary(report)

	return report, nil
}

// DetectFlakyTests runs flaky test detection on known test patterns
func (n *Nornir) DetectFlakyTests() []FlakyResult {
	results := n.flakyTrack.GetFlakyResults()
	return results
}

// RunBenchmarks runs latency benchmarks for critical phases
func (n *Nornir) RunBenchmarks() []BenchmarkResult {
	results := n.benchMarker.RunAllBenchmarks()
	return results
}

// RunMatrixTesting runs tests across multiple OS targets
func (n *Nornir) RunMatrixTesting() []MatrixResult {
	targets := n.config.MatrixTargets
	results := make([]MatrixResult, len(targets))

	for i, target := range targets {
		result := n.runTestsForTarget(target)
		results[i] = result
	}

	return results
}

// runTestsForTarget runs tests for a specific OS target
func (n *Nornir) runTestsForTarget(target string) MatrixResult {
	result := MatrixResult{
		Target:   target,
		Passed:   true,
		Duration: time.Now().Sub(time.Now()), // Zero duration placeholder
		Tests:    []TestResult{},
	}

	// Simulate test execution - in real impl would use testcontainers
	// For now, report as not implemented
	result.Errors = []string{"matrix testing requires testcontainers-go integration"}
	result.Passed = false

	return result
}

// calculateSummary calculates report summary statistics
func (n *Nornir) calculateSummary(report *Report) {
	totalTests := 0
	passedTests := 0
	failedTests := 0
	flakyCount := 0

	for _, fr := range report.FlakyTests {
		totalTests++
		if fr.IsFlaky {
			flakyCount++
			passedTests++ // Flaky tests that passed their retry threshold
		}
	}

	if totalTests > 0 {
		report.Summary.PassRate = float64(passedTests) / float64(totalTests) * 100
	}

	report.Summary.TotalTests = totalTests
	report.Summary.PassedTests = passedTests
	report.Summary.FailedTests = failedTests
	report.Summary.FlakyTests = flakyCount

	// Check benchmarks
	allPass := true
	var totalLatency float64
	for _, b := range report.Benchmarks {
		totalLatency += b.LatencyMs
		if !b.Passed {
			allPass = false
		}
	}
	report.Summary.AllBenchmarksPass = allPass
	if len(report.Benchmarks) > 0 {
		report.Summary.AvgLatencyMs = totalLatency / float64(len(report.Benchmarks))
	}
}

// ExportReport exports the report to JSON
func (n *Nornir) ExportReport(report *Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

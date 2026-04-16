// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ReportGenerator generates verification reports
type ReportGenerator struct {
	format string
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(format string) *ReportGenerator {
	if format == "" {
		format = "json"
	}
	return &ReportGenerator{format: format}
}

// Generate creates a verification report
func (rg *ReportGenerator) Generate(n *Nornir) (*Report, error) {
	report := &Report{
		GeneratedAt:   time.Now(),
		Platform:      runtime.GOOS,
		NornirVersion: "1.0",
	}

	return report, nil
}

// GenerateFromResults creates a report from raw results
func (rg *ReportGenerator) GenerateFromResults(
	flaky []FlakyResult,
	benchmarks []BenchmarkResult,
	matrix []MatrixResult,
) (*Report, error) {
	report := &Report{
		GeneratedAt:   time.Now(),
		Platform:      runtime.GOOS,
		NornirVersion: "1.0",
		FlakyTests:    flaky,
		Benchmarks:    benchmarks,
		Matrix:        matrix,
	}

	// Calculate summary
	rg.calculateSummary(report)

	return report, nil
}

// calculateSummary computes summary statistics
func (rg *ReportGenerator) calculateSummary(report *Report) {
	totalTests := 0
	passedTests := 0
	failedTests := 0
	flakyCount := 0

	for _, fr := range report.FlakyTests {
		totalTests++
		if fr.IsFlaky {
			flakyCount++
		}
		if fr.PassCount >= fr.RunCount {
			passedTests++
		} else {
			failedTests++
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

// ExportJSON exports the report as JSON
func (rg *ReportGenerator) ExportJSON(report *Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// ExportJSONToWriter exports the report as JSON to stdout
func (rg *ReportGenerator) ExportJSONToWriter(report *Report, w *os.File) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// ReportSummary holds summary statistics for a report
type ReportSummary struct {
	TotalTests        int     `json:"total_tests"`
	PassedTests       int     `json:"passed_tests"`
	FailedTests       int     `json:"failed_tests"`
	FlakyTests        int     `json:"flaky_tests"`
	PassRate          float64 `json:"pass_rate"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	AllBenchmarksPass bool    `json:"all_benchmarks_pass"`
}

// CompactReport returns a compact summary of the report
func (rg *ReportGenerator) CompactReport(report *Report) string {
	summary := fmt.Sprintf(`Verification Report
=================
Platform: %s
Generated: %s

Summary:
  Total Tests: %d
  Passed: %d
  Failed: %d
  Flaky: %d
  Pass Rate: %.1f%%

Benchmarks:
`,
		report.Platform,
		report.GeneratedAt.Format(time.RFC3339),
		report.Summary.TotalTests,
		report.Summary.PassedTests,
		report.Summary.FailedTests,
		report.Summary.FlakyTests,
		report.Summary.PassRate,
	)

	for _, b := range report.Benchmarks {
		status := "✓ PASS"
		if !b.Passed {
			status = "✗ FAIL"
		}
		summary += fmt.Sprintf("  %s %s (%.2fms / %.2fms threshold)\n",
			status, b.Phase, b.LatencyMs, b.ThresholdMs)
	}

	return summary
}

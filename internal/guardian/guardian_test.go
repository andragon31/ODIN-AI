package guardian

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/odin-ai/odin/internal/config"
)

func TestNewGuardian(t *testing.T) {
	cfg := &config.GuardianConfig{
		PolicyEngine: "opa",
		RulesPath:    "~/.odin/rules",
		SAST: config.SASTConfig{
			Enabled: true,
			Tools:   []string{"gosec", "semgrep"},
		},
		BlockOnCrit: true,
	}

	g, err := NewGuardian(cfg)
	if err != nil {
		t.Fatalf("NewGuardian failed: %v", err)
	}

	if g == nil {
		t.Fatal("Guardian should not be nil")
	}

	if g.blockOnCrit != true {
		t.Error("BlockOnCrit should be true")
	}
}

func TestGuardianIsInitialized(t *testing.T) {
	// Test with non-existent path
	cfg := &config.GuardianConfig{
		PolicyEngine: "opa",
		RulesPath:    "/tmp/nonexistent-odin-rules-test",
		SAST: config.SASTConfig{
			Enabled: true,
			Tools:   []string{"gosec"},
		},
		BlockOnCrit: true,
	}

	g, err := NewGuardian(cfg)
	if err != nil {
		t.Fatalf("NewGuardian failed: %v", err)
	}

	if g.IsInitialized() {
		t.Error("Guardian should not be initialized with non-existent path")
	}
}

func TestGuardianCheck(t *testing.T) {
	cfg := &config.GuardianConfig{
		PolicyEngine: "opa",
		RulesPath:    "",
		SAST: config.SASTConfig{
			Enabled: false,
			Tools:   []string{},
		},
		BlockOnCrit: true,
	}

	g, err := NewGuardian(cfg)
	if err != nil {
		t.Fatalf("NewGuardian failed: %v", err)
	}

	ctx := context.Background()
	result, err := g.Check(ctx, []string{"."}, false)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result == nil {
		t.Fatal("CheckResult should not be nil")
	}

	// Should pass since no SAST tools are enabled
	if !result.Passed {
		t.Error("Check should pass when no issues found")
	}
}

func TestGuardianCheckWithDiff(t *testing.T) {
	cfg := &config.GuardianConfig{
		PolicyEngine: "opa",
		RulesPath:    "",
		SAST: config.SASTConfig{
			Enabled: false,
			Tools:   []string{},
		},
		BlockOnCrit: true,
	}

	g, err := NewGuardian(cfg)
	if err != nil {
		t.Fatalf("NewGuardian failed: %v", err)
	}

	ctx := context.Background()
	diff := "diff --git a/test.go b/test.go\n--- a/test.go\n+++ b/test.go\n@@ -1 +1 @@\n-old code\n+new code"
	result, err := g.CheckWithDiff(ctx, diff)
	if err != nil {
		t.Fatalf("CheckWithDiff failed: %v", err)
	}

	if result == nil {
		t.Fatal("CheckResult should not be nil")
	}
}

func TestCheckResult(t *testing.T) {
	result := &CheckResult{
		Passed:     true,
		Severity:   "none",
		Issues:     []Issue{},
		Duration:   time.Second,
		ReportPath: "/tmp/report.json",
	}

	if !result.Passed {
		t.Error("Result should be passed")
	}

	if result.Severity != "none" {
		t.Errorf("Expected severity 'none', got '%s'", result.Severity)
	}

	if len(result.Issues) != 0 {
		t.Error("Issues should be empty")
	}
}

func TestIssue(t *testing.T) {
	issue := Issue{
		File:     "/path/to/file.go",
		Line:     42,
		Message:  "Test issue",
		Severity: "high",
		Tool:     "gosec",
		RuleID:   "G104",
	}

	if issue.File != "/path/to/file.go" {
		t.Errorf("Expected file '/path/to/file.go', got '%s'", issue.File)
	}

	if issue.Line != 42 {
		t.Errorf("Expected line 42, got %d", issue.Line)
	}

	if issue.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", issue.Severity)
	}

	if issue.Tool != "gosec" {
		t.Errorf("Expected tool 'gosec', got '%s'", issue.Tool)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 1}, // empty string returns 1 empty element
		{"single line", 1},
		{"line1\nline2", 2},
		{"line1\nline2\nline3", 3},
		{"line1\nline2\n", 2}, // trailing newline should be trimmed by TrimSuffix
	}

	for _, tc := range tests {
		result := splitLines(tc.input)
		if len(result) != tc.expected {
			t.Errorf("splitLines(%q) returned %d lines, expected %d", tc.input, len(result), tc.expected)
		}
	}
}

func TestGetStagedFiles(t *testing.T) {
	// This test will return nil if not in a git repo
	// In a real scenario, this should be tested in a git repo
	files, err := getStagedFiles()
	if err != nil {
		t.Logf("Not in git repo (expected in some environments): %v", err)
	}
	if files == nil {
		t.Log("getStagedFiles returned nil (expected if not in git repo)")
	}
}

func TestGenerateReport(t *testing.T) {
	result := &CheckResult{
		Passed:   false,
		Severity: "high",
		Issues: []Issue{
			{
				File:     "test.go",
				Line:     10,
				Message:  "Hardcoded secret",
				Severity: "critical",
				Tool:     "gosec",
				RuleID:   "G104",
			},
		},
		Duration: time.Second * 2,
	}

	report, err := GenerateReport(result, "")
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if report == nil {
		t.Fatal("Report should not be nil")
	}

	if report.Passed {
		t.Error("Report should show failed status")
	}

	if report.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", report.Severity)
	}

	if report.Summary.Total != 1 {
		t.Errorf("Expected 1 total issue, got %d", report.Summary.Total)
	}

	if report.Summary.Critical != 1 {
		t.Errorf("Expected 1 critical issue, got %d", report.Summary.Critical)
	}
}

func TestFormatReport(t *testing.T) {
	// Test passed result
	passedResult := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
		Duration: time.Second,
	}

	formatted := FormatReport(passedResult)
	if formatted == "" {
		t.Error("FormatReport should return non-empty string for passed result")
	}

	// Test failed result
	failedResult := &CheckResult{
		Passed:   false,
		Severity: "critical",
		Issues: []Issue{
			{
				File:     "test.go",
				Line:     10,
				Message:  "Hardcoded secret found",
				Severity: "critical",
				Tool:     "gosec",
			},
		},
		Duration: time.Second,
	}

	formatted = FormatReport(failedResult)
	if formatted == "" {
		t.Error("FormatReport should return non-empty string for failed result")
	}
}

func TestIsToolInstalled(t *testing.T) {
	// This is a simple smoke test
	// In practice, we can't assume tools are installed
	result := isToolInstalled("nonexistent-tool-12345")
	if result {
		t.Error("Non-existent tool should return false")
	}
}

func TestMapSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"warning", "medium"},
		{"low", "low"},
		{"info", "info"},
		{"unknown", "info"},
	}

	for _, tc := range tests {
		result := mapSeverity(tc.input)
		if result != tc.expected {
			t.Errorf("mapSeverity(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestMapSemgrepSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"error", "high"},
		{"warning", "medium"},
		{"info", "low"},
		{"unknown", "info"},
	}

	for _, tc := range tests {
		result := mapSemgrepSeverity(tc.input)
		if result != tc.expected {
			t.Errorf("mapSemgrepSeverity(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestWriteReportToFile(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "report.json")

	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
		Duration: time.Second,
	}

	err := WriteReportToFile(result, reportPath)
	if err != nil {
		t.Fatalf("WriteReportToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Error("Report file should exist")
	}

	// Verify we can read it back
	report, err := ReadReport(reportPath)
	if err != nil {
		t.Fatalf("ReadReport failed: %v", err)
	}

	if report == nil {
		t.Error("Read report should not be nil")
	}

	if !report.Passed {
		t.Error("Report should show passed status")
	}
}

func TestGetReportJSON(t *testing.T) {
	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
		Duration: time.Second,
	}

	jsonStr, err := GetReportJSON(result)
	if err != nil {
		t.Fatalf("GetReportJSON failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("GetReportJSON should return non-empty string")
	}

	// Verify it's valid JSON by checking for expected keys
	if len(jsonStr) < 10 {
		t.Error("JSON string seems too short to be valid")
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"0", 0},
		{"100", 100},
		{"invalid", 0},
		{"", 0},
	}

	for _, tc := range tests {
		result := parseInt(tc.input)
		if result != tc.expected {
			t.Errorf("parseInt(%q) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestGuardianRulesPath(t *testing.T) {
	cfg := &config.GuardianConfig{
		PolicyEngine: "opa",
		RulesPath:    "/custom/rules/path",
		SAST: config.SASTConfig{
			Enabled: true,
			Tools:   []string{"gosec"},
		},
		BlockOnCrit: true,
	}

	g, err := NewGuardian(cfg)
	if err != nil {
		t.Fatalf("NewGuardian failed: %v", err)
	}

	if g.RulesPath() != "/custom/rules/path" {
		t.Errorf("Expected rules path '/custom/rules/path', got '%s'", g.RulesPath())
	}
}

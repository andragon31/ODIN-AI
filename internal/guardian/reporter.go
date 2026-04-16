package guardian

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Report represents a CI/CD security report
type Report struct {
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    float64   `json:"duration_seconds"`
	Passed      bool      `json:"passed"`
	Severity    string    `json:"severity"`
	Summary     Summary   `json:"summary"`
	Issues      []Issue   `json:"issues"`
	Remediation []string  `json:"remediation,omitempty"`
}

// Summary contains counts by severity
type Summary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Total    int `json:"total"`
}

// GenerateReport creates a JSON report from a CheckResult
func GenerateReport(result *CheckResult, outputPath string) (*Report, error) {
	report := &Report{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		Duration:  result.Duration.Seconds(),
		Passed:    result.Passed,
		Severity:  result.Severity,
		Issues:    result.Issues,
	}

	// Count issues by severity
	summary := Summary{}
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		default:
			summary.Info++
		}
		summary.Total++
	}
	report.Summary = summary

	// Generate remediation suggestions
	report.Remediation = generateRemediation(result.Issues)

	// Write to file if output path specified
	if outputPath != "" {
		if err := writeReport(report, outputPath); err != nil {
			return nil, fmt.Errorf("failed to write report: %w", err)
		}
		reportPath := outputPath
		result.ReportPath = outputPath
		_ = reportPath // silence unused warning
	}

	return report, nil
}

// writeReport writes the report to a JSON file
func writeReport(report *Report, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// generateRemediation generates remediation suggestions for issues
func generateRemediation(issues []Issue) []string {
	seen := make(map[string]bool)
	var remediations []string

	for _, issue := range issues {
		var remediation string

		switch {
		case issue.RuleID == "G104" || issue.RuleID == "gosec-hardcoded":
			remediation = "Use environment variables or secure secret management instead of hardcoded secrets"
		case issue.RuleID == "G107" || issue.RuleID == "gosec-url-redirect":
			remediation = "Validate and sanitize URL redirects to prevent open redirect vulnerabilities"
		case issue.RuleID == "G204" || issue.RuleID == "gosec-command-injection":
			remediation = "Avoid shell commands with user input. Use parameterized exec instead"
		case issue.RuleID == "G304" || issue.RuleID == "gosec-file-access":
			remediation = "Use secure path resolution and validate file paths to prevent path traversal"
		case issue.RuleID == "G505" || issue.RuleID == "gosec-weak-crypto":
			remediation = "Use strong cryptographic algorithms (AES-256, RSA-2048+)"
		case issue.Tool == "semgrep" && issue.RuleID != "":
			remediation = fmt.Sprintf("Review %s rule at: https://semgrep.dev/rules/%s", issue.RuleID, issue.RuleID)
		case issue.Severity == "critical":
			remediation = "Critical issue detected - immediate action required"
		case issue.Severity == "high":
			remediation = "High severity issue - address before production deployment"
		case issue.Severity == "medium":
			remediation = "Medium severity issue - address in next sprint"
		default:
			remediation = fmt.Sprintf("Review code at %s:%d", issue.File, issue.Line)
		}

		if !seen[remediation] {
			seen[remediation] = true
			remediations = append(remediations, remediation)
		}
	}

	return remediations
}

// WriteReportToFile writes a CheckResult directly to a file
func WriteReportToFile(result *CheckResult, outputPath string) error {
	report, err := GenerateReport(result, "")
	if err != nil {
		return err
	}

	return writeReport(report, outputPath)
}

// GetReportJSON returns the JSON string for a report
func GetReportJSON(result *CheckResult) (string, error) {
	report, err := GenerateReport(result, "")
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	return string(data), nil
}

// ReadReport reads a report from a JSON file
func ReadReport(inputPath string) (*Report, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read report file: %w", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse report: %w", err)
	}

	return &report, nil
}

// FormatReport formats a report for terminal output
func FormatReport(result *CheckResult) string {
	if result.Passed {
		return fmt.Sprintf("✅ Security check passed (%.2fs)", result.Duration.Seconds())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("❌ Security check failed (%.2fs)\n", result.Duration.Seconds()))

	if len(result.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("\nFound %d issues:\n", len(result.Issues)))

		// Group by severity
		critical := []Issue{}
		high := []Issue{}
		medium := []Issue{}
		low := []Issue{}

		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				critical = append(critical, issue)
			case "high":
				high = append(high, issue)
			case "medium":
				medium = append(medium, issue)
			default:
				low = append(low, issue)
			}
		}

		if len(critical) > 0 {
			sb.WriteString(fmt.Sprintf("\n🔴 CRITICAL (%d):\n", len(critical)))
			for _, issue := range critical {
				sb.WriteString(fmt.Sprintf("  [%s:%d] %s\n", issue.File, issue.Line, issue.Message))
			}
		}

		if len(high) > 0 {
			sb.WriteString(fmt.Sprintf("\n🟠 HIGH (%d):\n", len(high)))
			for _, issue := range high {
				sb.WriteString(fmt.Sprintf("  [%s:%d] %s\n", issue.File, issue.Line, issue.Message))
			}
		}

		if len(medium) > 0 {
			sb.WriteString(fmt.Sprintf("\n🟡 MEDIUM (%d):\n", len(medium)))
			for _, issue := range medium {
				sb.WriteString(fmt.Sprintf("  [%s:%d] %s\n", issue.File, issue.Line, issue.Message))
			}
		}

		if len(low) > 0 {
			sb.WriteString(fmt.Sprintf("\n🔵 LOW (%d):\n", len(low)))
			for _, issue := range low {
				sb.WriteString(fmt.Sprintf("  [%s:%d] %s\n", issue.File, issue.Line, issue.Message))
			}
		}
	}

	return sb.String()
}

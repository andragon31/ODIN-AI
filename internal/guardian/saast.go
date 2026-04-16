package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// SASTTool represents a static analysis tool
type SASTTool struct {
	Name    string
	Version string
}

// runGosec runs gosec security checker
func (g *Guardian) runGosec(ctx context.Context, files []string) (*CheckResult, error) {
	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
	}

	// Check if gosec is available
	if !isToolInstalled("gosec") {
		logger.Warn("gosec not installed, skipping")
		return result, nil
	}

	// Build gosec command
	args := []string{"-fmt", "json", "-quiet"}

	// Add files if specified (gosec will scan current directory if none)
	if len(files) > 0 && files[0] != "." {
		args = append(args, files...)
	}

	cmd := exec.CommandContext(ctx, "gosec", args...)
	output, err := cmd.Output()

	// gosec returns non-zero on findings, which is expected
	if err != nil && len(output) == 0 {
		// No output means it failed to run properly
		return nil, fmt.Errorf("gosec execution failed: %w", err)
	}

	// Parse gosec JSON output
	if len(output) > 0 {
		var gosecOutput GosecOutput
		if err := json.Unmarshal(output, &gosecOutput); err != nil {
			logger.Warn("Failed to parse gosec output", "error", err)
			return result, nil
		}

		for _, issue := range gosecOutput.Issues {
			severity := mapSeverity(issue.Severity)
			lineNum := parseInt(issue.Line)
			result.Issues = append(result.Issues, Issue{
				File:     issue.File,
				Line:     lineNum,
				Message:  issue.Details,
				Severity: severity,
				Tool:     "gosec",
				RuleID:   issue.RuleID,
			})

			if !result.Passed {
				if severity == "critical" {
					result.Severity = "critical"
				} else if severity == "high" && result.Severity != "critical" {
					result.Severity = "high"
				} else if severity == "medium" && result.Severity == "none" {
					result.Severity = "medium"
				}
			}

			if severity != "info" && severity != "low" {
				result.Passed = false
			}
		}
	}

	return result, nil
}

// runSemgrep runs semgrep static analyzer
func (g *Guardian) runSemgrep(ctx context.Context, files []string) (*CheckResult, error) {
	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
	}

	// Check if semgrep is available
	if !isToolInstalled("semgrep") {
		logger.Warn("semgrep not installed, skipping")
		return result, nil
	}

	// Build semgrep command
	args := []string{"--json", "--quiet"}

	// Add target path
	if len(files) > 0 && files[0] != "." {
		args = append(args, "--targets", strings.Join(files, ","))
	} else {
		args = append(args, ".")
	}

	// Run semgrep with rules from rules path if available
	if g.rulesPath != "" {
		rulesPath := findSemgrepRules(g.rulesPath)
		if rulesPath != "" {
			args = append(args, "--config", rulesPath)
		}
	}

	cmd := exec.CommandContext(ctx, "semgrep", args...)
	output, err := cmd.Output()

	// semgrep returns non-zero on findings
	if err != nil && len(output) == 0 {
		return nil, fmt.Errorf("semgrep execution failed: %w", err)
	}

	// Parse semgrep JSON output
	if len(output) > 0 {
		var semgrepOutput SemgrepOutput
		if err := json.Unmarshal(output, &semgrepOutput); err != nil {
			logger.Warn("Failed to parse semgrep output", "error", err)
			return result, nil
		}

		for _, result2 := range semgrepOutput.Results {
			severity := mapSemgrepSeverity(result2.Extra.Severity)
			issue := Issue{
				File:     result2.Path,
				Line:     result2.Start.Line,
				Message:  result2.Extra.Message,
				Severity: severity,
				Tool:     "semgrep",
				RuleID:   result2.CheckID,
			}

			// Try to extract line number from different format
			if issue.Line == 0 && result2.Start.Line > 0 {
				issue.Line = result2.Start.Line
			}

			result.Issues = append(result.Issues, issue)

			if severity != "info" && severity != "low" {
				result.Passed = false
				if severity == "critical" {
					result.Severity = "critical"
				} else if severity == "high" && result.Severity != "critical" {
					result.Severity = "high"
				} else if severity == "medium" && result.Severity == "none" {
					result.Severity = "medium"
				}
			}
		}
	}

	return result, nil
}

// GosecOutput represents gosec JSON output format
type GosecOutput struct {
	Issues []GosecIssue `json:"Issues"`
}

// GosecIssue represents a gosec finding
type GosecIssue struct {
	RuleID   string `json:"rule_id"`
	File     string `json:"file"`
	Line     string `json:"line"`
	Details  string `json:"details"`
	Severity string `json:"severity"`
}

// SemgrepOutput represents semgrep JSON output format
type SemgrepOutput struct {
	Results []SemgrepResult `json:"results"`
}

// SemgrepResult represents a semgrep finding
type SemgrepResult struct {
	CheckID string          `json:"check_id"`
	Path    string          `json:"path"`
	Start   SemgrepPosition `json:"start"`
	End     SemgrepPosition `json:"end"`
	Extra   SemgrepExtra    `json:"extra"`
}

// SemgrepPosition represents a position in the code
type SemgrepPosition struct {
	Line int `json:"line"`
}

// SemgrepExtra represents extra semgrep result info
type SemgrepExtra struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// isToolInstalled checks if a tool is available
func isToolInstalled(name string) bool {
	cmd := exec.Command(name, "version")
	err := cmd.Run()
	return err == nil
}

// mapSeverity maps gosec severity to standard severity
func mapSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium", "warning":
		return "medium"
	case "low":
		return "low"
	default:
		return "info"
	}
}

// mapSemgrepSeverity maps semgrep severity to standard severity
func mapSemgrepSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "info":
		return "low"
	default:
		return "info"
	}
}

// findSemgrepRules finds semgrep rules directory in the rules path
func findSemgrepRules(rulesPath string) string {
	semgrepPath := strings.ReplaceAll(rulesPath, "~", os.Getenv("HOME"))
	semgrepPath = strings.ReplaceAll(semgrepPath, "$HOME", os.Getenv("HOME"))

	// Check for semgrep directory
	semgrepDir := fmt.Sprintf("%s/semgrep", semgrepPath)
	if _, err := os.Stat(semgrepDir); err == nil {
		return semgrepDir
	}

	// Check for .semgrep directory
	semgrepDir = fmt.Sprintf("%s/.semgrep", semgrepPath)
	if _, err := os.Stat(semgrepDir); err == nil {
		return semgrepDir
	}

	return ""
}

// runToolWithTimeout runs a SAST tool with a timeout
func runToolWithTimeout(ctx context.Context, name string, args []string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// parseGosecOutput parses gosec text output into issues
func parseGosecOutput(output string) []Issue {
	var issues []Issue

	// Gosec text output format: /path/to/file:line:col: issue details
	// Example: /path/to/file:42:5: [CWE-78] command injection
	re := regexp.MustCompile(`(.+):(\d+):\d+:\s+\[(.+?)\]\s+(.+)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 5 {
			issue := Issue{
				File:     matches[1],
				Line:     parseInt(matches[2]),
				Severity: mapSeverity(matches[3]),
				Message:  matches[4],
				Tool:     "gosec",
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// parseInt safely parses a string to int
func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

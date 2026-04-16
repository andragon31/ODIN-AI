// Package guardian provides the Heimdall security guardian for ODIN
package guardian

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/pkg/logger"
)

// Guardian is the main security guardian struct
type Guardian struct {
	rulesPath    string
	blockOnCrit  bool
	saastTools   []string
	policyEngine PolicyEngine
}

// PolicyEngine interface for OPA policy evaluation
type PolicyEngine interface {
	LoadPolicies(path string) error
	Evaluate(ctx context.Context, input PolicyInput) ([]PolicyResult, error)
}

// PolicyInput is the input for policy evaluation
type PolicyInput struct {
	Files     []string
	Code      string
	Language  string
	CommitMsg string
	Diff      string
}

// PolicyResult represents a policy evaluation result
type PolicyResult struct {
	PolicyID   string
	PolicyName string
	Severity   string
	Message    string
	Passed     bool
}

// CheckResult contains the analysis results
type CheckResult struct {
	Passed     bool
	Severity   string // critical, high, medium, low, none
	Issues     []Issue
	Duration   time.Duration
	ReportPath string
	Tool       string // Which tool found the issue
}

// Issue represents a security finding
type Issue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Tool     string `json:"tool"` // gosec, semgrep, opa
	RuleID   string `json:"rule_id,omitempty"`
}

// NewGuardian creates a new Guardian instance
func NewGuardian(cfg *config.GuardianConfig) (*Guardian, error) {
	g := &Guardian{
		rulesPath:   cfg.RulesPath,
		blockOnCrit: cfg.BlockOnCrit,
		saastTools:  cfg.SAST.Tools,
	}

	// Initialize OPA policy engine
	g.policyEngine = NewOPAPolicyEngine()

	// Load policies if rules path exists
	if _, err := os.Stat(g.rulesPath); err == nil {
		if err := g.policyEngine.LoadPolicies(g.rulesPath); err != nil {
			logger.Warn("Failed to load policies", "error", err)
		}
	}

	return g, nil
}

// Check runs security analysis on the specified files or all staged files
func (g *Guardian) Check(ctx context.Context, files []string, stagedOnly bool) (*CheckResult, error) {
	start := time.Now()

	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
		Duration: time.Since(start),
	}

	// If no files specified and stagedOnly, get staged files
	if len(files) == 0 && stagedOnly {
		staged, err := getStagedFiles()
		if err != nil {
			logger.Warn("Failed to get staged files", "error", err)
		} else {
			files = staged
		}
	}

	// If still no files, use current directory
	if len(files) == 0 {
		files = []string{"."}
	}

	// Run SAST tools
	for _, tool := range g.saastTools {
		var toolResult *CheckResult
		var err error

		switch tool {
		case "gosec":
			toolResult, err = g.runGosec(ctx, files)
		case "semgrep":
			toolResult, err = g.runSemgrep(ctx, files)
		default:
			continue
		}

		if err != nil {
			logger.Warn("SAST tool failed", "tool", tool, "error", err)
			continue
		}

		if toolResult != nil {
			result.Issues = append(result.Issues, toolResult.Issues...)
			if !toolResult.Passed {
				result.Passed = false
				if toolResult.Severity == "critical" {
					result.Severity = "critical"
				} else if toolResult.Severity == "high" && result.Severity != "critical" {
					result.Severity = "high"
				}
			}
		}
	}

	// Run OPA policy evaluation
	if len(files) > 0 {
		policyResults, err := g.policyEngine.Evaluate(ctx, PolicyInput{
			Files: files,
		})
		if err != nil {
			logger.Warn("Policy evaluation failed", "error", err)
		} else {
			for _, pr := range policyResults {
				if !pr.Passed {
					result.Passed = false
					result.Issues = append(result.Issues, Issue{
						Message:  pr.Message,
						Severity: pr.Severity,
						Tool:     "opa",
						RuleID:   pr.PolicyID,
					})
					if pr.Severity == "critical" {
						result.Severity = "critical"
					}
				}
			}
		}
	}

	result.Duration = time.Since(start)

	// Block if critical issues found and blockOnCrit is enabled
	if !result.Passed && g.blockOnCrit && result.Severity == "critical" {
		result.Passed = false
	}

	return result, nil
}

// CheckWithDiff runs security analysis on a git diff
func (g *Guardian) CheckWithDiff(ctx context.Context, diff string) (*CheckResult, error) {
	start := time.Now()

	result := &CheckResult{
		Passed:   true,
		Severity: "none",
		Issues:   []Issue{},
		Duration: time.Since(start),
	}

	// Evaluate diff against OPA policies
	policyResults, err := g.policyEngine.Evaluate(ctx, PolicyInput{
		Diff: diff,
	})
	if err != nil {
		logger.Warn("Policy evaluation failed", "error", err)
	} else {
		for _, pr := range policyResults {
			if !pr.Passed {
				result.Passed = false
				result.Issues = append(result.Issues, Issue{
					Message:  pr.Message,
					Severity: pr.Severity,
					Tool:     "opa",
					RuleID:   pr.PolicyID,
				})
				if pr.Severity == "critical" {
					result.Severity = "critical"
				}
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// IsInitialized returns true if rules directory exists
func (g *Guardian) IsInitialized() bool {
	_, err := os.Stat(g.rulesPath)
	return err == nil
}

// RulesPath returns the rules path
func (g *Guardian) RulesPath() string {
	return g.rulesPath
}

// getStagedFiles returns list of staged files from git
func getStagedFiles() ([]string, error) {
	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return nil, nil
	}

	// Run git diff --cached --name-only
	cmd := execCommandContext(context.Background(), "git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, nil
	}

	files := splitLines(string(output))
	var result []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			result = append(result, f)
		}
	}

	return result, nil
}

// splitLines splits a string by newlines
func splitLines(s string) []string {
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

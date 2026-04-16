package guardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/odin-ai/odin/pkg/logger"
)

// OPAPolicyEngine implements PolicyEngine using OPA
type OPAPolicyEngine struct {
	policies map[string]string // policy ID -> rego code
	loaded   bool
}

// NewOPAPolicyEngine creates a new OPA policy engine
func NewOPAPolicyEngine() *OPAPolicyEngine {
	return &OPAPolicyEngine{
		policies: make(map[string]string),
		loaded:   false,
	}
}

// LoadPolicies loads .rego files from the specified path
func (p *OPAPolicyEngine) LoadPolicies(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("rules path does not exist: %s", path)
	}

	p.policies = make(map[string]string)

	// Find all .rego files
	err := filepath.Walk(path, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(fullPath, ".rego") {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				logger.Warn("Failed to read policy file", "path", fullPath, "error", err)
				return nil
			}

			// Use filename (without extension) as policy ID
			relPath, _ := filepath.Rel(path, fullPath)
			policyID := strings.TrimSuffix(relPath, ".rego")
			policyID = strings.ReplaceAll(policyID, string(filepath.Separator), ".")

			p.policies[policyID] = string(data)
			logger.Info("Loaded policy", "id", policyID, "path", fullPath)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk rules path: %w", err)
	}

	p.loaded = true
	return nil
}

// Evaluate evaluates policies against the input
func (p *OPAPolicyEngine) Evaluate(ctx context.Context, input PolicyInput) ([]PolicyResult, error) {
	var results []PolicyResult

	// If no policies loaded, return empty results
	if !p.loaded || len(p.policies) == 0 {
		// Return default pass for built-in checks
		return results, nil
	}

	// For each loaded policy, we simulate evaluation
	// In a full implementation, this would call OPA via subprocess or SDK
	for policyID, rego := range p.policies {
		result := p.evaluateRegoPolicy(policyID, rego, input)
		if result != nil {
			results = append(results, *result)
		}
	}

	return results, nil
}

// evaluateRegoPolicy evaluates a single Rego policy
// In production, this would use the OPA SDK or call `opa eval`
func (p *OPAPolicyEngine) evaluateRegoPolicy(policyID, rego string, input PolicyInput) *PolicyResult {
	// Skip empty policies
	if strings.TrimSpace(rego) == "" {
		return nil
	}

	// Extract policy name from ID
	parts := strings.Split(policyID, ".")
	policyName := parts[len(parts)-1]

	// Default: policy passes
	result := &PolicyResult{
		PolicyID:   policyID,
		PolicyName: policyName,
		Passed:     true,
		Severity:   "none",
		Message:    "",
	}

	// Simple pattern-based evaluation for common security rules
	// In production, use OPA SDK for proper Rego evaluation

	// Check for hardcoded secrets patterns
	if strings.Contains(rego, "hardcoded_secret") || strings.Contains(rego, " Hardcoded ") {
		// Check if input files contain potential secrets
		for _, file := range input.Files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Simple pattern matching for common secrets
			secretPatterns := []string{
				"password=",
				"api_key=",
				"secret=",
				"AWS_ACCESS_KEY",
				"PRIVATE_KEY",
			}

			for _, pattern := range secretPatterns {
				if strings.Contains(string(content), pattern) && !strings.Contains(string(content), "// "+pattern) {
					result.Passed = false
					result.Severity = "critical"
					result.Message = fmt.Sprintf("Potential hardcoded secret found: %s", pattern)
					return result
				}
			}
		}
	}

	// Check for SQL injection patterns
	if strings.Contains(rego, "sql_injection") || strings.Contains(rego, "SQL") {
		for _, file := range input.Files {
			if !strings.HasSuffix(file, ".go") && !strings.HasSuffix(file, ".py") {
				continue
			}

			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Check for SQL construction with string concatenation
			if strings.Contains(string(content), "\"SELECT\"") && strings.Contains(string(content), "+") {
				result.Passed = false
				result.Severity = "high"
				result.Message = "Potential SQL injection: string concatenation in SQL query"
				return result
			}
		}
	}

	return result
}

// checkOPAInstalled checks if opa binary is available
func checkOPAInstalled() bool {
	cmd := exec.Command("opa", "version")
	err := cmd.Run()
	return err == nil
}

// execOPA eval runs OPA with the given input
func execOPAeval(ctx context.Context, input PolicyInput, policyPath string) ([]PolicyResult, error) {
	var results []PolicyResult

	// Prepare input JSON
	inputJSON, err := json.Marshal(map[string]interface{}{
		"files": input.Files,
		"diff":  input.Diff,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Run OPA eval
	cmd := exec.CommandContext(ctx, "opa", "eval", "--format", "json", "--bundle", policyPath, "data")
	cmd.Stdin = bytes.NewReader(inputJSON)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("opa eval failed: %w", err)
	}

	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse OPA output: %w", err)
	}

	return results, nil
}

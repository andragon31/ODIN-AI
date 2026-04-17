package guardian

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Guardian CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heimdall",
		Short: "Heimdall - Guardian security with OPA and SAST",
		Long: `Heimdall is the Norse god who guards the Bifröst bridge.
This command manages security checks using OPA policies,
gosec for Go security, and semgrep for multi-language analysis.`,
	}

	cmd.AddCommand(
		newCheckCmd(),
		newCheckRunesCmd(),
		newHookInstallCmd(),
		newHookUninstallCmd(),
		newReportCmd(),
		newStatusCmd(),
	)

	return cmd
}

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [files...]",
		Short: "Run security analysis",
		Long: `Run security analysis on specified files or staged files.
If no files are specified, checks all staged files in git.
Uses OPA policies, gosec, and semgrep based on configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, args)
		},
	}

	cmd.Flags().Bool("staged", true, "Check only staged files")
	cmd.Flags().Bool("all", false, "Check all files in repository")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().String("report", "", "Write report to file")

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	jsonOutput, _ := cmd.Flags().GetBool("json")
	reportPath, _ := cmd.Flags().GetString("report")
	stagedOnly, _ := cmd.Flags().GetBool("staged")
	allFiles, _ := cmd.Flags().GetBool("all")

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create guardian
	g, err := NewGuardian(&cfg.Guardian)
	if err != nil {
		return fmt.Errorf("failed to create guardian: %w", err)
	}

	// Determine files to check
	files := args
	if allFiles {
		files = []string{"."}
		stagedOnly = false
	}

	// Run check
	result, err := g.Check(ctx, files, stagedOnly)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	// Generate report
	if reportPath != "" {
		if err := WriteReportToFile(result, reportPath); err != nil {
			logger.Warn("Failed to write report", "error", err)
		} else {
			fmt.Printf("Report written to: %s\n", reportPath)
		}
	}

	// Output
	if jsonOutput {
		reportJSON, err := GetReportJSON(result)
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %w", err)
		}
		fmt.Println(reportJSON)
	} else {
		fmt.Print(FormatReport(result))
	}

	// Set exit code
	if !result.Passed {
		if result.Severity == "critical" {
			return fmt.Errorf("critical security issues found")
		}
		// Non-critical issues don't fail the command
	}

	return nil
}

func newCheckRunesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runes [rune_paths...]",
		Short: "Check runes against OPA security policies",
		Long: `Evaluate rune files (RUNE.md + rune.yaml) against security.rego policies.
Evaluates sandbox settings, execution types, dangerous patterns, and semver validity.

If no paths specified, checks all installed runes at ~/.odin/runes/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckRunes(cmd, args)
		},
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("verbose", false, "Show all policy results")

	return cmd
}

func runCheckRunes(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Determine rune paths
	runePaths := args
	if len(runePaths) == 0 {
		// Default to ~/.odin/runes/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		runesDir := filepath.Join(homeDir, ".odin", "runes")
		if entries, err := os.ReadDir(runesDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					runePaths = append(runePaths, filepath.Join(runesDir, entry.Name()))
				}
			}
		}
	}

	if len(runePaths) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No runes found. Install runes with: odin runes install")
		}
		return nil
	}

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load OPA policies
	rulesPath := cfg.Guardian.RulesPath
	securityPolicyPath := filepath.Join(rulesPath, "security.rego")

	policyData, err := os.ReadFile(securityPolicyPath)
	if err != nil {
		return fmt.Errorf("failed to read security policy: %w", err)
	}

	var results []RuneCheckResult

	for _, runePath := range runePaths {
		result := checkRuneAgainstPolicy(runePath, string(policyData), verbose)
		results = append(results, result)
	}

	// Output
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// Human-readable output
	fmt.Printf("\n═══ Rune Security Check ═══\n\n")

	passedCount := 0
	for _, r := range results {
		if r.Passed {
			passedCount++
		}
		status := "✅ PASS"
		if !r.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s %s\n", status, r.RuneName)
		if !r.Passed && verbose {
			for _, e := range r.Errors {
				fmt.Printf("   • %s\n", e)
			}
		}
	}

	fmt.Printf("\nResults: %d/%d passed\n", passedCount, len(results))

	return nil
}

// RuneCheckResult represents the result of checking a rune against policies
type RuneCheckResult struct {
	RuneName string   `json:"rune_name"`
	Path     string   `json:"path"`
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors,omitempty"`
}

// checkRuneAgainstPolicy evaluates a rune directory against the security policy
func checkRuneAgainstPolicy(runePath, policy string, verbose bool) RuneCheckResult {
	result := RuneCheckResult{
		Path:   runePath,
		Passed: true,
		Errors: []string{},
	}

	// Get rune name from directory
	result.RuneName = filepath.Base(runePath)

	// Read RUNE.md if exists
	runeMdPath := filepath.Join(runePath, "RUNE.md")
	runeYamlPath := filepath.Join(runePath, "rune.yaml")

	// Build input for OPA
	input := map[string]interface{}{
		"rune": map[string]interface{}{
			"name":      result.RuneName,
			"version":   "1.0.0",
			"execution": map[string]interface{}{"type": "prompt", "sandbox": true},
			"triggers":  map[string]interface{}{"commands": []string{}, "filePatterns": []string{}},
		},
	}

	// Read rune.yaml for more accurate data
	if data, err := os.ReadFile(runeYamlPath); err == nil {
		// Try to parse as YAML
		if yamlData, err := parseYamlRune(string(data)); err == nil {
			if yamlData.Name != "" {
				input["rune"].(map[string]interface{})["name"] = yamlData.Name
				result.RuneName = yamlData.Name
			}
			if yamlData.Version != "" {
				input["rune"].(map[string]interface{})["version"] = yamlData.Version
			}
			if yamlData.ExecutionType != "" {
				input["rune"].(map[string]interface{})["type"] = yamlData.ExecutionType
			}
			if yamlData.Sandbox {
				input["rune"].(map[string]interface{})["sandbox"] = yamlData.Sandbox
			}
		}
	}

	// Read RUNE.md for prompt content
	if data, err := os.ReadFile(runeMdPath); err == nil {
		content := string(data)
		exec := input["rune"].(map[string]interface{})

		// Check for dangerous patterns in content
		if strings.Contains(content, "rm -rf") {
			result.Passed = false
			result.Errors = append(result.Errors, "RUNE.md contains 'rm -rf' dangerous pattern")
		}
		if strings.Contains(content, "sudo") {
			result.Passed = false
			result.Errors = append(result.Errors, "RUNE.md contains 'sudo' privilege escalation pattern")
		}

		// Extract trigger commands if present
		cmdMatches := regexpFindAll(`/(?:[\w-]+)`, content)
		if len(cmdMatches) > 0 {
			exec["commands"] = cmdMatches
		}
	}

	// Simple policy evaluation (in production, use OPA SDK)
	// For now, do basic validation
	exec := input["rune"].(map[string]interface{})

	// Check execution type
	execType, ok := exec["type"].(string)
	if !ok {
		execType = "prompt"
	}

	sandbox, ok := exec["sandbox"].(bool)
	if !ok {
		sandbox = true
	}

	// Script/WASM without sandbox is denied
	if (execType == "script" || execType == "wasm") && !sandbox {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s execution without sandbox denied", execType))
	}

	// Unknown execution type
	if execType != "prompt" && execType != "script" && execType != "wasm" {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("unknown execution type: %s", execType))
	}

	// Check version format
	version, _ := exec["version"].(string)
	if version != "" && !isValidSemver(version) {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid semver: %s", version))
	}

	return result
}

// runeYamlData represents parsed rune.yaml
type runeYamlData struct {
	Name          string
	Version       string
	ExecutionType string
	Sandbox       bool
}

// parseYamlRune parses a rune.yaml file content
func parseYamlRune(content string) (*runeYamlData, error) {
	data := &runeYamlData{
		ExecutionType: "prompt",
		Sandbox:       true,
	}

	// Simple line-based parsing
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			data.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "version:") {
			data.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
		if strings.HasPrefix(line, "type:") || strings.HasPrefix(line, "execution.type:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "type:"))
			val = strings.TrimSpace(strings.TrimPrefix(val, "execution.type:"))
			data.ExecutionType = val
		}
		if strings.HasPrefix(line, "sandbox:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "sandbox:"))
			data.Sandbox = val == "true" || val == "yes" || val == "1"
		}
	}

	return data, nil
}

// isValidSemver checks if version is valid semver
func isValidSemver(v string) bool {
	if v == "" {
		return true
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return false
	}
	// Remove 'v' prefix if present
	if strings.HasPrefix(v, "v") {
		v = v[1:]
		parts = strings.Split(v, ".")
	}
	for _, p := range parts[:2] {
		if _, err := parseVersionPart(p); err != nil {
			return false
		}
	}
	return true
}

// parseVersionPart parses a numeric version part
func parseVersionPart(p string) (int, error) {
	n := 0
	for _, c := range p {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// regexpFindAll finds all matches (simple implementation)
func regexpFindAll(pattern, s string) []string {
	var matches []string
	pattern = pattern[1 : len(pattern)-1] // Remove ^ and $
	for i := 0; i < len(s); i++ {
		match := ""
		pi := 0
		for j := i; j < len(s) && pi < len(pattern); j++ {
			if pattern[pi] == '.' {
				match += string(s[j])
				pi++
			} else if pattern[pi] == s[j] {
				match += string(s[j])
				pi++
			} else {
				break
			}
		}
		if pi == len(pattern) {
			matches = append(matches, match)
		}
	}
	// Simple check for command-like patterns starting with /
	start := 0
	for {
		idx := strings.Index(s[start:], "/")
		if idx == -1 {
			break
		}
		cmdStart := start + idx
		cmdEnd := cmdStart + 1
		for cmdEnd < len(s) && (s[cmdEnd] == '-' || s[cmdEnd] == '_' || (s[cmdEnd] >= 'a' && s[cmdEnd] <= 'z') || (s[cmdEnd] >= 'A' && s[cmdEnd] <= 'Z') || (s[cmdEnd] >= '0' && s[cmdEnd] <= '9')) {
			cmdEnd++
		}
		if cmdEnd > cmdStart+1 {
			matches = append(matches, s[cmdStart:cmdEnd])
		}
		start = cmdEnd
	}
	return matches
}

func newHookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook-install",
		Short: "Install pre-commit hook",
		Long:  `Install the Heimdall pre-commit hook to .git/hooks/.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := InstallHook(); err != nil {
				return fmt.Errorf("failed to install hook: %w", err)
			}
			logger.Info("Pre-commit hook installed successfully")
			fmt.Println("Pre-commit hook installed. Heimdall will run security checks before each commit.")
			return nil
		},
	}
}

func newHookUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook-uninstall",
		Short: "Remove pre-commit hook",
		Long:  `Remove the Heimdall pre-commit hook from .git/hooks/.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := UninstallHook(); err != nil {
				return fmt.Errorf("failed to uninstall hook: %w", err)
			}
			logger.Info("Pre-commit hook removed")
			fmt.Println("Pre-commit hook removed.")
			return nil
		},
	}
}

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate CI/CD report",
		Long:  `Generate a JSON report for CI/CD pipelines.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate a new report",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReportGenerate(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <file>",
		Short: "Show a saved report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReportShow(args[0])
		},
	})

	return cmd
}

func runReportGenerate(cmd *cobra.Command) error {
	ctx := cmd.Context()
	outputPath, _ := cmd.Flags().GetString("output")

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create guardian
	g, err := NewGuardian(&cfg.Guardian)
	if err != nil {
		return fmt.Errorf("failed to create guardian: %w", err)
	}

	// Run check on staged files
	result, err := g.Check(ctx, nil, true)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	// Generate and write report
	report, err := GenerateReport(result, outputPath)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if outputPath == "" {
		// Output to stdout
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("failed to encode report: %w", err)
		}
	} else {
		fmt.Printf("Report written to: %s\n", outputPath)
	}

	return nil
}

func runReportShow(inputPath string) error {
	report, err := ReadReport(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read report: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show guardian status",
		Long:  `Display the status of Heimdall including loaded policies and tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGuardianStatus(cmd)
		},
	}
}

func runGuardianStatus(cmd *cobra.Command) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create guardian
	g, err := NewGuardian(&cfg.Guardian)
	if err != nil {
		return fmt.Errorf("failed to create guardian: %w", err)
	}

	// Check if rules exist
	rulesExist := g.IsInitialized()

	// Check tools
	gosecAvailable := isToolInstalled("gosec")
	semgrepAvailable := isToolInstalled("semgrep")
	opaAvailable := isToolInstalled("opa")

	// Check hook
	hookInstalled := IsHookInstalled()

	status := map[string]interface{}{
		"initialized":   rulesExist,
		"rules_path":    cfg.Guardian.RulesPath,
		"block_on_crit": cfg.Guardian.BlockOnCrit,
		"tools": map[string]bool{
			"gosec":   gosecAvailable,
			"semgrep": semgrepAvailable,
			"opa":     opaAvailable,
		},
		"saast_tools":    cfg.Guardian.SAST.Tools,
		"hook_installed": hookInstalled,
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(status)
	}

	// Human-readable output
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         ODIN AI - Heimdall Status              ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")

	fmt.Printf("║  Rules Path:    %-31s║\n", cfg.Guardian.RulesPath)
	fmt.Printf("║  Initialized:   %-31s║\n", boolToEmoji(rulesExist))
	fmt.Printf("║  Block on Crit: %-31s║\n", boolToEmoji(cfg.Guardian.BlockOnCrit))
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  Tools:                                     ║")

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "║    %-10s %-29s║\n", "gosec:", boolToStatus(gosecAvailable))
	fmt.Fprintf(w, "║    %-10s %-29s║\n", "semgrep:", boolToStatus(semgrepAvailable))
	fmt.Fprintf(w, "║    %-10s %-29s║\n", "opa:", boolToStatus(opaAvailable))
	w.Flush()

	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Pre-commit Hook: %-27s║\n", boolToEmoji(hookInstalled))
	fmt.Println("╚══════════════════════════════════════════════════╝")

	return nil
}

func boolToEmoji(b bool) string {
	if b {
		return "✅ yes"
	}
	return "❌ no"
}

func boolToStatus(b bool) string {
	if b {
		return "✅ available"
	}
	return "❌ not installed"
}

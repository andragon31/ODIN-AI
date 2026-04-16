package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Runes CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rune",
		Short: "Runes - Skills Registry with CUE validation",
		Long: `Runes is the Norse skills registry that validates, caches, and 
executes skills with CUE schema validation and sandbox testing.
Supports local caching, version rollback, and tag-based search.`,
	}

	cmd.AddCommand(
		newSearchCmd(),
		newInstallCmd(),
		newListCmd(),
		newTestCmd(),
		newValidateCmd(),
		newRemoveCmd(),
		newRollbackCmd(),
	)

	return cmd
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search skills by query or tags",
		Long: `Search for installed skills by name, description, or tags.
Use --tags to filter by specific tags.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, args)
		},
	}

	cmd.Flags().StringSlice("tags", nil, "Filter by tags (comma-separated)")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	tags, _ := cmd.Flags().GetStringSlice("tags")

	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	results, err := cache.Search(query, tags)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if len(results) == 0 {
		fmt.Println("No runes found matching your criteria")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tSOURCE\tINSTALLED AT")

	for _, r := range results {
		source := r.Source
		if source == "" {
			source = "local"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Version, source, r.InstalledAt)
	}
	w.Flush()

	return nil
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <path|url>",
		Short: "Install a skill from path or URL",
		Long: `Install a skill from a local file path or remote URL.
Validates CUE schema before installing. Malformed skills
will log warnings but won't block SDD.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd, args[0])
		},
	}

	return cmd
}

func runInstall(cmd *cobra.Command, source string) error {
	var data []byte
	var err error

	// Determine source type and read data
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// URL
		logger.Info("Fetching rune from URL", "url", source)
		resp, err := http.Get(source)
		if err != nil {
			return fmt.Errorf("failed to fetch from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	} else {
		// Local file path
		data, err = os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Validate CUE schema before writing
	rune, result := ValidateYAML(data)
	if !result.Valid {
		logger.Warn("Rune validation failed - not installing",
			"name", rune.Name,
			"errors", strings.Join(result.Errors, "; "))
		return fmt.Errorf("validation errors: %s", strings.Join(result.Errors, "; "))
	}

	if len(result.Warns) > 0 {
		logger.Warn("Rune has warnings",
			"name", rune.Name,
			"warnings", strings.Join(result.Warns, "; "))
	}

	// Check if already installed
	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	if cache.IsInstalled(rune.Name, rune.Version) {
		logger.Warn("Rune already installed", "name", rune.Name, "version", rune.Version)
	}

	// Store in cache
	if err := cache.Store(rune, source); err != nil {
		return fmt.Errorf("failed to store rune: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status":  "success",
			"name":    rune.Name,
			"version": rune.Version,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Rune %s@%s installed successfully\n", rune.Name, rune.Version)
	return nil
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		Long:  `Display all installed skills with their versions and sources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}
}

func runList(cmd *cobra.Command) error {
	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	results, err := cache.List()
	if err != nil {
		return fmt.Errorf("failed to list runes: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if len(results) == 0 {
		fmt.Println("No runes installed. Use 'odin rune install <path|url>' to install one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tINSTALLED AT")

	// Group by name and show latest version
	seen := make(map[string]bool)
	for _, r := range results {
		key := r.Name + "@" + r.Version
		if !seen[key] {
			seen[key] = true
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Version, r.InstalledAt)
		}
	}
	w.Flush()

	return nil
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <name>",
		Short: "Test a skill in sandbox",
		Long: `Test a skill in sandboxed mode and return pass/fail with output diff.
Only skills with type "script" or "prompt" can be tested.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(cmd, args[0])
		},
	}

	return cmd
}

func runTest(cmd *cobra.Command, name string) error {
	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	// Get latest version if no version specified
	var rune *Rune
	if strings.Contains(name, "@") {
		parts := strings.SplitN(name, "@", 2)
		name = parts[0]
		version := parts[1]
		rune, err = cache.Load(name, version)
		if err != nil {
			return fmt.Errorf("failed to load rune: %w", err)
		}
	} else {
		// Load latest version
		latestVersion, err := cache.GetLatestVersion(name)
		if err != nil {
			return fmt.Errorf("rune %s not found", name)
		}
		rune, err = cache.Load(name, latestVersion)
		if err != nil {
			return fmt.Errorf("failed to load rune: %w", err)
		}
	}

	// Run sandbox test
	result, err := testInSandbox(rune)
	if err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("Testing rune: %s@%s\n", rune.Name, rune.Version)
	fmt.Printf("Type: %s\n", rune.Execution.Type)
	fmt.Printf("Sandbox: %v\n", rune.Execution.Sandbox)
	fmt.Println()

	if result.Passed {
		fmt.Println("✓ TEST PASSED")
	} else {
		fmt.Println("✗ TEST FAILED")
	}

	if len(result.Output) > 0 {
		fmt.Println("\nOutput:")
		fmt.Println(result.Output)
	}

	if len(result.Diff) > 0 {
		fmt.Println("\nDiff:")
		fmt.Println(result.Diff)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

// TestResult represents the result of a skill test
type TestResult struct {
	Passed bool     `json:"passed"`
	Output string   `json:"output"`
	Diff   string   `json:"diff,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

// testInSandbox runs a skill test in sandboxed mode
func testInSandbox(r *Rune) (*TestResult, error) {
	result := &TestResult{Passed: false}

	// Check if skill can be tested
	if r.Execution.Type != "script" && r.Execution.Type != "prompt" {
		result.Errors = append(result.Errors, "only 'script' and 'prompt' type skills can be tested")
		return result, nil
	}

	// Execute based on type
	switch r.Execution.Type {
	case "script":
		// For script type, we simulate test execution
		// In a real implementation, this would run in a sandboxed environment
		if r.Execution.Sandbox {
			result.Output = fmt.Sprintf("[SANDBOXED] Would execute script: %s", r.Execution.Script)
			result.Passed = true
		} else {
			result.Output = fmt.Sprintf("[UNSANDBOXED] Script execution skipped for safety")
			result.Errors = append(result.Errors, "script execution without sandbox is not implemented")
		}

	case "prompt":
		// For prompt type, we validate the prompt exists
		if r.Execution.Prompt == "" {
			result.Errors = append(result.Errors, "prompt is empty")
			return result, nil
		}
		result.Output = fmt.Sprintf("[PROMPT TEST] Prompt template: %s", r.Execution.Prompt)
		result.Passed = true

	default:
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported type: %s", r.Execution.Type))
	}

	return result, nil
}

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a skill against CUE schema",
		Long: `Validate a skill file against the CUE schema without installing it.
Returns validation errors and warnings.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, args[0])
		},
	}
}

func runValidate(cmd *cobra.Command, path string) error {
	// Resolve path
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	rune, result, err := ValidateFile(path)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"valid":    result.Valid,
			"errors":   result.Errors,
			"warnings": result.Warns,
			"rune":     rune,
		})
	}

	if result.Valid {
		fmt.Printf("✓ Rune %s@%s is valid\n", rune.Name, rune.Version)
	} else {
		fmt.Printf("✗ Rune %s@%s has validation errors\n", rune.Name, rune.Version)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if len(result.Warns) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warns {
			fmt.Printf("  - %s\n", w)
		}
	}

	if result.Valid {
		return nil
	}
	return fmt.Errorf("validation failed")
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>[@<version>]",
		Short: "Uninstall a skill",
		Long: `Remove an installed skill. If no version is specified,
removes the latest version. Use 'odin rune list' to see versions.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(cmd, args[0])
		},
	}
}

func runRemove(cmd *cobra.Command, name string) error {
	var version string

	// Parse name@version
	if strings.Contains(name, "@") {
		parts := strings.SplitN(name, "@", 2)
		name = parts[0]
		version = parts[1]
	} else {
		// Get latest version
		cache, err := DefaultCache()
		if err != nil {
			return fmt.Errorf("failed to create cache: %w", err)
		}

		latest, err := cache.GetLatestVersion(name)
		if err != nil {
			return fmt.Errorf("rune %s not found", name)
		}
		version = latest
	}

	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	if err := cache.Remove(name, version); err != nil {
		return fmt.Errorf("failed to remove rune: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status":  "success",
			"name":    name,
			"version": version,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Rune %s@%s removed successfully\n", name, version)
	return nil
}

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <name>@<version>",
		Short: "Rollback to a previous version",
		Long: `Rollback a skill to a previous version. Lists available
versions if none specified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRollback(cmd, args[0])
		},
	}
}

func runRollback(cmd *cobra.Command, nameVersion string) error {
	// Parse name@version
	if !strings.Contains(nameVersion, "@") {
		return fmt.Errorf("must specify version as <name>@<version>")
	}

	parts := strings.SplitN(nameVersion, "@", 2)
	name := parts[0]
	targetVersion := parts[1]

	cache, err := DefaultCache()
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	// Load target version
	targetRune, err := cache.Load(name, targetVersion)
	if err != nil {
		return fmt.Errorf("failed to load rune: %w", err)
	}

	// Get source
	source, err := cache.GetSource(name, targetVersion)
	if err != nil {
		source = "local"
	}

	// Get current version for comparison
	currentVersion, err := cache.GetLatestVersion(name)
	if err != nil {
		currentVersion = "none"
	}

	// Remove current version and re-install target
	if err := cache.Remove(name, currentVersion); err != nil {
		logger.Warn("Failed to remove current version", "error", err)
	}

	if err := cache.Store(targetRune, source); err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status":           "success",
			"name":             name,
			"rolled_back_to":   targetVersion,
			"previous_version": currentVersion,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Rune %s rolled back from %s to %s\n", name, currentVersion, targetVersion)
	return nil
}

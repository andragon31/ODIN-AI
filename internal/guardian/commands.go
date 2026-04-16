package guardian

import (
	"encoding/json"
	"fmt"
	"os"
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

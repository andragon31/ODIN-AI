// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Commands returns the Nornir CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nornir",
		Short: "Nornir - Verification Suite",
		Long: `Nornir is the verification suite for ODIN AI.
		
Provides:
- Flaky test detection with retry and consistency tracking
- Latency benchmarking for critical phases
- Multi-OS matrix testing
- JSON report generation`,
	}

	cmd.AddCommand(
		newRunCmd(),
		newFlakyCmd(),
		newBenchmarkCmd(),
		newReportCmd(),
		newMatrixCmd(),
	)

	return cmd
}

// newRunCmd creates the 'nornir run' command
func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the full verification suite",
		Long:  `Runs all verification tests including flaky detection, benchmarks, and matrix tests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := DefaultVerifyConfig()
			n := New(cfg)

			report, err := n.RunVerification()
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			// Output as JSON if requested
			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}

			// Console output
			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Nornir - Verification Suite           ║")
			fmt.Println("╠══════════════════════════════════════════════════╣")
			fmt.Printf("║  Platform: %-37s║\n", report.Platform)
			fmt.Printf("║  Duration: %-37s║\n", report.Duration.Round(time.Millisecond))
			fmt.Println("╠══════════════════════════════════════════════════╣")
			fmt.Printf("║  Total Tests:  %-35d║\n", report.Summary.TotalTests)
			fmt.Printf("║  Passed:       %-35d║\n", report.Summary.PassedTests)
			fmt.Printf("║  Failed:       %-35d║\n", report.Summary.FailedTests)
			fmt.Printf("║  Flaky:        %-35d║\n", report.Summary.FlakyTests)
			fmt.Printf("║  Pass Rate:    %-35.1f%%║\n", report.Summary.PassRate)
			fmt.Println("╚══════════════════════════════════════════════════╝")

			return nil
		},
	}
}

// newFlakyCmd creates the 'nornir flaky' command
func newFlakyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flaky",
		Short: "Detect flaky tests",
		Long:  `Analyzes test history to identify flaky tests based on consistency patterns.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := DefaultVerifyConfig()
			n := New(cfg)

			results := n.DetectFlakyTests()

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			if len(results) == 0 {
				fmt.Println("No flaky tests detected")
				return nil
			}

			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Flaky Test Detection                  ║")
			fmt.Println("╠══════════════════════════════════════════════════╣")

			for _, r := range results {
				status := "✓ STABLE"
				if r.IsFlaky {
					status = "⚠ FLAKY"
				}
				fmt.Printf("║  %s %s\n", status, r.TestName)
				fmt.Printf("║    Runs: %d | Pass: %d | Consistency: %.1f%%\n",
					r.RunCount, r.PassCount, r.Consistency*100)
				if len(r.FailedRuns) > 0 {
					fmt.Printf("║    Failed on runs: %v\n", r.FailedRuns)
				}
			}

			fmt.Println("╚══════════════════════════════════════════════════╝")

			return nil
		},
	}
}

// newBenchmarkCmd creates the 'nornir benchmark' command
func newBenchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "benchmark",
		Short: "Run latency benchmarks",
		Long:  `Benchmarks critical phases and compares against thresholds.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			n := New(DefaultVerifyConfig())
			results := n.RunBenchmarks()

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Latency Benchmarks                    ║")
			fmt.Println("╠══════════════════════════════════════════════════╣")

			for _, r := range results {
				status := "✓ PASS"
				if !r.Passed {
					status = "✗ FAIL"
				}
				fmt.Printf("║  %s %s (%.2fms)\n", status, r.Phase, r.LatencyMs)
				fmt.Printf("║    Threshold: %.2fms | Iterations: %d\n",
					r.ThresholdMs, r.Iterations)
				fmt.Printf("║    P50: %.2fms | P95: %.2fms | P99: %.2fms\n",
					r.P50LatencyMs, r.P95LatencyMs, r.P99LatencyMs)
			}

			fmt.Println("╚══════════════════════════════════════════════════╝")

			return nil
		},
	}
}

// newReportCmd creates the 'nornir report' command
func newReportCmd() *cobra.Command {
	var outputPath string
	var reportFormat string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate verification report",
		Long:  `Generates a comprehensive verification report in JSON format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			n := New(DefaultVerifyConfig())

			report, err := n.RunVerification()
			if err != nil {
				return fmt.Errorf("failed to run verification: %w", err)
			}

			rg := NewReportGenerator(reportFormat)

			if outputPath != "" {
				if err := rg.ExportJSON(report, outputPath); err != nil {
					return fmt.Errorf("failed to export report: %w", err)
				}
				fmt.Printf("Report saved to: %s\n", outputPath)
			} else {
				return rg.ExportJSONToWriter(report, os.Stdout)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&reportFormat, "format", "json", "Report format (only json supported)")
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}

// newMatrixCmd creates the 'nornir matrix' command
func newMatrixCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "matrix",
		Short: "Show matrix testing targets",
		Long:  `Displays the available multi-OS testing matrix targets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mr := NewMatrixRunner()

			if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(mr.GetTargets())
			}

			fmt.Println("╔══════════════════════════════════════════════════╗")
			fmt.Println("║           Matrix Testing Targets                 ║")
			fmt.Println("╠══════════════════════════════════════════════════╣")

			for i, t := range mr.GetTargets() {
				fmt.Printf("║  %d. %s/%s (%s)\n", i+1, t.OS, t.Arch, t.Version)
			}

			fmt.Println("╚══════════════════════════════════════════════════╝")

			return nil
		},
	}
}

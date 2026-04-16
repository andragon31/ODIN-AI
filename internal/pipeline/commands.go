// Package pipeline provides the installation pipeline with staged execution and rollback
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// Commands returns all pipeline CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline - Install with staged execution and rollback",
		Long: `Pipeline system for staged installations with automatic rollback.
Executes installation in stages: Detect → Backup → Install → Verify → Commit.
If any stage fails, automatic rollback restores the previous state.`,
	}

	cmd.AddCommand(
		newInstallCmd(),
		newStatusCmd(),
		newListStagesCmd(),
	)

	return cmd
}

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <component>",
		Short: "Install a component via pipeline",
		Long: `Install a component using the staged pipeline with automatic
backup and rollback. Ctrl+C during execution will cancel
the pipeline and roll back any completed stages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("component name required")
			}
			return runInstall(cmd, args[0])
		},
	}
}

func runInstall(cmd *cobra.Command, componentID string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Create pipeline
	p := NewPipeline(componentID)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Run pipeline in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Run()
	}()

	// Wait for completion or interrupt
	select {
	case err := <-errChan:
		if err != nil {
			if jsonOutput {
				data := map[string]interface{}{
					"status":      "failed",
					"component":   componentID,
					"error":       err.Error(),
					"backup_path": p.GetBackupPath(),
					"results":     p.GetResults(),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			return err
		}

		if jsonOutput {
			data := map[string]interface{}{
				"status":      "success",
				"component":   componentID,
				"backup_path": p.GetBackupPath(),
				"results":     p.GetResults(),
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}

		// Print stage results
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "STAGE\tSTATUS\tDURATION")
		for _, result := range p.GetResults() {
			status := "OK"
			if !result.Success {
				status = "FAILED"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", result.Stage, status, result.Duration)
		}
		w.Flush()

		fmt.Printf("\nInstallation completed: %s\n", componentID)
		return nil

	case <-sigChan:
		p.Cancel()
		return fmt.Errorf("installation cancelled")
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show pipeline status",
		Long: `Show current system detection information that
would be used during installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}
}

func runStatus(cmd *cobra.Command) error {
	p := NewPipeline("status-check")

	// Run just the detect stage
	detection := &SystemDetection{}
	p.detected = detection

	// Run detect manually
	result := p.executeStage(StageDetect)
	if !result.Success {
		return result.Error
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(detection)
	}

	fmt.Println("\n=== System Detection ===")
	fmt.Printf("OS:          %s\n", detection.OS)
	fmt.Printf("Arch:        %s\n", detection.Arch)
	fmt.Printf("Container:   %s\n", detection.Container)
	fmt.Printf("User:        %s\n", detection.User)
	fmt.Printf("Home:        %s\n", detection.HomeDir)
	fmt.Printf("Can Install: %v\n", detection.CanInstall)

	fmt.Println("\n=== Installed Agents ===")
	if len(detection.Agents) == 0 {
		fmt.Println("  None detected")
	} else {
		for _, agent := range detection.Agents {
			fmt.Printf("  - %s\n", agent)
		}
	}

	fmt.Println("\n=== Installed Components ===")
	if len(detection.Components) == 0 {
		fmt.Println("  None")
	} else {
		for _, comp := range detection.Components {
			fmt.Printf("  - %s\n", comp)
		}
	}

	fmt.Println("\n=== Installed Runes ===")
	if len(detection.Runes) == 0 {
		fmt.Println("  None")
	} else {
		for _, rune := range detection.Runes {
			fmt.Printf("  - %s\n", rune)
		}
	}

	return nil
}

func newListStagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stages",
		Short: "List pipeline stages",
		Long:  `List all stages in the pipeline execution order.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListStages(cmd)
		},
	}
}

func runListStages(cmd *cobra.Command) error {
	stages := []Stage{StageDetect, StageBackup, StageInstall, StageVerify, StageCommit}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ORDER\tSTAGE\tDESCRIPTION")

	descriptions := map[Stage]string{
		StageDetect:  "Detect OS, architecture, installed agents",
		StageBackup:  "Create backup at ~/.odin/backup/<timestamp>/",
		StageInstall: "Install runes, configs, hooks",
		StageVerify:  "Verify installation with Nornir",
		StageCommit:  "Commit or rollback based on verify result",
	}

	for i, stage := range stages {
		fmt.Fprintf(w, "%d\t%s\t%s\n", i+1, stage, descriptions[stage])
	}
	w.Flush()

	return nil
}

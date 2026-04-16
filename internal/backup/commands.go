// Package backup provides backup and restore functionality for ODIN
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all backup CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup - Create and restore ODIN backups",
		Long: `Backup system for ODIN components.
Creates timestamped backups of the .odin directory
and supports restoration from any backup point.`,
	}

	cmd.AddCommand(
		newCreateCmd(),
		newRestoreCmd(),
		newListCmd(),
	)

	return cmd
}

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [path]",
		Short: "Create a backup",
		Long: `Create a backup of the ODIN directory.
If path is not specified, backs up ~/.odin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getDefaultOdinPath()
			if len(args) > 0 {
				path = args[0]
			}
			return runCreate(cmd, path)
		},
	}
}

func runCreate(cmd *cobra.Command, path string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	logger.Info("Creating backup", "path", path)

	backupPath, err := CreateBackup(path)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	if jsonOutput {
		data := map[string]string{
			"status":      "success",
			"backup_path": backupPath,
			"source":      path,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Backup created at: %s\n", backupPath)
	return nil
}

func newRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [backup-path]",
		Short: "Restore from a backup",
		Long: `Restore ODIN from a backup.
Lists available backups if path is not specified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return runList(cmd)
			}
			return runRestore(cmd, args[0])
		},
	}
}

func runRestore(cmd *cobra.Command, backupPath string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	logger.Info("Restoring backup", "backup", backupPath)

	if err := RestoreBackup(backupPath); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	if jsonOutput {
		data := map[string]string{
			"status":        "success",
			"restored_from": backupPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Println("Restore completed successfully")
	return nil
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available backups",
		Long:  `List all available backups that can be restored.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd)
		},
	}
}

func runList(cmd *cobra.Command) error {
	backups, err := ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(backups)
	}

	if len(backups) == 0 {
		fmt.Println("No backups available")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tSIZE\tPATH")

	for _, b := range backups {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			b.Timestamp.Format("2006-01-02 15:04:05"),
			formatSize(b.Size),
			b.Path)
	}
	w.Flush()

	return nil
}

func getDefaultOdinPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".odin")
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

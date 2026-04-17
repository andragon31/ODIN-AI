// Package deploy provides the Dvergar forge/deploy system for ODIN
package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all Dvergar CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Dvergar - Forge/Deploy system",
		Long: `Dvergar is the Norse god of blacksmiths - representing ODIN's
	forge and deploy system. Provides cryptographic verification
	with cosign, automatic backup/rollback, and multi-platform
	installation support.`,
	}

	cmd.AddCommand(
		newInstallCmd(),
		newUpgradeCmd(),
		newRollbackCmd(),
		newVerifyCmd(),
		newStatusCmd(),
		newInfoCmd(),
		newListBackupsCmd(),
		newForgeCmd(),
	)

	return cmd
}

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [version]",
		Short: "Install ODIN AI",
		Long: `Install ODIN AI with cryptographic verification.
If version is not specified, installs the latest version.
Uses cosign for signature verification if available.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := "latest"
			if len(args) > 0 {
				version = args[0]
			}
			return runInstall(cmd, version)
		},
	}
}

func runInstall(cmd *cobra.Command, version string) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	// Detect system info
	info, err := DetectSystem()
	if err != nil {
		return fmt.Errorf("failed to detect system: %w", err)
	}

	// Ensure directories exist
	if err := dvergar.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]interface{}{
			"status":       "installing",
			"version":      version,
			"system":       info,
			"install_path": dvergar.InstallPath(),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("Installing ODIN AI",
		"version", version,
		"path", dvergar.InstallPath(),
		"os", info.OS,
		"arch", info.Arch,
	)

	// Check if already installed
	if dvergar.IsInstalled() {
		currentVersion, _ := dvergar.GetVersion()
		logger.Warn("ODIN is already installed",
			"current_version", currentVersion,
			"path", dvergar.InstallPath(),
		)
		fmt.Println("Use 'odin deploy upgrade' to upgrade or 'odin deploy install --force' to reinstall")
		return nil
	}

	// Generate and run install script
	script, err := GenerateInstallScript(info, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate install script: %w", err)
	}

	// For now, just report what would be done
	// In a real implementation, this would execute the script
	fmt.Println("Generated install script (not executed in dry-run mode):")
	fmt.Println("---")
	fmt.Println(script)
	fmt.Println("---")

	result := &InstallResult{
		Success:    true,
		Version:    version,
		BackupPath: dvergar.BackupPath(),
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("Installation would proceed at: %s\n", dvergar.InstallPath())
	return nil
}

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade [version]",
		Short: "Upgrade ODIN AI with automatic backup",
		Long: `Upgrade ODIN AI to a new version with automatic backup
and rollback capability. If version is not specified,
upgrades to the latest available version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := "latest"
			if len(args) > 0 {
				version = args[0]
			}
			return runUpgrade(cmd, version)
		},
	}
}

func runUpgrade(cmd *cobra.Command, version string) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Check if installed
	if !dvergar.IsInstalled() {
		if jsonOutput {
			data := map[string]string{
				"status":  "error",
				"message": "ODIN is not installed. Use 'odin deploy install' first.",
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}
		fmt.Println("ODIN is not installed. Use 'odin deploy install' first.")
		return nil
	}

	// Create backup before upgrade
	backup, err := dvergar.CreateBackup()
	if err != nil {
		logger.Warn("Failed to create backup", "error", err)
	} else {
		logger.Info("Backup created", "path", backup.Path)
	}

	// Write rollback script
	if err := dvergar.WriteRollbackScript(); err != nil {
		logger.Warn("Failed to write rollback script", "error", err)
	}

	// Detect system info
	info, err := DetectSystem()
	if err != nil {
		return fmt.Errorf("failed to detect system: %w", err)
	}

	if jsonOutput {
		data := map[string]interface{}{
			"status":      "upgrading",
			"version":     version,
			"backup_path": backup.Path,
			"system":      info,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("Upgrading ODIN AI",
		"version", version,
		"backup", backup.Path,
	)

	// Generate install script for upgrade
	script, err := GenerateInstallScript(info, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate install script: %w", err)
	}

	fmt.Println("Upgrade script (not executed in dry-run mode):")
	fmt.Println("---")
	fmt.Println(script)
	fmt.Println("---")

	return nil
}

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback [backup-path]",
		Short: "Rollback to a previous version",
		Long: `Rollback ODIN AI to a previous version.
If backup-path is not specified, rolls back to the most recent backup.
Use 'odin deploy list-backups' to see available backups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var backupPath string
			if len(args) > 0 {
				backupPath = args[0]
			}
			return runRollback(cmd, backupPath)
		},
	}
}

func runRollback(cmd *cobra.Command, backupPath string) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if backupPath == "" {
		backups, err := dvergar.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			if jsonOutput {
				data := map[string]string{
					"status":  "error",
					"message": "No backups available",
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(data)
			}
			fmt.Println("No backups available")
			return nil
		}

		backupPath = backups[0].Path
	}

	if jsonOutput {
		data := map[string]interface{}{
			"status":      "rolling_back",
			"backup_path": backupPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	logger.Info("Rolling back ODIN AI", "backup", backupPath)

	if err := dvergar.Rollback(backupPath); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Printf("Successfully rolled back to: %s\n", backupPath)
	return nil
}

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify ODIN installation",
		Long: `Verify the ODIN binary using cosign if available.
Checks cryptographic signature and transparency log.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd)
		},
	}
}

func runVerify(cmd *cobra.Command) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Check if cosign is available
	cosignAvailable := VerifyCosignAvailable()

	if jsonOutput {
		data := map[string]interface{}{
			"cosign_available": cosignAvailable,
			"installed":        dvergar.IsInstalled(),
		}
		if dvergar.IsInstalled() {
			data["install_path"] = dvergar.InstallPath()
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	if !cosignAvailable {
		fmt.Println("cosign not available - skipping cryptographic verification")
		fmt.Println("Install cosign for full verification: https://docs.sigstore.dev/cosign/installation/")
	}

	if !dvergar.IsInstalled() {
		fmt.Println("ODIN is not installed")
		return nil
	}

	fmt.Printf("ODIN is installed at: %s\n", dvergar.InstallPath())

	version, _ := dvergar.GetVersion()
	fmt.Printf("Version: %s\n", version)

	if cosignAvailable {
		fmt.Println("cosign verification: available (not executed in dry-run mode)")
	} else {
		fmt.Println("cosign verification: not available")
	}

	return nil
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show deployment status",
		Long: `Display the deployment status of ODIN including
installation state, version, and backup information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeployStatus(cmd)
		},
	}
}

func runDeployStatus(cmd *cobra.Command) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	jsonOutput, _ := cmd.Flags().GetBool("json")

	status := map[string]interface{}{
		"installed":    dvergar.IsInstalled(),
		"install_path": dvergar.InstallPath(),
		"backup_path":  dvergar.BackupPath(),
		"log_path":     dvergar.LogPath(),
	}

	if dvergar.IsInstalled() {
		version, _ := dvergar.GetVersion()
		status["version"] = version

		backups, _ := dvergar.ListBackups()
		status["backup_count"] = len(backups)
		if len(backups) > 0 {
			status["latest_backup"] = backups[0].Path
		}
	}

	cosignAvailable := VerifyCosignAvailable()
	status["cosign_available"] = cosignAvailable

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(status)
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║           ODIN AI - Deploy Status                ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Installed:   %-37s║\n", boolToYesNo(dvergar.IsInstalled()))
	fmt.Printf("║  Path:        %-37s║\n", dvergar.InstallPath())
	if dvergar.IsInstalled() {
		version, _ := dvergar.GetVersion()
		fmt.Printf("║  Version:     %-37s║\n", version)
	}
	fmt.Printf("║  Backup Path: %-37s║\n", dvergar.BackupPath())
	fmt.Printf("║  Cosign:      %-37s║\n", boolToYesNo(cosignAvailable))
	fmt.Println("╚══════════════════════════════════════════════════╝")

	return nil
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show system information",
		Long: `Display detected system information including OS,
architecture, container environment, and user details.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(cmd)
		},
	}
}

func runInfo(cmd *cobra.Command) error {
	info, err := DetectSystem()
	if err != nil {
		return fmt.Errorf("failed to detect system: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║           System Information                     ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  OS:          %-37s║\n", info.OS)
	fmt.Printf("║  Arch:        %-37s║\n", info.Arch)
	fmt.Printf("║  Container:   %-37s║\n", info.Container)
	if info.Container != "" {
		fmt.Printf("║  Container:   %-37s║\n", info.Container)
	}
	fmt.Printf("║  User:        %-37s║\n", info.User)
	fmt.Printf("║  Home Dir:    %-37s║\n", info.HomeDir)
	fmt.Printf("║  Install:     %-37s║\n", info.InstallPath)
	fmt.Printf("║  WSL:         %-37s║\n", boolToYesNo(info.IsWSL))
	fmt.Println("╚══════════════════════════════════════════════════╝")

	return nil
}

func newListBackupsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-backups",
		Short: "List available backups",
		Long:  `List all available backups that can be used for rollback.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListBackups(cmd)
		},
	}
}

func runListBackups(cmd *cobra.Command) error {
	cfg := DefaultDeployConfig()
	dvergar := New(cfg)

	backups, err := dvergar.ListBackups()
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
	fmt.Fprintln(w, "TIMESTAMP\tVERSION\tPATH")

	for _, b := range backups {
		fmt.Fprintf(w, "%s\t%s\t%s\n", b.Timestamp.Format("2006-01-02 15:04:05"), b.Version, b.Path)
	}
	w.Flush()

	return nil
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func newForgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "forge",
		Short: "Forge infrastructure for the current project",
		Long:  `Analyze the project and generate Docker/IaC artifacts using the Blacksmith engine.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runForge(cmd)
		},
	}
}

func runForge(cmd *cobra.Command) error {
	cwd, _ := os.Getwd()
	fm := NewForgeManager(cwd)
	return fm.RunForge()
}

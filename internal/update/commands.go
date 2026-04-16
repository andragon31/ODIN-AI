package update

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns all update CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Self-update ODIN to a newer version",
		Long: `Self-update ODIN by downloading the latest (or specified) version
from GitHub releases and replacing the current binary.

Supports verification using cosign signatures when a public key is provided.

Examples:
  odin self-update                    # Update to latest stable
  odin self-update --channel beta    # Update to beta channel
  odin self-update --check            # Check for updates without installing
  odin self-update --version 1.0.0   # Install specific version`,
	}

	cmd.AddCommand(
		newUpdateCmd(),
		newCheckCmd(),
	)

	return cmd
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update to the latest version",
		Long:  `Download and install the latest version of ODIN.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd)
		},
	}

	cmd.Flags().String("channel", "stable", "Update channel (stable|beta)")
	cmd.Flags().String("version", "", "Specific version to install")
	cmd.Flags().String("key", "", "Cosign public key for signature verification")
	cmd.Flags().Bool("force", false, "Force update even if version is same")

	return cmd
}

func runUpdate(cmd *cobra.Command) error {
	channel, _ := cmd.Flags().GetString("channel")
	version, _ := cmd.Flags().GetString("version")
	cosignKey, _ := cmd.Flags().GetString("key")
	force, _ := cmd.Flags().GetBool("force")

	cfg := DefaultConfig()
	cfg.Channel = channel

	if version != "" {
		cfg.CurrentVersion = version
	}

	currentVersion := GetCurrentVersion()

	// If version specified, download that directly
	if version != "" {
		return downloadAndInstall(cfg, version, cosignKey)
	}

	// Check for updates
	updateInfo, err := CheckForUpdate(cfg)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if updateInfo == nil {
		logger.Info("No update available", "current", currentVersion, "channel", channel)
		fmt.Printf("You are running the latest version (%s) on the %s channel.\n", currentVersion, channel)
		return nil
	}

	// Compare versions
	if !force && !IsNewer(updateInfo, currentVersion) {
		logger.Info("No update needed", "current", currentVersion, "available", updateInfo.Version)
		fmt.Printf("You are running the latest version (%s).\n", currentVersion)
		return nil
	}

	logger.Info("Update available", "current", currentVersion, "new", updateInfo.Version)
	fmt.Printf("Update available: %s (you have %s)\n", updateInfo.Version, currentVersion)

	return downloadAndInstall(cfg, updateInfo.Version, cosignKey)
}

func downloadAndInstall(cfg *Config, version string, cosignKey string) error {
	downloadURL := GetDownloadPath(cfg, version)

	// Download the update
	downloadedPath, err := DownloadUpdate(version, downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Verify signature if key provided
	if cosignKey != "" {
		valid, err := VerifySignature(downloadedPath, cosignKey)
		if err != nil {
			logger.Warn("Signature verification failed", "error", err)
			fmt.Printf("Warning: Signature verification failed: %v\n", err)
			// Continue anyway with user consent in production would ask
		}
		if !valid {
			os.Remove(downloadedPath)
			return fmt.Errorf("signature verification failed: invalid signature")
		}
		logger.Info("Signature verified successfully")
	}

	// Apply the update
	if err := ApplyUpdate(downloadedPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	logger.Info("Update applied successfully", "version", version)
	fmt.Printf("Successfully updated to version %s\n", version)
	fmt.Println("Restart ODIN to use the new version.")

	return nil
}

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check for available updates",
		Long:  `Check if a newer version is available without installing it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd)
		},
	}

	cmd.Flags().String("channel", "stable", "Channel to check (stable|beta)")
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}

func runCheck(cmd *cobra.Command) error {
	channel, _ := cmd.Flags().GetString("channel")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	cfg := DefaultConfig()
	cfg.Channel = channel

	currentVersion := GetCurrentVersion()

	updateInfo, err := CheckForUpdate(cfg)
	if err != nil {
		if jsonOutput {
			data := map[string]interface{}{
				"error":   err.Error(),
				"current": currentVersion,
				"channel": channel,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if jsonOutput {
		data := map[string]interface{}{
			"current": currentVersion,
			"channel": channel,
			"update":  updateInfo != nil,
		}
		if updateInfo != nil {
			data["latest_version"] = updateInfo.Version
			data["release_notes"] = updateInfo.ReleaseNotes
			data["is_newer"] = IsNewer(updateInfo, currentVersion)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	if updateInfo == nil {
		fmt.Printf("No update available on the %s channel.\n", channel)
		fmt.Printf("Current version: %s\n", currentVersion)
		return nil
	}

	if IsNewer(updateInfo, currentVersion) {
		fmt.Printf("Update available: %s (you have %s)\n", updateInfo.Version, currentVersion)
		if updateInfo.ReleaseNotes != "" {
			fmt.Println("\nRelease notes:")
			fmt.Println(updateInfo.ReleaseNotes)
		}
	} else {
		fmt.Printf("You are running the latest version (%s).\n", currentVersion)
	}

	return nil
}

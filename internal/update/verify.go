package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// VerifySignature verifies the binary using cosign
// Returns true if the signature is valid, false otherwise
func VerifySignature(binaryPath string, cosignKeyPath string) (bool, error) {
	// Check if cosign is installed
	if !isCosignInstalled() {
		return false, fmt.Errorf("cosign is not installed. Install from https://docs.sigstore.dev/cosign/installation/")
	}

	// Verify the signature
	cmd := exec.Command("cosign", "verify", "--key", cosignKeyPath, binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	// Check output for successful verification
	outputStr := string(output)
	if strings.Contains(outputStr, "Verified OK") || strings.Contains(outputStr, "successfully verified") {
		return true, nil
	}

	return false, nil
}

// isCosignInstalled checks if cosign is available in PATH
func isCosignInstalled() bool {
	_, err := exec.LookPath("cosign")
	return err == nil
}

// ApplyUpdate replaces the current binary with the downloaded one
func ApplyUpdate(downloadedPath string) error {
	// Get the current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable: %w", err)
	}

	// Create backup of current binary
	backupPath := currentExe + ".bak"
	if _, err := os.Stat(currentExe); err == nil {
		// Backup exists or current exe exists
		if err := os.Rename(currentExe, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Move downloaded file to replace current binary
	if err := os.Rename(downloadedPath, currentExe); err != nil {
		// Try to restore backup
		os.Rename(backupPath, currentExe)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Make the new binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			// Non-fatal on Windows
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Remove backup on success
	os.Remove(backupPath)

	return nil
}

// GetDownloadPath returns the appropriate download path for a version
func GetDownloadPath(cfg *Config, version string) string {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		platform = "windows-amd64"
	}

	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	return fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s-%s%s",
		cfg.Owner, cfg.Repo, tag, cfg.BinaryName, platform, archiveExt(),
	)
}

// ExtractArchive extracts a downloaded archive to a binary
func ExtractArchive(archivePath string, destDir string) (string, error) {
	// For now, just rename the file if it's already a binary
	// In a full implementation, this would handle .tar.gz and .zip extraction
	_, err := os.Stat(archivePath)
	if err != nil {
		return "", err
	}

	// Simple case: assume the archive IS the binary (for development)
	// Real implementation would detect archive type and extract properly
	filename := filepath.Base(archivePath)
	destPath := filepath.Join(destDir, strings.TrimSuffix(filename, archiveExt()))

	if err := os.Rename(archivePath, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

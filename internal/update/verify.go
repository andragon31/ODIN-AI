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
	if _, err := os.Stat(archivePath); err != nil {
		return "", fmt.Errorf("archive not found: %w", err)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination: %w", err)
	}

	ext := filepath.Ext(archivePath)

	// Handle .tar.gz
	if ext == ".gz" && strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	}

	// Handle .zip
	if ext == ".zip" {
		return extractZip(archivePath, destDir)
	}

	// Handle .gz (single file)
	if ext == ".gz" && !strings.HasSuffix(archivePath, ".tar.gz") {
		return extractGzip(archivePath, destDir)
	}

	// Handle .tar
	if ext == ".tar" {
		return extractTar(archivePath, destDir)
	}

	// Unknown format - assume it's already a binary
	filename := filepath.Base(archivePath)
	destPath := filepath.Join(destDir, strings.TrimSuffix(filename, ext))

	if err := copyFile(archivePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy binary: %w", err)
	}

	return destPath, nil
}

// extractTarGz extracts a .tar.gz archive
func extractTarGz(archivePath string, destDir string) (string, error) {
	// Check for tar command first
	if _, err := exec.LookPath("tar"); err == nil {
		args := []string{"-xzf", archivePath, "-C", destDir}
		cmd := exec.Command("tar", args...)
		if err := cmd.Run(); err == nil {
			// Return the first file extracted
			entries, _ := os.ReadDir(destDir)
			for _, entry := range entries {
				if !entry.IsDir() {
					return filepath.Join(destDir, entry.Name()), nil
				}
			}
		}
	}

	// Fallback: manual extraction using Go (requires external library)
	return "", fmt.Errorf("tar extraction requires 'tar' command")
}

// extractZip extracts a .zip archive
func extractZip(archivePath string, destDir string) (string, error) {
	if _, err := exec.LookPath("unzip"); err == nil {
		cmd := exec.Command("unzip", "-o", archivePath, "-d", destDir)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("unzip failed: %w", err)
		}
		// Return first file found
		filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && err == nil {
				return nil
			}
			return nil
		})
		entries, _ := os.ReadDir(destDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				return filepath.Join(destDir, entry.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("zip extraction requires 'unzip' command")
}

// extractGzip extracts a .gz file
func extractGzip(archivePath string, destDir string) (string, error) {
	if _, err := exec.LookPath("gunzip"); err == nil {
		baseName := strings.TrimSuffix(filepath.Base(archivePath), ".gz")
		destPath := filepath.Join(destDir, baseName)

		cmd := exec.Command("gunzip", "-c", archivePath)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("gunzip failed: %w", err)
		}

		if err := os.WriteFile(destPath, output, 0755); err != nil {
			return "", fmt.Errorf("failed to write extracted file: %w", err)
		}

		return destPath, nil
	}

	return "", fmt.Errorf("gunzip not available")
}

// extractTar extracts a .tar archive
func extractTar(archivePath string, destDir string) (string, error) {
	if _, err := exec.LookPath("tar"); err == nil {
		args := []string{"-xf", archivePath, "-C", destDir}
		cmd := exec.Command("tar", args...)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("tar extraction failed: %w", err)
		}
		entries, _ := os.ReadDir(destDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				return filepath.Join(destDir, entry.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("tar not available")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Check if it's a valid archive by magic bytes
	if len(data) > 2 {
		if data[0] == 0x1f && data[1] == 0x8b { // gzip magic
			return fmt.Errorf("file is compressed, not a binary")
		}
	}

	return os.WriteFile(dst, data, 0755)
}

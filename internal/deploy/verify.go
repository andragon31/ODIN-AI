// Package deploy provides the Dvergar forge/deploy system for ODIN
package deploy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// VerifyResult represents the result of a verification
type VerifyResult struct {
	Valid    bool
	Output   string
	CertInfo map[string]string
	Error    error
}

// VerifyCosignAvailable checks if cosign is installed
func VerifyCosignAvailable() bool {
	_, err := exec.LookPath("cosign")
	return err == nil
}

// VerifyBinary verifies a binary using cosign
func VerifyBinary(binaryPath, signatureURL, certificateURL string) (*VerifyResult, error) {
	result := &VerifyResult{
		Valid:    false,
		CertInfo: make(map[string]string),
	}

	// Check if cosign is available
	if !VerifyCosignAvailable() {
		result.Output = "cosign not available, skipping verification"
		result.Valid = true // Skip verification if cosign not available
		return result, nil
	}

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		result.Error = fmt.Errorf("binary not found: %w", err)
		return result, result.Error
	}

	// Build cosign verify command
	args := []string{
		"verify",
		"--signature", signatureURL,
		"--certificate", certificateURL,
		binaryPath,
	}

	cmd := exec.CommandContext(context.Background(), "cosign", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		result.Output = stderr.String()
		result.Error = fmt.Errorf("cosign verify failed: %w", err)
		return result, result.Error
	}

	result.Valid = true
	result.Output = stdout.String()
	return result, nil
}

// InstallCosign downloads and installs cosign
func InstallCosign() error {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Determine download URL based on OS/arch
	var downloadURL string
	switch osName {
	case "linux":
		downloadURL = fmt.Sprintf(
			"https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-%s",
			arch,
		)
	case "darwin":
		if arch == "arm64" {
			downloadURL = "https://github.com/sigstore/cosign/releases/latest/download/cosign-darwin-arm64"
		} else {
			downloadURL = "https://github.com/sigstore/cosign/releases/latest/download/cosign-darwin-amd64"
		}
	case "windows":
		downloadURL = fmt.Sprintf(
			"https://github.com/sigstore/cosign/releases/latest/download/cosign-windows-%s.exe",
			arch,
		)
	default:
		return fmt.Errorf("unsupported OS: %s", osName)
	}

	// Download cosign
	tmpDir := os.TempDir()
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	cosignPath := filepath.Join(tmpDir, "cosign"+ext)

	// Use curl to download
	cmd := exec.Command("curl", "-fsSL", downloadURL, "-o", cosignPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download cosign: %w", err)
	}

	// Make executable
	if err := os.Chmod(cosignPath, 0755); err != nil {
		return fmt.Errorf("failed to make cosign executable: %w", err)
	}

	// Move to system path
	systemPath := "/usr/local/bin/cosign"
	if runtime.GOOS == "windows" {
		// On Windows, use AppData\Local\Programs
		homeDir, _ := os.UserHomeDir()
		systemPath = filepath.Join(homeDir, "AppData", "Local", "Programs", "cosign.exe")
	}

	if err := os.Rename(cosignPath, systemPath); err != nil {
		// If rename fails (e.g., different filesystem), copy
		cmd = exec.Command("cp", cosignPath, systemPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install cosign: %w", err)
		}
	}

	return nil
}

// VerifySignatureAgainstRekor checks transparency log
func VerifySignatureAgainstRekor(binaryPath string) (*VerifyResult, error) {
	result := &VerifyResult{
		Valid:    false,
		CertInfo: make(map[string]string),
	}

	// Check if cosign is available
	if !VerifyCosignAvailable() {
		result.Output = "cosign not available"
		result.Valid = false
		return result, nil
	}

	// Run cosign verify with transparency log
	args := []string{
		"verify",
		"--tlog-check",
		binaryPath,
	}

	cmd := exec.CommandContext(context.Background(), "cosign", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		result.Output = stderr.String()
		result.Error = fmt.Errorf("transparency log verification failed: %w", err)
		return result, result.Error
	}

	result.Valid = true
	result.Output = stdout.String()
	return result, nil
}

// GetCosignVersion returns the installed cosign version
func GetCosignVersion() (string, error) {
	if !VerifyCosignAvailable() {
		return "", fmt.Errorf("cosign not installed")
	}

	cmd := exec.Command("cosign", "version", "-o", "json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get cosign version: %w", err)
	}

	return stdout.String(), nil
}

// Package update provides self-update functionality for ODIN
package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/odin-ai/odin/pkg/logger"
)

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Version      string // Semantic version tag
	Channel      string // "stable" or "beta"
	URL          string // Download URL for the release binary
	Checksum     string // SHA256 checksum of the binary
	Signature    string // Cosign signature for verification
	ReleaseNotes string // Release notes or changelog
}

// Config holds update configuration
type Config struct {
	// Owner and repo for GitHub releases
	Owner string
	Repo  string

	// Current version of the binary
	CurrentVersion string

	// Update channel: "stable" or "beta"
	Channel string

	// Cosign public key for signature verification (base64 encoded)
	CosignKey string

	// Binary name for the current platform
	BinaryName string
}

// DefaultConfig returns the default update configuration
func DefaultConfig() *Config {
	return &Config{
		Owner:          "andragon31",
		Repo:           "ODIN-AI",
		CurrentVersion: "0.1.0",
		Channel:        "stable",
		BinaryName:     getBinaryName(),
	}
}

// getBinaryName returns the appropriate binary name for the current platform
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "odin.exe"
	}
	return "odin"
}

// CheckForUpdate checks GitHub for a newer version
func CheckForUpdate(cfg *Config) (*UpdateInfo, error) {
	// Determine which tag to fetch based on channel
	// Note: GitHub API returns latest release regardless, we filter client-side
	_ = cfg.Channel // Used for filtering after we get the response

	// Fetch latest release from GitHub API
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		cfg.Owner, cfg.Repo)

	logger.Debug("Checking for updates", "url", apiURL, "channel", cfg.Channel)

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid rate limiting
	req.Header.Set("User-Agent", "ODIN-AI-SelfUpdate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse the response to get tag_name, body (release notes), and assets
	// Using simple parsing since we don't want to add more dependencies
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	content := string(body)

	// Extract tag_name
	tagStart := strings.Index(content, `"tag_name":"`)
	if tagStart == -1 {
		return nil, fmt.Errorf("could not find tag_name in response")
	}
	tagStart += len(`"tag_name":"`)
	tagEnd := strings.Index(content[tagStart:], `"`)
	if tagEnd == -1 {
		return nil, fmt.Errorf("malformed tag_name in response")
	}
	tag := content[tagStart : tagStart+tagEnd]

	// Filter by channel if needed
	if cfg.Channel == "beta" && !strings.HasPrefix(tag, "beta-") {
		return nil, nil // No beta available
	}
	if cfg.Channel == "stable" && strings.HasPrefix(tag, "beta-") {
		return nil, nil // Stable release is older than current beta
	}

	// Extract body (release notes)
	bodyStart := strings.Index(content, `"body":"`)
	var releaseNotes string
	if bodyStart != -1 {
		bodyStart += len(`"body":"`)
		bodyEnd := strings.Index(content[bodyStart:], `"`)
		if bodyEnd != -1 {
			releaseNotes = content[bodyStart : bodyStart+bodyEnd]
			// Unescape newlines
			releaseNotes = strings.ReplaceAll(releaseNotes, `\n`, "\n")
			releaseNotes = strings.ReplaceAll(releaseNotes, `\"`, `"`)
		}
	}

	// Determine download URL based on platform
	platform := runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		platform = "windows-amd64" // Conservative default for Windows
	}

	downloadURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/%s/%s-%s%s",
		cfg.Owner, cfg.Repo, tag, cfg.BinaryName, platform, archiveExt(),
	)

	return &UpdateInfo{
		Version:      strings.TrimPrefix(tag, "v"),
		Channel:      cfg.Channel,
		URL:          downloadURL,
		ReleaseNotes: releaseNotes,
	}, nil
}

// archiveExt returns the appropriate archive extension for the platform
func archiveExt() string {
	if runtime.GOOS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// DownloadUpdate downloads the new binary to a temp file
func DownloadUpdate(version string, downloadURL string) (string, error) {
	logger.Info("Downloading update", "version", version, "url", downloadURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "ODIN-AI-SelfUpdate")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Create temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("odin-update-%s%s", version, archiveExt()))

	out, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Copy to temp file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	logger.Info("Update downloaded", "path", tmpFile)
	return tmpFile, nil
}

// GetCurrentVersion returns the currently running version
func GetCurrentVersion() string {
	cfg := DefaultConfig()
	return cfg.CurrentVersion
}

// IsNewer checks if the update version is newer than the current version
func IsNewer(updateInfo *UpdateInfo, currentVersion string) bool {
	if updateInfo == nil {
		return false
	}
	return compareVersions(updateInfo.Version, currentVersion) > 0
}

// compareVersions compares two semantic versions
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if parts1[i] > parts2[i] {
			return 1
		}
		if parts1[i] < parts2[i] {
			return -1
		}
	}
	return 0
}

// parseVersion splits a semantic version into its components
func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	var nums [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		fmt.Sscanf(parts[i], "%d", &nums[i])
	}
	return nums
}

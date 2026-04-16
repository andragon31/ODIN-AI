// Package deploy provides the Dvergar forge/deploy system for ODIN
// Dvergar is the Norse god of blacksmiths - representing the forge and build system
package deploy

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// Dvergar represents the forge/deploy system
type Dvergar struct {
	installPath string
	backupPath  string
	logPath     string
	mode        string // "auto", "local", "docker", "cluster"
}

// InstallResult represents the result of an installation
type InstallResult struct {
	Success     bool
	Version     string
	BackupPath  string
	RollbackCmd string
	Error       error
}

// SystemInfo contains detected system information
type SystemInfo struct {
	OS          string // "linux", "darwin", "windows"
	Arch        string // "amd64", "arm64"
	Container   string // "docker", "podman", "kubernetes", ""
	User        string
	HomeDir     string
	InstallPath string
	IsWSL       bool
}

// DeployConfig holds deployment configuration
type DeployConfig struct {
	InstallPath string `mapstructure:"install_path"`
	BackupPath  string `mapstructure:"backup_path"`
	LogPath     string `mapstructure:"log_path"`
	Mode        string `mapstructure:"mode"`
	CosignKey   string `mapstructure:"cosign_key"`
	DownloadURL string `mapstructure:"download_url"`
	Version     string `mapstructure:"version"`
}

// DefaultDeployConfig returns the default deployment configuration
func DefaultDeployConfig() *DeployConfig {
	homeDir, _ := os.UserHomeDir()
	return &DeployConfig{
		InstallPath: filepath.Join(homeDir, ".local", "bin", "odin"),
		BackupPath:  filepath.Join(homeDir, ".odin", "backups"),
		LogPath:     filepath.Join(homeDir, ".odin", "logs"),
		Mode:        "auto",
		CosignKey:   "",
		DownloadURL: "https://get.odin.ai/releases",
		Version:     "1.0.0",
	}
}

// New creates a new Dvergar instance
func New(cfg *DeployConfig) *Dvergar {
	if cfg == nil {
		cfg = DefaultDeployConfig()
	}

	// Expand tilde in paths
	installPath := expandPath(cfg.InstallPath)
	backupPath := expandPath(cfg.BackupPath)
	logPath := expandPath(cfg.LogPath)

	return &Dvergar{
		installPath: installPath,
		backupPath:  backupPath,
		logPath:     logPath,
		mode:        cfg.Mode,
	}
}

// expandPath expands ~ to user home directory
func expandPath(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(homeDir, path[2:])
}

// DetectSystem gathers system information for deployment
func DetectSystem() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Container: DetectContainer(),
		IsWSL:     DetectWSL(),
	}

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	info.HomeDir = homeDir

	// Get current user
	usr, err := user.Current()
	if err != nil {
		info.User = "unknown"
	} else {
		info.User = usr.Username
	}

	// Set default install path based on OS
	info.InstallPath = DefaultInstallPath(info.OS)

	return info, nil
}

// DefaultInstallPath returns the default install path for the OS
func DefaultInstallPath(os string) string {
	homeDir, _ := os.UserHomeDir()
	switch os {
	case "linux":
		// Check if /usr/local/bin exists and is writable
		if _, err := os.Stat("/usr/local/bin"); err == nil {
			return "/usr/local/bin/odin"
		}
		return filepath.Join(homeDir, ".local", "bin", "odin")
	case "darwin":
		return filepath.Join(homeDir, ".local", "bin", "odin")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "Programs", "odin", "odin.exe")
	default:
		return filepath.Join(homeDir, ".local", "bin", "odin")
	}
}

// InstallPath returns the configured install path
func (d *Dvergar) InstallPath() string {
	return d.installPath
}

// BackupPath returns the configured backup path
func (d *Dvergar) BackupPath() string {
	return d.backupPath
}

// LogPath returns the configured log path
func (d *Dvergar) LogPath() string {
	return d.logPath
}

// Mode returns the deployment mode
func (d *Dvergar) Mode() string {
	return d.mode
}

// EnsureDirs creates necessary directories for deployment
func (d *Dvergar) EnsureDirs() error {
	dirs := []string{
		filepath.Dir(d.installPath),
		d.backupPath,
		d.logPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// IsInstalled checks if ODIN is already installed
func (d *Dvergar) IsInstalled() bool {
	_, err := os.Stat(d.installPath)
	return err == nil
}

// GetVersion returns the installed ODIN version
func (d *Dvergar) GetVersion() (string, error) {
	if !d.IsInstalled() {
		return "", fmt.Errorf("odin is not installed at %s", d.installPath)
	}

	// Read the binary to get version
	// In a real implementation, we would parse the binary or read a version file
	return "1.0.0", nil
}

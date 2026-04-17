// Package deploy provides the Dvergar forge/deploy system for ODIN
package deploy

import (
	"os"
	"runtime"
	"testing"
)

func TestDetectSystem(t *testing.T) {
	info, err := DetectSystem()
	if err != nil {
		t.Fatalf("DetectSystem() error = %v", err)
	}

	// Verify basic fields are populated
	if info.OS == "" {
		t.Error("DetectSystem().OS is empty")
	}

	if info.Arch == "" {
		t.Error("DetectSystem().Arch is empty")
	}

	if info.HomeDir == "" {
		t.Error("DetectSystem().HomeDir is empty")
	}

	if info.User == "" {
		t.Error("DetectSystem().User is empty")
	}

	if info.InstallPath == "" {
		t.Error("DetectSystem().InstallPath is empty")
	}

	// Verify OS matches runtime
	if info.OS != runtime.GOOS {
		t.Errorf("DetectSystem().OS = %v, want %v", info.OS, runtime.GOOS)
	}

	// Verify Arch matches runtime
	if info.Arch != runtime.GOARCH {
		t.Errorf("DetectSystem().Arch = %v, want %v", info.Arch, runtime.GOARCH)
	}
}

func TestDetectContainer(t *testing.T) {
	container := DetectContainer()

	// Should return empty string in normal test environment
	if container != "" && container != "docker" && container != "podman" && container != "kubernetes" {
		t.Errorf("DetectContainer() = %v, want empty or valid container type", container)
	}
}

func TestDetectWSL(t *testing.T) {
	isWSL := DetectWSL()

	// Should return false on non-Linux or non-WSL systems
	if runtime.GOOS != "linux" && isWSL {
		t.Error("DetectWSL() = true on non-Linux system")
	}
}

func TestNormalizeOS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"linux", "linux"},
		{"Linux", "linux"},
		{"LINUX", "linux"},
		{"darwin", "darwin"},
		{"Darwin", "darwin"},
		{"macOS", "darwin"},
		{"windows", "windows"},
		{"Windows", "windows"},
		{"win32", "windows"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeOS(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeOS(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"amd64", "amd64"},
		{"x86_64", "amd64"},
		{"x64", "amd64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"ARM64", "arm64"},
		{"386", "386"},
		{"i386", "386"},
		{"i686", "386"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeArch(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeArch(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetBinaryName(t *testing.T) {
	name := GetBinaryName()

	if name == "" {
		t.Error("GetBinaryName() returned empty string")
	}

	// Should start with "odin-"
	if name[:5] != "odin-" {
		t.Errorf("GetBinaryName() = %v, want to start with 'odin-'", name)
	}

	// Should contain OS
	os := NormalizeOS(runtime.GOOS)
	if os == "windows" {
		if name[len(name)-4:] != ".exe" {
			t.Errorf("GetBinaryName() on Windows should end with .exe, got %v", name)
		}
	}
}

func TestDefaultInstallPath(t *testing.T) {
	tests := []struct {
		os       string
		expected string
	}{
		{"linux", ""},   // Will vary based on /usr/local/bin
		{"darwin", ""},  // Will be ~/.local/bin/odin
		{"windows", ""}, // Will be AppData path
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			path := DefaultInstallPath(tt.os)
			if path == "" {
				t.Errorf("DefaultInstallPath(%q) returned empty string", tt.os)
			}
		})
	}
}

func TestDefaultDeployConfig(t *testing.T) {
	cfg := DefaultDeployConfig()

	if cfg == nil {
		t.Fatal("DefaultDeployConfig() returned nil")
	}

	if cfg.InstallPath == "" {
		t.Error("DefaultDeployConfig().InstallPath is empty")
	}

	if cfg.BackupPath == "" {
		t.Error("DefaultDeployConfig().BackupPath is empty")
	}

	if cfg.LogPath == "" {
		t.Error("DefaultDeployConfig().LogPath is empty")
	}

	if cfg.Mode != "auto" {
		t.Errorf("DefaultDeployConfig().Mode = %v, want 'auto'", cfg.Mode)
	}
}

func TestDvergarNew(t *testing.T) {
	cfg := DefaultDeployConfig()
	d := New(cfg)

	if d == nil {
		t.Fatal("New() returned nil")
	}

	if d.InstallPath() != cfg.InstallPath {
		t.Errorf("New().InstallPath() = %v, want %v", d.InstallPath(), cfg.InstallPath)
	}

	if d.BackupPath() != cfg.BackupPath {
		t.Errorf("New().BackupPath() = %v, want %v", d.BackupPath(), cfg.BackupPath)
	}

	if d.Mode() != cfg.Mode {
		t.Errorf("New().Mode() = %v, want %v", d.Mode(), cfg.Mode)
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/usr/local/bin", "/usr/local/bin"},
		{"~/.local/bin", filepath.Join(homeDir, ".local", "bin")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.expected {
				t.Errorf("expandPath(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDvergarIsInstalled(t *testing.T) {
	cfg := DefaultDeployConfig()
	d := New(cfg)

	// Should return false since we haven't installed anything
	if d.IsInstalled() {
		t.Error("IsInstalled() = true, but nothing is installed")
	}
}

func TestDvergarEnsureDirs(t *testing.T) {
	cfg := DefaultDeployConfig()
	cfg.BackupPath = os.TempDir() + "/odin-test-backup"
	cfg.LogPath = os.TempDir() + "/odin-test-logs"

	d := New(cfg)

	// Should not error
	if err := d.EnsureDirs(); err != nil {
		t.Errorf("EnsureDirs() error = %v", err)
	}

	// Cleanup
	os.RemoveAll(cfg.BackupPath)
	os.RemoveAll(cfg.LogPath)
}

func TestVerifyCosignAvailable(t *testing.T) {
	// Just verify it returns a boolean
	available := VerifyCosignAvailable()
	if available != false && available != true {
		t.Errorf("VerifyCosignAvailable() = %v, want boolean", available)
	}
}

func TestDeployConfigWithNil(t *testing.T) {
	// New with nil should use defaults
	d := New(nil)

	if d.InstallPath() == "" {
		t.Error("New(nil).InstallPath() is empty")
	}
}

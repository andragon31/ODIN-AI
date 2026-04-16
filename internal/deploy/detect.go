// Package deploy provides the Dvergar forge/deploy system for ODIN
package deploy

import (
	"os"
	"runtime"
	"strings"
)

// DetectContainer detects if running inside a container
func DetectContainer() string {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	// Check for Podman
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return "podman"
	}

	// Check for Kubernetes
	if _, err := os.Stat("/var/run/secrets/kubernetes.io"); err == nil {
		return "kubernetes"
	}

	// Check environment variables
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return "docker"
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "kubernetes"
	}

	// Check cgroup info on Linux
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
			content := string(data)
			if strings.Contains(content, "docker") || strings.Contains(content, "containerd") {
				return "docker"
			}
			if strings.Contains(content, "kubepods") {
				return "kubernetes"
			}
		}
	}

	return ""
}

// DetectWSL detects if running in Windows Subsystem for Linux
func DetectWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// Check for WSL-specific file
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}

	// Check for WSL in the kernel command line
	if data, err := os.ReadFile("/proc/version"); err == nil {
		content := string(data)
		if strings.Contains(strings.ToLower(content), "microsoft") {
			return true
		}
	}

	return false
}

// NormalizeOS returns a normalized OS identifier
func NormalizeOS(os string) string {
	switch os {
	case "linux", "Linux", "LINUX":
		return "linux"
	case "darwin", "Darwin", "macOS", "osx":
		return "darwin"
	case "windows", "Windows", "win32":
		return "windows"
	default:
		return os
	}
}

// NormalizeArch returns a normalized architecture identifier
func NormalizeArch(arch string) string {
	switch arch {
	case "amd64", "x86_64", "x64":
		return "amd64"
	case "arm64", "aarch64", "ARM64":
		return "arm64"
	case "386", "i386", "i686":
		return "386"
	default:
		return arch
	}
}

// GetBinaryName returns the binary name for the current OS/Arch
func GetBinaryName() string {
	os := NormalizeOS(runtime.GOOS)
	arch := NormalizeArch(runtime.GOARCH)
	ext := ""

	if os == "windows" {
		ext = ".exe"
	}

	return "odin-" + os + "-" + arch + ext
}

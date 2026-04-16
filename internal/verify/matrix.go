// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"fmt"
	"runtime"
)

// MatrixTarget represents a target platform for testing
type MatrixTarget struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
}

// MatrixRunner runs tests across multiple platforms
type MatrixRunner struct {
	targets []MatrixTarget
}

// NewMatrixRunner creates a new matrix test runner
func NewMatrixRunner() *MatrixRunner {
	return &MatrixRunner{
		targets: []MatrixTarget{
			{OS: "linux", Arch: "amd64", Version: "ubuntu-20.04"},
			{OS: "linux", Arch: "arm64", Version: "ubuntu-20.04"},
			{OS: "darwin", Arch: "amd64", Version: "macos-12"},
			{OS: "darwin", Arch: "arm64", Version: "macos-12"},
			{OS: "windows", Arch: "amd64", Version: "windows-10"},
			{OS: "linux", Arch: "amd64", Version: "arch-linux"},
			{OS: "linux", Arch: "amd64", Version: "fedora-36"},
		},
	}
}

// GetTargets returns the list of available targets
func (mr *MatrixRunner) GetTargets() []MatrixTarget {
	return mr.targets
}

// AddTarget adds a target to the matrix
func (mr *MatrixRunner) AddTarget(target MatrixTarget) {
	mr.targets = append(mr.targets, target)
}

// FilterByOS filters targets by operating system
func (mr *MatrixRunner) FilterByOS(os string) []MatrixTarget {
	filtered := make([]MatrixTarget, 0)
	for _, t := range mr.targets {
		if t.OS == os {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// FilterByArch filters targets by architecture
func (mr *MatrixRunner) FilterByArch(arch string) []MatrixTarget {
	filtered := make([]MatrixTarget, 0)
	for _, t := range mr.targets {
		if t.Arch == arch {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// CurrentPlatform returns the current platform target
func (mr *MatrixRunner) CurrentPlatform() MatrixTarget {
	return MatrixTarget{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: runtime.Version(),
	}
}

// IsSupported returns true if the given platform is supported
func (mr *MatrixRunner) IsSupported(platform string) bool {
	for _, t := range mr.targets {
		if t.OS == platform {
			return true
		}
	}
	return false
}

// GetMatrixSummary returns a human-readable matrix summary
func (mr *MatrixRunner) GetMatrixSummary() string {
	summary := fmt.Sprintf("Matrix Testing Targets (%d platforms):\n", len(mr.targets))
	for i, t := range mr.targets {
		summary += fmt.Sprintf("  %d. %s/%s (%s)\n", i+1, t.OS, t.Arch, t.Version)
	}
	return summary
}

// GetPlatformCompatibility returns compatibility info for current platform
func (mr *MatrixRunner) GetPlatformCompatibility() map[string]interface{} {
	current := mr.CurrentPlatform()

	compat := map[string]interface{}{
		"current_platform": current,
		"supported":        mr.IsSupported(current.OS),
		"all_targets":      len(mr.targets),
	}

	// Count compatible targets
	compatibleCount := 0
	for _, t := range mr.targets {
		if t.OS == current.OS {
			compatibleCount++
		}
	}
	compat["compatible_targets"] = compatibleCount

	return compat
}

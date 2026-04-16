package update

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"v1 greater patch", "1.0.1", "1.0.0", 1},
		{"v1 lesser patch", "1.0.0", "1.0.1", -1},
		{"v1 greater minor", "1.1.0", "1.0.0", 1},
		{"v1 lesser minor", "1.0.0", "1.1.0", -1},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 lesser major", "1.0.0", "2.0.0", -1},
		{"with v prefix", "v1.0.0", "1.0.0", 0},
		{"complex", "1.10.0", "1.9.0", 1},
		{"complex2", "1.9.0", "1.10.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %d, expected %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name     string
		update   *UpdateInfo
		current  string
		expected bool
	}{
		{"nil update", nil, "1.0.0", false},
		{"newer version", &UpdateInfo{Version: "1.1.0"}, "1.0.0", true},
		{"older version", &UpdateInfo{Version: "0.9.0"}, "1.0.0", false},
		{"same version", &UpdateInfo{Version: "1.0.0"}, "1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewer(tt.update, tt.current)
			if result != tt.expected {
				t.Errorf("IsNewer(%v, %s) = %v, expected %v", tt.update, tt.current, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected [3]int
	}{
		{"simple", "1.2.3", [3]int{1, 2, 3}},
		{"with v prefix", "v1.2.3", [3]int{1, 2, 3}},
		{"single digit", "1.0.0", [3]int{1, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersion(%s) = %v, expected %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetDownloadPath(t *testing.T) {
	cfg := &Config{
		Owner:      "test-owner",
		Repo:       "test-repo",
		BinaryName: "odin",
	}

	path := GetDownloadPath(cfg, "1.0.0")

	// Should contain the version with v prefix
	if path == "" {
		t.Error("GetDownloadPath returned empty string")
	}
}

func TestArchiveExt(t *testing.T) {
	ext := archiveExt()
	if ext != ".tar.gz" && ext != ".zip" {
		t.Errorf("archiveExt() returned unexpected value: %s", ext)
	}
}

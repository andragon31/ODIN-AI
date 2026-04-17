package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/odin-ai/odin/internal/config"
)

// OpenCodeDetector implements detection for OpenCode (Open source Cursor fork/alternative)
type OpenCodeDetector struct{}

func (d *OpenCodeDetector) Name() string {
	return "OpenCode"
}

func (d *OpenCodeDetector) Detect() (*config.DiscoveryResult, error) {
	configDir, err := getOpenCodeConfigDir()
	if err != nil {
		return nil, err
	}

	result := &config.DiscoveryResult{
		ToolName: "OpenCode",
		Models:   []config.ToolModel{},
		APIKeys:  make(map[string]string),
		Path:     configDir,
	}

	// OpenCode often uses models.json and auth.json in its config dir
	modelsPath := filepath.Join(configDir, "models.json")
	if data, err := os.ReadFile(modelsPath); err == nil {
		var models []config.ToolModel
		if err := json.Unmarshal(data, &models); err == nil {
			result.Models = models
		}
	}

	authPath := filepath.Join(configDir, "auth.json")
	if data, err := os.ReadFile(authPath); err == nil {
		var auth map[string]string
		if err := json.Unmarshal(data, &auth); err == nil {
			result.APIKeys = auth
		}
	}

	if len(result.Models) == 0 && len(result.APIKeys) == 0 {
		return nil, nil
	}

	return result, nil
}

func getOpenCodeConfigDir() (string, error) {
	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
	case "darwin":
		baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "linux":
		baseDir = filepath.Join(os.Getenv("HOME"), ".config")
	default:
		return "", &cursorDetectError{message: "unsupported operating system: " + runtime.GOOS}
	}
	return filepath.Join(baseDir, "OpenCode"), nil
}

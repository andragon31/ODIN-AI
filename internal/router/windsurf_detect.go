package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/odin-ai/odin/internal/config"
)

// WindsurfDetector implements the Detector interface for Windsurf
type WindsurfDetector struct{}

func (d *WindsurfDetector) Name() string {
	return "Windsurf"
}

func (d *WindsurfDetector) Detect() (*config.DiscoveryResult, error) {
	configDir, err := getWindsurfConfigDir()
	if err != nil {
		return nil, err
	}

	result := &config.DiscoveryResult{
		ToolName: "Windsurf",
		Models:   []config.ToolModel{},
		APIKeys:  make(map[string]string),
		Path:     configDir,
	}

	// Detect models and keys from settings.json
	settingsPath := filepath.Join(configDir, "User", "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		parseWindsurfSettings(data, result)
	}

	if len(result.Models) == 0 && len(result.APIKeys) == 0 {
		return nil, nil
	}

	return result, nil
}

func getWindsurfConfigDir() (string, error) {
	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
	case "darwin":
		baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "linux":
		baseDir = filepath.Join(os.Getenv("HOME"), ".config")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return filepath.Join(baseDir, "Windsurf"), nil
}

func parseWindsurfSettings(data []byte, result *config.DiscoveryResult) {
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return
	}

	// Windsurf keys often follow the VSCode pattern or specific windsurf keys
	keyMappings := map[string][]string{
		"openai":    {"windsurf.openai.apiKey", "openai.apiKey"},
		"anthropic": {"windsurf.anthropic.apiKey", "anthropic.apiKey"},
		"google":    {"windsurf.google.apiKey"},
	}

	for provider, keys := range keyMappings {
		for _, key := range keys {
			if val, ok := settings[key]; ok {
				var keyVal string
				if err := json.Unmarshal(val, &keyVal); err == nil && keyVal != "" {
					result.APIKeys[provider] = keyVal
					break
				}
			}
		}
	}
}

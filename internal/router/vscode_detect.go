package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/odin-ai/odin/internal/config"
)

// VSCodeDetector implements the Detector interface for VS Code
type VSCodeDetector struct{}

func (d *VSCodeDetector) Name() string {
	return "VS Code"
}

func (d *VSCodeDetector) Detect() (*config.DiscoveryResult, error) {
	configDir, err := getVSCodeConfigDir()
	if err != nil {
		return nil, err
	}

	result := &config.DiscoveryResult{
		ToolName: "VS Code",
		Models:   []config.ToolModel{},
		APIKeys:  make(map[string]string),
		Path:     configDir,
	}

	// Detect models and keys from settings.json
	settingsPath := filepath.Join(configDir, "User", "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		parseVSCodeSettings(data, result)
	}

	if len(result.Models) == 0 && len(result.APIKeys) == 0 {
		return nil, nil // Nothing found for this tool
	}

	return result, nil
}

func getVSCodeConfigDir() (string, error) {
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
	return filepath.Join(baseDir, "Code"), nil
}

func parseVSCodeSettings(data []byte, result *config.DiscoveryResult) {
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return
	}

	// Popular extension keys
	keyMappings := map[string][]string{
		"openai":    {"github.copilot.advanced", "openai.apiKey", "chatgpt.apiKey"},
		"anthropic": {"anthropic.apiKey", "claude.apiKey"},
		"google":    {"google.apiKey", "gemini.apiKey"},
	}

	for provider, keys := range keyMappings {
		for _, key := range keys {
			if val, ok := settings[key]; ok {
				var keyVal string
				// Some might be objects, some strings
				if err := json.Unmarshal(val, &keyVal); err == nil && keyVal != "" {
					result.APIKeys[provider] = keyVal
					break
				}
			}
		}
	}
}

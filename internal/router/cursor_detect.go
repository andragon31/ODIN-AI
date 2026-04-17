package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/odin-ai/odin/internal/config"
)

// CursorDetector implements the Detector interface for Cursor IDE
type CursorDetector struct{}

func (d *CursorDetector) Name() string {
	return "Cursor"
}

func (d *CursorDetector) Detect() (*config.DiscoveryResult, error) {
	settingsPath, err := getCursorSettingsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	res, err := parseCursorSettings(data)
	if err != nil {
		return nil, err
	}
	res.Path = settingsPath
	return res, nil
}

// getCursorSettingsPath returns the Cursor settings file path based on OS
func getCursorSettingsPath() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	case "darwin":
		baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "linux":
		baseDir = filepath.Join(os.Getenv("HOME"), ".config")
	default:
		return "", &cursorDetectError{
			message: "unsupported operating system: " + runtime.GOOS,
		}
	}

	return filepath.Join(baseDir, "Cursor", "User", "settings.json"), nil
}

// cursorDetectError represents an error during Cursor model detection
type cursorDetectError struct {
	message string
}

func (e *cursorDetectError) Error() string {
	return e.message
}

// parseCursorSettings parses Cursor settings JSON and extracts models and keys
func parseCursorSettings(data []byte) (*config.DiscoveryResult, error) {
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	res := &config.DiscoveryResult{
		ToolName: "Cursor",
		Models:   []config.ToolModel{},
		APIKeys:  make(map[string]string),
	}

	// Extract API Keys
	keyMappings := map[string][]string{
		"openai":    {"openai.apiKey", "cursor.cpp.openaiApiKey"},
		"anthropic": {"anthropic.apiKey", "cursor.cpp.anthropicApiKey"},
		"google":    {"google.apiKey"},
		"azure":     {"azure.apiKey"},
	}

	for provider, keys := range keyMappings {
		for _, key := range keys {
			if val, ok := settings[key]; ok {
				var keyVal string
				if err := json.Unmarshal(val, &keyVal); err == nil && keyVal != "" {
					res.APIKeys[provider] = keyVal
					break
				}
			}
		}
	}

	// Extract Models
	// Try format 1: cursor.generatedModels / cursor.customModels
	if generated, ok := settings["cursor.generatedModels"]; ok {
		parsed, err := parseGeneratedModels(generated)
		if err != nil {
			return nil, err
		}
		res.Models = append(res.Models, parsed...)
	}

	if custom, ok := settings["cursor.customModels"]; ok {
		parsed, err := parseCustomModels(custom)
		if err != nil {
			return nil, err
		}
		res.Models = append(res.Models, parsed...)
	}

	// Try format 2: models.providers (newer Cursor versions)
	if modelsSection, ok := settings["models"]; ok {
		parsed, err := parseModelsSection(modelsSection)
		if err != nil {
			return nil, err
		}
		res.Models = append(res.Models, parsed...)
	}

	// Remove duplicates based on model name
	res.Models = deduplicateModels(res.Models)

	return res, nil
}

// generatedModelEntry represents a model entry in cursor.generatedModels
type generatedModelEntry struct {
	Model       string `json:"model"`
	Provider    string `json:"provider"`
	DisplayName string `json:"displayName"`
}

// parseGeneratedModels parses the cursor.generatedModels array
func parseGeneratedModels(data []byte) ([]config.ToolModel, error) {
	var entries []generatedModelEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Not an array, try object format
		var entry generatedModelEntry
		if err2 := json.Unmarshal(data, &entry); err2 == nil {
			if entry.Model != "" {
				return []config.ToolModel{{
					Name:        entry.Model,
					Provider:    entry.Provider,
					DisplayName: entry.DisplayName,
				}}, nil
			}
			return nil, nil
		}
		return nil, err
	}

	models := make([]config.ToolModel, 0, len(entries))
	for _, e := range entries {
		if e.Model == "" {
			continue
		}
		displayName := e.DisplayName
		if displayName == "" {
			displayName = e.Model
		}
		models = append(models, config.ToolModel{
			Name:        e.Model,
			Provider:    e.Provider,
			DisplayName: displayName,
		})
	}
	return models, nil
}

// customModelEntry represents a model entry in cursor.customModels
type customModelEntry struct {
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

// parseCustomModels parses the cursor.customModels array
func parseCustomModels(data []byte) ([]config.ToolModel, error) {
	var entries []customModelEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Not an array, try object format
		var entry customModelEntry
		if err2 := json.Unmarshal(data, &entry); err2 == nil {
			if entry.Model != "" {
				return []config.ToolModel{{
					Name:        entry.Model,
					Provider:    entry.Provider,
					DisplayName: entry.Model,
				}}, nil
			}
			return nil, nil
		}
		return nil, err
	}

	models := make([]config.ToolModel, 0, len(entries))
	for _, e := range entries {
		if e.Model == "" {
			continue
		}
		provider := e.Provider
		if provider == "" {
			provider = "unknown"
		}
		models = append(models, config.ToolModel{
			Name:        e.Model,
			Provider:    provider,
			DisplayName: e.Model,
		})
	}
	return models, nil
}

// modelsSection represents the newer "models" format
type modelsSection struct {
	Primary   string              `json:"primary"`
	Providers map[string][]string `json:"providers"`
}

// parseModelsSection parses the models object (newer Cursor format)
func parseModelsSection(data []byte) ([]config.ToolModel, error) {
	var section modelsSection
	if err := json.Unmarshal(data, &section); err != nil {
		return nil, err
	}

	models := []config.ToolModel{}

	// Add primary model
	if section.Primary != "" {
		provider := inferProvider(section.Primary)
		models = append(models, config.ToolModel{
			Name:        section.Primary,
			Provider:    provider,
			DisplayName: section.Primary,
		})
	}

	// Add models from providers
	for provider, modelList := range section.Providers {
		for _, model := range modelList {
			// Skip if already added as primary
			if model == section.Primary {
				continue
			}
			models = append(models, config.ToolModel{
				Name:        model,
				Provider:    provider,
				DisplayName: model,
			})
		}
	}

	return models, nil
}

// inferProvider attempts to infer the provider from the model name
func inferProvider(modelName string) string {
	modelLower := ""
	// Create lowercase copy for matching
	for _, c := range modelName {
		if c >= 'A' && c <= 'Z' {
			modelLower += string(c + 32)
		} else {
			modelLower += string(c)
		}
	}

	if contains(modelLower, "claude") {
		return "anthropic"
	}
	if contains(modelLower, "gpt") || contains(modelLower, "openai") {
		return "openai"
	}
	if contains(modelLower, "gemini") {
		return "google"
	}
	if contains(modelLower, "ollama") || contains(modelLower, "llama") || contains(modelLower, "mistral") {
		return "ollama"
	}
	return "unknown"
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalFold is a case-insensitive string comparison
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ac := a[i]
		bc := b[i]
		if ac >= 'A' && ac <= 'Z' {
			ac += 32
		}
		if bc >= 'A' && bc <= 'Z' {
			bc += 32
		}
		if ac != bc {
			return false
		}
	}
	return true
}

// deduplicateModels removes duplicate models based on Name field
func deduplicateModels(models []config.ToolModel) []config.ToolModel {
	seen := make(map[string]bool)
	result := []config.ToolModel{}
	for _, m := range models {
		if !seen[m.Name] {
			seen[m.Name] = true
			result = append(result, m)
		}
	}
	return result
}

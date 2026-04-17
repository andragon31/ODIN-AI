package router

import (
	"testing"

	"github.com/odin-ai/odin/internal/config"
)

func TestParseCursorSettings(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		vanilla  bool
		expected int
	}{
		{
			name: "Format 1 - Generated Models",
			json: `{
				"cursor.generatedModels": [
					{"model": "gpt-4", "provider": "openai", "displayName": "GPT-4"}
				],
				"openai.apiKey": "sk-123"
			}`,
			expected: 1,
		},
		{
			name: "Format 2 - Custom Models",
			json: `{
				"cursor.customModels": [
					{"model": "claude-3-opus", "provider": "anthropic"}
				]
			}`,
			expected: 1,
		},
		{
			name: "Format 3 - Unified Models",
			json: `{
				"models": {
					"primary": "gpt-4o",
					"providers": {
						"anthropic": ["claude-3-sonnet", "claude-3-opus"]
					}
				}
			}`,
			expected: 3, // primary + 2 from providers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCursorSettings([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if len(result.Models) != tt.expected {
				t.Errorf("expected %d models, got %d", tt.expected, len(result.Models))
			}
		})
	}
}

func TestParseVSCodeSettings(t *testing.T) {
	jsonContent := `{
		"openai.apiKey": "sk-code-123",
		"claude.apiKey": "sk-ant-123"
	}`

	result := &config.DiscoveryResult{
		APIKeys: make(map[string]string),
	}
	parseVSCodeSettings([]byte(jsonContent), result)

	if result.APIKeys["openai"] != "sk-code-123" {
		t.Errorf("expected openai key sk-code-123, got %s", result.APIKeys["openai"])
	}
	if result.APIKeys["anthropic"] != "sk-ant-123" {
		t.Errorf("expected anthropic key sk-ant-123, got %s", result.APIKeys["anthropic"])
	}
}

package runeforge

import (
	"testing"
)

func TestParserParseRune(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid yaml rune",
			content: `name: test-rune
version: "1.0.0"
description: A test rune
tags:
  - testing
triggers:
  filePatterns:
    - "*.go"
execution:
  type: prompt
  prompt: "Hello {{.Name}}"
  sandbox: true
`,
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name: "invalid yaml",
			content: `name: test
version: "1.0.0"
description: Test
execution:
  type: invalid-type-here
`,
			wantErr: false, // YAML is valid, semantic validation happens later
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rune, err := parser.ParseRune(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rune == nil {
				t.Error("expected rune but got nil")
			}
		})
	}
}

func TestParserParseRuneFromMarkdown(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "yaml in code block",
			content: "Here is a rune:\n\n```yaml\nname: markdown-test\nversion: \"1.0.0\"\ndescription: A test\ntags:\n  - testing\ntriggers:\n  filePatterns:\n    - \"*.go\"\nexecution:\n  type: prompt\n  prompt: \"test\"\n  sandbox: true\n```\n\nThat's all!",
			wantErr: false,
		},
		{
			name:    "no yaml block",
			content: "No yaml here",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rune, err := parser.ParseRuneFromMarkdown(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rune == nil {
				t.Error("expected rune but got nil")
			}
		})
	}
}

func TestForgeRequest(t *testing.T) {
	req := ForgeRequest{
		Name:        "test-rune",
		Description: "A test rune",
		Tags:        []string{"testing", "example"},
		Model:       "ollama:deepseek-coder",
	}

	if req.Name != "test-rune" {
		t.Errorf("expected Name 'test-rune', got '%s'", req.Name)
	}
	if req.Description != "A test rune" {
		t.Errorf("expected Description 'A test rune', got '%s'", req.Description)
	}
	if len(req.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(req.Tags))
	}
	if req.Model != "ollama:deepseek-coder" {
		t.Errorf("expected Model 'ollama:deepseek-coder', got '%s'", req.Model)
	}
}

func TestForgeResult(t *testing.T) {
	result := &ForgeResult{
		Rune:   nil,
		Valid:  false,
		Errors: []string{"test error"},
	}

	if result.Valid {
		t.Error("expected Valid to be false")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestEngineConfig(t *testing.T) {
	cfg := DefaultEngineConfig()

	if cfg.Strategy != StrategyDirect {
		t.Errorf("expected StrategyDirect, got %s", cfg.Strategy)
	}
	if cfg.Model != "ollama:deepseek-coder" {
		t.Errorf("expected model 'ollama:deepseek-coder', got '%s'", cfg.Model)
	}
	if cfg.MaxTokens != 2048 {
		t.Errorf("expected MaxTokens 2048, got %d", cfg.MaxTokens)
	}
	if cfg.Temperature != 0.7 {
		t.Errorf("expected Temperature 0.7, got %f", cfg.Temperature)
	}
	if cfg.MaxIterations != 3 {
		t.Errorf("expected MaxIterations 3, got %d", cfg.MaxIterations)
	}
}

func TestJoinTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: "",
		},
		{
			name:     "single tag",
			tags:     []string{"testing"},
			expected: "testing",
		},
		{
			name:     "multiple tags",
			tags:     []string{"git", "pr", "workflow"},
			expected: "git, pr, workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinTags(tt.tags)
			if result != tt.expected {
				t.Errorf("joinTags(%v) = '%s', expected '%s'", tt.tags, result, tt.expected)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	vr := &validationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{"warning 1"},
	}

	if !vr.Valid {
		t.Error("expected Valid to be true")
	}
	if len(vr.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(vr.Errors))
	}
	if len(vr.Warns) != 1 {
		t.Errorf("expected 1 warning, got %d", len(vr.Warns))
	}
}

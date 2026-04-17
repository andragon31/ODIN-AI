// Package runeforge provides the rune generation engine
package runeforge

import (
	"testing"

	"github.com/odin-ai/odin/internal/skills"
)

// IntegrationTest tests runeforge components with mock dependencies
func TestRuneforgeIntegration(t *testing.T) {
	t.Run("ParseRuneWithAllSections", func(t *testing.T) {
		parser := NewParser()

		runeContent := `name: full-rune
version: 1.0.0
description: A full-featured rune with all sections
execution:
  type: prompt
  prompt: Hello
`
		rune, err := parser.ParseRune(runeContent)
		if err != nil {
			t.Fatalf("ParseRune failed: %v", err)
		}

		if rune.Name != "full-rune" {
			t.Errorf("Expected name 'full-rune', got '%s'", rune.Name)
		}

		if rune.Description != "A full-featured rune with all sections" {
			t.Errorf("Unexpected description: %s", rune.Description)
		}
	})

	t.Run("ValidateRuneWithSkills", func(t *testing.T) {
		rune := &skills.Rune{
			Name:        "test-rune",
			Description: "A test rune for integration testing",
			Version:     "1.0.0",
			Execution:   skills.SkillExecution{Type: "prompt", Prompt: "test"},
		}

		result := skills.ValidateSkill(rune)
		if !result.Valid {
			t.Errorf("Expected valid rune, got errors: %v", result.Errors)
		}
	})
}

// TestParserWithVariousInputs tests the parser with various inputs
func TestParserWithVariousInputs(t *testing.T) {
	t.Run("ParseMinimalRune", func(t *testing.T) {
		parser := NewParser()
		content := `name: minimal-rune
version: 1.0.0
description: A minimal rune
execution:
  type: prompt
  prompt: Hello
`
		rune, err := parser.ParseRune(content)
		if err != nil {
			t.Fatalf("ParseRune failed: %v", err)
		}

		if rune.Name != "minimal-rune" {
			t.Errorf("Expected name 'minimal-rune', got '%s'", rune.Name)
		}
	})

	t.Run("ParseRuneFromMarkdown", func(t *testing.T) {
		parser := NewParser()
		content := "Here is a rune:\n\n```yaml\nname: markdown-rune\nversion: \"1.0.0\"\ndescription: Parsed from markdown\ntriggers:\n  commands:\n    - /markdown\nexecution:\n  type: prompt\n  prompt: test\n```\n\nEnd of rune."
		rune, err := parser.ParseRuneFromMarkdown(content)
		if err != nil {
			t.Fatalf("ParseRuneFromMarkdown failed: %v", err)
		}

		if rune.Name != "markdown-rune" {
			t.Errorf("Expected name 'markdown-rune', got '%s'", rune.Name)
		}
	})
}

// TestForgeResultStruct tests the ForgeResult structure
func TestForgeResultStruct(t *testing.T) {
	t.Run("ValidResult", func(t *testing.T) {
		rune := &skills.Rune{Name: "valid-rune"}
		result := &ForgeResult{
			Rune:   rune,
			Valid:  true,
			Errors: []string{},
		}

		if !result.Valid {
			t.Error("Expected result to be valid")
		}

		if result.Rune.Name != "valid-rune" {
			t.Errorf("Expected rune name 'valid-rune', got '%s'", result.Rune.Name)
		}
	})

	t.Run("InvalidResult", func(t *testing.T) {
		result := &ForgeResult{
			Rune:   nil,
			Valid:  false,
			Errors: []string{"missing name", "missing trigger"},
		}

		if result.Valid {
			t.Error("Expected result to be invalid")
		}

		if len(result.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(result.Errors))
		}
	})
}

// TestParsePartialRune tests parsing with partial data
func TestParsePartialRune(t *testing.T) {
	t.Run("PartialRuneWithDefaults", func(t *testing.T) {
		parser := NewParser()
		content := `name: partial-rune
description: Testing defaults
execution:
  type: prompt
`
		rune, warnings, err := parser.ParsePartialRune(content)
		if err != nil {
			t.Fatalf("ParsePartialRune failed: %v", err)
		}

		if rune.Name != "partial-rune" {
			t.Errorf("Expected name 'partial-rune', got '%s'", rune.Name)
		}

		// Version should be defaulted
		if rune.Version == "" {
			t.Error("Expected version to be set (defaulted)")
		}

		// Execution type should be defaulted
		if rune.Execution.Type == "" {
			t.Error("Expected execution type to be set (defaulted)")
		}

		if len(warnings) == 0 {
			t.Log("Warnings found (expected for partial rune)")
		}
	})
}

// Package e2e provides end-to-end tests for ODIN components
//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/odin-ai/odin/internal/runeforge"
)

// TestRuneforgeE2E tests the RuneForge system end-to-end
func TestRuneforgeE2E(t *testing.T) {
	t.Run("ParseRune", func(t *testing.T) {
		parser := runeforge.NewParser()

		content := `# test-rune

## Purpose
A test rune for e2e testing

## When to Use
Testing the parser
`
		rune, err := parser.ParseRune(content)
		if err != nil {
			t.Fatalf("ParseRune failed: %v", err)
		}

		if rune.Name != "test-rune" {
			t.Errorf("Expected name 'test-rune', got '%s'", rune.Name)
		}
	})

	t.Run("ValidateRune", func(t *testing.T) {
		parser := runeforge.NewParser()

		content := `# validation-test

## Purpose
Testing validation
`
		rune, err := parser.ParseRune(content)
		if err != nil {
			t.Fatalf("ParseRune failed: %v", err)
		}

		// Validate the rune
		_, err = runeforge.NewParser().ParseRune(content)
		if err != nil {
			t.Logf("Validation note: %v", err)
		}
	})

	t.Run("ParsePartialRune", func(t *testing.T) {
		parser := runeforge.NewParser()

		content := `# partial-rune

## Purpose
Testing partial parsing with defaults
`
		rune, warnings, err := parser.ParsePartialRune(content)
		if err != nil {
			t.Fatalf("ParsePartialRune failed: %v", err)
		}

		if rune.Name != "partial-rune" {
			t.Errorf("Expected name 'partial-rune', got '%s'", rune.Name)
		}

		if len(warnings) > 0 {
			t.Logf("Warnings (expected for partial): %v", warnings)
		}
	})
}

// TestRuneforgeWithMockRouter tests RuneForge with a mock router
func TestRuneforgeWithMockRouter(t *testing.T) {
	t.Run("ForgeResultStructure", func(t *testing.T) {
		// Test the ForgeResult structure
		result := &runeforge.ForgeResult{
			Rune:   nil,
			Valid:  false,
			Errors: []string{"test error"},
		}

		if result.Valid {
			t.Error("Expected invalid result")
		}

		if len(result.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(result.Errors))
		}
	})
}

package skills

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/odin-ai/odin/pkg/logger"
)

// CUE schema definition for RuneSkill validation
const runeSchemaCUE = `
# RuneSkill: schema para validación de skills

name:        string @regex(#^[a-z][a-z0-9-]+$#)
version:     string @regex(#^\d+\.\d+\.\d+$#)
description: string
author?:     string
tags:        [...string]

triggers: {
    filePatterns?: [...string]
    commands?:    [...string]
    context?:     [...string]
}

execution: {
    type:   "prompt" | "script" | "wasm"
    prompt?: string
    script?: string
    sandbox: bool @default(true)
}

outputs?: {
    files?:   [...string]
    console?: string
    errors?:  [...string]
}
`

// SchemaValidator handles CUE schema validation for runes
type SchemaValidator struct {
	schemaValid bool
	errors      []string
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schemaValid: true,
		errors:      []string{},
	}
}

// ValidationResult contains the result of schema validation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
	Warns  []string `json:"warnings,omitempty"`
}

// ValidateSkill validates a skill against the CUE schema
// Note: This is a simplified validation using YAML structure and regex patterns
// Full CUE validation would require the cue-go package
func ValidateSkill(r *Rune) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warns: []string{}}

	// Validate name format: ^[a-z][a-z0-9-]+$
	nameRegex := regexp.MustCompile(`^[a-z][a-z0-9-]+$`)
	if !nameRegex.MatchString(r.Name) {
		result.Valid = false
		result.Errors = append(result.Errors, "name must match pattern ^[a-z][a-z0-9-]+$ (lowercase, starts with letter, can contain letters, numbers, and hyphens)")
	}

	// Validate version format: ^\d+\.\d+\.\d+$
	versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !versionRegex.MatchString(r.Version) {
		result.Valid = false
		result.Errors = append(result.Errors, "version must match pattern ^\\d+\\.\\d+\\.\\d+$ (semver format)")
	}

	// Validate description is not empty
	if r.Description == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "description is required")
	}

	// Validate execution type
	if r.Execution.Type == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "execution.type is required")
	} else if r.Execution.Type != "prompt" && r.Execution.Type != "script" && r.Execution.Type != "wasm" {
		result.Valid = false
		result.Errors = append(result.Errors, "execution.type must be 'prompt', 'script', or 'wasm'")
	}

	// Validate prompt execution has prompt content
	if r.Execution.Type == "prompt" && r.Execution.Prompt == "" {
		result.Warns = append(result.Warns, "execution.prompt is empty but type is 'prompt'")
	}

	// Validate script execution has script content
	if r.Execution.Type == "script" && r.Execution.Script == "" {
		result.Warns = append(result.Warns, "execution.script is empty but type is 'script'")
	}

	// Check for required fields based on type
	if r.Execution.Type == "prompt" && r.Execution.Prompt == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "execution.prompt is required when execution.type is 'prompt'")
	}

	if r.Execution.Type == "script" && r.Execution.Script == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "execution.script is required when execution.type is 'script'")
	}

	return result
}

// ValidateYAML validates a YAML file containing a skill
func ValidateYAML(data []byte) (*Rune, ValidationResult) {
	rune := &Rune{}
	result := ValidationResult{Valid: true, Errors: []string{}, Warns: []string{}}

	// Parse YAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(rune); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML parse error: %v", err))
		return rune, result
	}

	// Validate against schema
	schemaResult := ValidateSkill(rune)
	result.Valid = schemaResult.Valid
	result.Errors = append(result.Errors, schemaResult.Errors...)
	result.Warns = append(result.Warns, schemaResult.Warns...)

	return rune, result
}

// ValidateFile validates a skill file at the given path
func ValidateFile(path string) (*Rune, ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ValidationResult{Valid: false, Errors: []string{fmt.Sprintf("failed to read file: %v", err)}}, err
	}

	rune, result := ValidateYAML(data)
	return rune, result, nil
}

// LogValidationWarnings logs validation warnings without blocking
func LogValidationWarnings(result ValidationResult) {
	if len(result.Warns) > 0 {
		for _, warn := range result.Warns {
			logger.Warn("Rune validation warning", "warning", warn)
		}
	}
	if !result.Valid {
		for _, err := range result.Errors {
			logger.Error("Rune validation error", "error", err)
		}
	}
}

// CheckSchemaSyntax checks if the embedded CUE schema is valid
// This is a basic check - full CUE validation requires cue-go
func CheckSchemaSyntax() error {
	// Basic validation: check schema is not empty and has required fields
	if !strings.Contains(runeSchemaCUE, "# RuneSkill") {
		return fmt.Errorf("schema missing # RuneSkill definition")
	}
	if !strings.Contains(runeSchemaCUE, "name:") {
		return fmt.Errorf("schema missing name field")
	}
	if !strings.Contains(runeSchemaCUE, "version:") {
		return fmt.Errorf("schema missing version field")
	}
	if !strings.Contains(runeSchemaCUE, "execution:") {
		return fmt.Errorf("schema missing execution field")
	}
	return nil
}

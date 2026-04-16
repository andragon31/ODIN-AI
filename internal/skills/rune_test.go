package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRune(t *testing.T) {
	r := NewRune("test-skill", "1.0.0", "A test skill")

	if r.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got '%s'", r.Name)
	}
	if r.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", r.Version)
	}
	if r.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got '%s'", r.Description)
	}
	if r.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestRuneValidate(t *testing.T) {
	tests := []struct {
		name    string
		rune    *Rune
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid prompt rune",
			rune: &Rune{
				Name:        "test-skill",
				Version:     "1.0.0",
				Description: "A test skill",
				Execution:   SkillExecution{Type: "prompt", Prompt: "Hello {{.Name}}"},
			},
			wantErr: false,
		},
		{
			name: "valid script rune",
			rune: &Rune{
				Name:        "test-script",
				Version:     "0.1.0",
				Description: "A script skill",
				Execution:   SkillExecution{Type: "script", Script: "echo hello"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			rune: &Rune{
				Version:     "1.0.0",
				Description: "A test skill",
				Execution:   SkillExecution{Type: "prompt", Prompt: "test"},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing version",
			rune: &Rune{
				Name:        "test-skill",
				Description: "A test skill",
				Execution:   SkillExecution{Type: "prompt", Prompt: "test"},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing description",
			rune: &Rune{
				Name:      "test-skill",
				Version:   "1.0.0",
				Execution: SkillExecution{Type: "prompt", Prompt: "test"},
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "missing execution type",
			rune: &Rune{
				Name:        "test-skill",
				Version:     "1.0.0",
				Description: "A test skill",
			},
			wantErr: true,
			errMsg:  "execution.type is required",
		},
		{
			name: "invalid execution type",
			rune: &Rune{
				Name:        "test-skill",
				Version:     "1.0.0",
				Description: "A test skill",
				Execution:   SkillExecution{Type: "invalid"},
			},
			wantErr: true,
			errMsg:  "execution.type must be 'prompt', 'script', or 'wasm'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rune.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		file    string
		want    bool
	}{
		{"*", "anything.go", true},
		{"*.go", "test.go", true},
		{"*.go", "test.txt", false},
		{"go*", "golang.go", true},
		{"go*", "ego.go", false},
		{"*test*", "test_file.go", true},
		{"file", "file", true},
		{"file", "file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.file, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.file)
			if got != tt.want {
				t.Errorf("matchPattern(%s, %s) = %v, want %v", tt.pattern, tt.file, got, tt.want)
			}
		})
	}
}

func TestRuneMatchesTrigger(t *testing.T) {
	r := &Rune{
		Name:        "test-skill",
		Version:     "1.0.0",
		Description: "A test skill",
		Triggers: SkillTriggers{
			FilePatterns: []string{"*.go", "go.mod"},
			Commands:     []string{"test", "build"},
			Context:      []string{"CI"},
		},
		Execution: SkillExecution{Type: "prompt"},
	}

	tests := []struct {
		name         string
		filePatterns []string
		commands     []string
		context      []string
		want         bool
	}{
		{
			name:         "matches all",
			filePatterns: []string{"test.go"},
			commands:     []string{"test"},
			context:      []string{"CI"},
			want:         true,
		},
		{
			name:         "matches file only",
			filePatterns: []string{"test.go"},
			commands:     []string{},
			context:      []string{},
			want:         true,
		},
		{
			name:         "no match - file pattern",
			filePatterns: []string{"test.txt"},
			commands:     []string{},
			context:      []string{},
			want:         false,
		},
		{
			name:         "no match - command",
			filePatterns: []string{},
			commands:     []string{"deploy"},
			context:      []string{},
			want:         false,
		},
		{
			name:         "no match - context",
			filePatterns: []string{},
			commands:     []string{},
			context:      []string{"production"},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.MatchesTrigger(tt.filePatterns, tt.commands, tt.context)
			if got != tt.want {
				t.Errorf("MatchesTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name      string
		r         *Rune
		wantValid bool
		wantErrs  int
		wantWarns int
	}{
		{
			name: "valid full rune",
			r: &Rune{
				Name:        "valid-skill",
				Version:     "1.0.0",
				Description: "A valid skill",
				Tags:        []string{"testing", "CI"},
				Execution:   SkillExecution{Type: "prompt", Prompt: "Hello", Sandbox: true},
			},
			wantValid: true,
			wantErrs:  0,
			wantWarns: 0,
		},
		{
			name: "invalid name format",
			r: &Rune{
				Name:        "Invalid-Skill", // uppercase not allowed
				Version:     "1.0.0",
				Description: "Invalid name format",
				Execution:   SkillExecution{Type: "prompt", Prompt: "test"},
			},
			wantValid: false,
			wantErrs:  1,
			wantWarns: 0,
		},
		{
			name: "invalid version format",
			r: &Rune{
				Name:        "valid-skill",
				Version:     "1.0", // should be semver
				Description: "Invalid version",
				Execution:   SkillExecution{Type: "prompt", Prompt: "test"},
			},
			wantValid: false,
			wantErrs:  1,
			wantWarns: 0,
		},
		{
			name: "empty prompt warning",
			r: &Rune{
				Name:        "valid-skill",
				Version:     "1.0.0",
				Description: "Empty prompt",
				Execution:   SkillExecution{Type: "prompt", Prompt: ""},
			},
			wantValid: true, // prompt is empty but type is prompt
			wantErrs:  1,    // will error because prompt is required
			wantWarns: 1,    // will also warn that prompt is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSkill(tt.r)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateSkill().Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if len(result.Errors) != tt.wantErrs {
				t.Errorf("ValidateSkill().Errors has %d items, want %d: %v", len(result.Errors), tt.wantErrs, result.Errors)
			}
			if len(result.Warns) != tt.wantWarns {
				t.Errorf("ValidateSkill().Warns has %d items, want %d: %v", len(result.Warns), tt.wantWarns, result.Warns)
			}
		})
	}
}

func TestValidateYAML(t *testing.T) {
	validYAML := `
name: test-skill
version: 1.0.0
description: A test skill
tags:
  - testing
triggers:
  filePatterns:
    - "*.go"
  commands:
    - test
execution:
  type: prompt
  prompt: "Hello {{.Name}}"
  sandbox: true
`

	invalidYAML := `
name: Invalid-Skill-Name
version: 1.0
description: Missing required fields
execution:
  type: invalid
`

	t.Run("valid YAML", func(t *testing.T) {
		data := []byte(validYAML)
		r, result := ValidateYAML(data)

		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
		if r.Name != "test-skill" {
			t.Errorf("expected name 'test-skill', got '%s'", r.Name)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		data := []byte(invalidYAML)
		_, result := ValidateYAML(data)

		if result.Valid {
			t.Error("expected invalid, got valid")
		}
	})
}

func TestCache(t *testing.T) {
	// Create temp directory for cache
	tempDir, err := os.MkdirTemp("", "rune-cache-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create cache
	cache, err := NewCache(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Create test rune
	r := NewRune("cache-test", "1.0.0", "A cache test skill")
	r.Tags = []string{"testing"}
	r.Execution = SkillExecution{Type: "prompt", Prompt: "test"}

	// Store rune
	err = cache.Store(r, "test-source")
	if err != nil {
		t.Fatalf("failed to store rune: %v", err)
	}

	// List runes
	runes, err := cache.List()
	if err != nil {
		t.Fatalf("failed to list runes: %v", err)
	}
	if len(runes) != 1 {
		t.Errorf("expected 1 rune, got %d", len(runes))
	}

	// Load rune
	loaded, err := cache.Load("cache-test", "1.0.0")
	if err != nil {
		t.Fatalf("failed to load rune: %v", err)
	}
	if loaded.Name != "cache-test" {
		t.Errorf("expected name 'cache-test', got '%s'", loaded.Name)
	}

	// Check if installed
	if !cache.IsInstalled("cache-test", "1.0.0") {
		t.Error("expected rune to be installed")
	}

	// Search runes
	results, err := cache.Search("cache", nil)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Search by tag
	results, err = cache.Search("", []string{"testing"})
	if err != nil {
		t.Fatalf("failed to search by tag: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Get history
	history, err := cache.GetHistory("cache-test")
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}

	// Remove rune
	err = cache.Remove("cache-test", "1.0.0")
	if err != nil {
		t.Fatalf("failed to remove rune: %v", err)
	}

	if cache.IsInstalled("cache-test", "1.0.0") {
		t.Error("expected rune to be removed")
	}
}

func TestRuneIndex(t *testing.T) {
	index := NewRuneIndex()

	r := NewRune("index-test", "1.0.0", "Test")
	r.InstalledAt = "2024-01-01T00:00:00Z"

	// Add rune
	index.Add(r, "local")

	versions, ok := index.Get("index-test")
	if !ok {
		t.Fatal("expected to get versions")
	}
	if len(versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(versions))
	}

	// Add another version
	r2 := NewRune("index-test", "2.0.0", "Test v2")
	r2.InstalledAt = "2024-01-02T00:00:00Z"
	index.Add(r2, "local")

	versions, _ = index.Get("index-test")
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}

	// Remove version
	err := index.Remove("index-test", "1.0.0")
	if err != nil {
		t.Fatalf("failed to remove: %v", err)
	}

	versions, _ = index.Get("index-test")
	if len(versions) != 1 {
		t.Errorf("expected 1 version after remove, got %d", len(versions))
	}

	// Remove non-existent
	err = index.Remove("index-test", "9.0.0")
	if err == nil {
		t.Error("expected error removing non-existent version")
	}
}

func TestValidateFile(t *testing.T) {
	// Create temp file with valid rune
	tempDir, err := os.MkdirTemp("", "rune-validate-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	validPath := filepath.Join(tempDir, "valid.yaml")
	validContent := `
name: valid-skill
version: 1.0.0
description: A valid skill
execution:
  type: prompt
  prompt: "test"
  sandbox: true
`
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write valid file: %v", err)
	}

	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := `
name: Invalid-Name
version: bad-version
description: 
execution:
  type: invalid
`
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	t.Run("validate valid file", func(t *testing.T) {
		r, result, err := ValidateFile(validPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("expected valid, got errors: %v", result.Errors)
		}
		if r.Name != "valid-skill" {
			t.Errorf("expected name 'valid-skill', got '%s'", r.Name)
		}
	})

	t.Run("validate invalid file", func(t *testing.T) {
		_, result, err := ValidateFile(invalidPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Valid {
			t.Error("expected invalid")
		}
	})

	t.Run("validate non-existent file", func(t *testing.T) {
		_, _, err := ValidateFile(filepath.Join(tempDir, "non-existent.yaml"))
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestDefaultRunesPath(t *testing.T) {
	path := DefaultRunesPath()
	if path == "" {
		t.Error("expected non-empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
}

func TestCheckSchemaSyntax(t *testing.T) {
	err := CheckSchemaSyntax()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

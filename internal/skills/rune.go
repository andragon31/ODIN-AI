// Package skills provides the Runes skills registry for ODIN
package skills

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/odin-ai/odin/pkg/logger"
)

// Rune represents a skill in the Runes registry
type Rune struct {
	Name        string         `json:"name" yaml:"name"`
	Version     string         `json:"version" yaml:"version"`
	Description string         `json:"description" yaml:"description"`
	Author      string         `json:"author,omitempty" yaml:"author,omitempty"`
	Tags        []string       `json:"tags" yaml:"tags"`
	Triggers    SkillTriggers  `json:"triggers" yaml:"triggers"`
	Execution   SkillExecution `json:"execution" yaml:"execution"`
	Outputs     SkillOutputs   `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Schema      interface{}    `json:"-" yaml:"-"`   // CUE schema for validation - not serialized
	ID          string         `json:"id"`           // Unique identifier for this installation
	InstalledAt string         `json:"installed_at"` // ISO timestamp
}

// SkillTriggers defines when a skill is relevant
type SkillTriggers struct {
	FilePatterns []string `json:"filePatterns,omitempty" yaml:"filePatterns,omitempty"`
	Commands     []string `json:"commands,omitempty" yaml:"commands,omitempty"`
	Context      []string `json:"context,omitempty" yaml:"context,omitempty"`
}

// SkillExecution defines how a skill executes
type SkillExecution struct {
	Type    string `json:"type" yaml:"type"` // "prompt", "script", "wasm"
	Prompt  string `json:"prompt,omitempty" yaml:"prompt,omitempty"`
	Script  string `json:"script,omitempty" yaml:"script,omitempty"`
	Sandbox bool   `json:"sandbox" yaml:"sandbox"` // Default true
}

// SkillOutputs defines expected outputs from a skill
type SkillOutputs struct {
	Files   []string `json:"files,omitempty" yaml:"files,omitempty"`
	Console string   `json:"console,omitempty" yaml:"console,omitempty"`
	Errors  []string `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// RuneMetadata is the metadata file stored with each installed rune
type RuneMetadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	Source      string `json:"source"` // path or URL where it was installed from
}

// NewRune creates a new Rune instance
func NewRune(name, version, description string) *Rune {
	return &Rune{
		Name:        name,
		Version:     version,
		Description: description,
		Tags:        []string{},
		Triggers:    SkillTriggers{},
		Execution:   SkillExecution{Sandbox: true},
		Outputs:     SkillOutputs{},
		ID:          uuid.New().String(),
		InstalledAt: "",
	}
}

// Validate validates the rune's required fields
func (r *Rune) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Version == "" {
		return fmt.Errorf("version is required")
	}
	if r.Description == "" {
		return fmt.Errorf("description is required")
	}
	if r.Execution.Type == "" {
		return fmt.Errorf("execution.type is required")
	}
	if r.Execution.Type != "prompt" && r.Execution.Type != "script" && r.Execution.Type != "wasm" {
		return fmt.Errorf("execution.type must be 'prompt', 'script', or 'wasm'")
	}
	return nil
}

// MatchesTrigger checks if this rune matches the given trigger conditions
func (r *Rune) MatchesTrigger(filePatterns, commands, context []string) bool {
	// Check file patterns
	if len(r.Triggers.FilePatterns) > 0 {
		matched := false
		for _, pattern := range r.Triggers.FilePatterns {
			for _, file := range filePatterns {
				if matchPattern(pattern, file) {
					matched = true
					break
				}
			}
		}
		if !matched && len(filePatterns) > 0 {
			return false
		}
	}

	// Check commands
	if len(r.Triggers.Commands) > 0 {
		matched := false
		for _, triggerCmd := range r.Triggers.Commands {
			for _, cmd := range commands {
				if triggerCmd == cmd {
					matched = true
					break
				}
			}
		}
		if !matched && len(commands) > 0 {
			return false
		}
	}

	// Check context
	if len(r.Triggers.Context) > 0 {
		matched := false
		for _, triggerCtx := range r.Triggers.Context {
			for _, ctx := range context {
				if triggerCtx == ctx {
					matched = true
					break
				}
			}
		}
		if !matched && len(context) > 0 {
			return false
		}
	}

	return true
}

// matchPattern is a simple glob-style pattern matcher
func matchPattern(pattern, file string) bool {
	// Simple implementation - supports * wildcard
	if pattern == "*" {
		return true
	}
	if len(pattern) < 2 {
		return pattern == file
	}
	// Check for * at start
	if pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return contains(file, pattern[1:len(pattern)-1])
	}
	if pattern[0] == '*' {
		return hasSuffix(file, pattern[1:])
	}
	if pattern[len(pattern)-1] == '*' {
		return hasPrefix(file, pattern[:len(pattern)-1])
	}
	return pattern == file
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// RuneIndex is the index of all installed runes
type RuneIndex struct {
	Runes     map[string][]RuneMetadata `json:"runes"` // name -> versions
	UpdatedAt string                    `json:"updated_at"`
}

// NewRuneIndex creates a new rune index
func NewRuneIndex() *RuneIndex {
	return &RuneIndex{
		Runes:     make(map[string][]RuneMetadata),
		UpdatedAt: "",
	}
}

// Add adds a rune to the index
func (idx *RuneIndex) Add(r *Rune, source string) {
	meta := RuneMetadata{
		ID:          r.ID,
		Name:        r.Name,
		Version:     r.Version,
		InstalledAt: r.InstalledAt,
		Source:      source,
	}
	idx.Runes[r.Name] = append(idx.Runes[r.Name], meta)
}

// Remove removes a specific version of a rune from the index
func (idx *RuneIndex) Remove(name, version string) error {
	versions, ok := idx.Runes[name]
	if !ok {
		return fmt.Errorf("rune %s not found", name)
	}
	for i, v := range versions {
		if v.Version == version {
			idx.Runes[name] = append(versions[:i], versions[i+1:]...)
			if len(idx.Runes[name]) == 0 {
				delete(idx.Runes, name)
			}
			return nil
		}
	}
	return fmt.Errorf("version %s of rune %s not found", version, name)
}

// Get returns all versions of a rune
func (idx *RuneIndex) Get(name string) ([]RuneMetadata, bool) {
	versions, ok := idx.Runes[name]
	return versions, ok
}

// DefaultRunesPath returns the default path for runes storage
func DefaultRunesPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".odin", "runes")
}

// EnsureRunesDir ensures the runes directory exists
func EnsureRunesDir() error {
	path := DefaultRunesPath()
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create runes directory: %w", err)
	}
	logger.Debug("Runes directory", "path", path)
	return nil
}

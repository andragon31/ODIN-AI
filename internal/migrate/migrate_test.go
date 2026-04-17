// Package migrate provides Gentle AI to ODIN migration functionality
package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// MigrationConfig
// ---------------------------------------------------------------------------

func TestMigrationConfig_Defaults(t *testing.T) {
	cfg := MigrationConfig{
		Configs: "~/.gentle-ai",
	}

	if cfg.DryRun {
		t.Error("DryRun should default to false")
	}
	if cfg.Overwrite {
		t.Error("Overwrite should default to false")
	}
	if cfg.Configs != "~/.gentle-ai" {
		t.Errorf("Expected Configs '~/.gentle-ai', got '%s'", cfg.Configs)
	}
}

func TestMigrationConfig_DryRun(t *testing.T) {
	cfg := MigrationConfig{
		DryRun:  true,
		Configs: "/tmp/test",
	}
	if !cfg.DryRun {
		t.Error("DryRun should be true")
	}
}

// ---------------------------------------------------------------------------
// MigrationResult
// ---------------------------------------------------------------------------

func TestMigrationResult_JSONSerialization(t *testing.T) {
	result := MigrationResult{
		ConfigFiles:   3,
		Memories:      10,
		Skills:        5,
		RulesPolicies: 1,
		Errors:        0,
		Warnings:      2,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal MigrationResult: %v", err)
	}

	var decoded MigrationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal MigrationResult: %v", err)
	}

	if decoded.ConfigFiles != 3 {
		t.Errorf("Expected ConfigFiles=3, got %d", decoded.ConfigFiles)
	}
	if decoded.Memories != 10 {
		t.Errorf("Expected Memories=10, got %d", decoded.Memories)
	}
	if decoded.Skills != 5 {
		t.Errorf("Expected Skills=5, got %d", decoded.Skills)
	}
	if decoded.RulesPolicies != 1 {
		t.Errorf("Expected RulesPolicies=1, got %d", decoded.RulesPolicies)
	}
	if decoded.Warnings != 2 {
		t.Errorf("Expected Warnings=2, got %d", decoded.Warnings)
	}
}

func TestMigrationResult_ZeroValue(t *testing.T) {
	result := MigrationResult{}
	if result.Errors != 0 {
		t.Error("Errors should be 0 by default")
	}
	if result.ConfigFiles != 0 {
		t.Error("ConfigFiles should be 0 by default")
	}
}

// ---------------------------------------------------------------------------
// migrateConfigFiles
// ---------------------------------------------------------------------------

func TestMigrateConfigFiles_DryRun(t *testing.T) {
	// Create a temp gentle-ai source directory with mock files
	srcDir := t.TempDir()

	// Write sample config files
	configs := map[string]string{
		"config.yaml":   "model: ollama\nhost: localhost",
		"settings.json": `{"theme":"dark"}`,
	}
	for name, content := range configs {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", name, err)
		}
	}

	cfg := MigrationConfig{
		DryRun:  true,
		Backup:  false,
		Configs: srcDir,
	}

	count, err := migrateConfigFiles(srcDir, cfg)
	if err != nil {
		t.Fatalf("migrateConfigFiles dry-run failed: %v", err)
	}

	// DryRun: should count files but not write anything
	if count != 2 {
		t.Errorf("Expected 2 config files found, got %d", count)
	}

	// Verify no files were written to ~/.odin/config
	home, _ := os.UserHomeDir()
	odinCfgDir := filepath.Join(home, ".odin", "config")
	for name := range configs {
		destFile := filepath.Join(odinCfgDir, name)
		if _, err := os.Stat(destFile); err == nil {
			t.Errorf("DryRun should not write file %s", destFile)
		}
	}
}

func TestMigrateConfigFiles_EmptySource(t *testing.T) {
	srcDir := t.TempDir() // Empty — no config files

	cfg := MigrationConfig{
		DryRun: true,
	}

	count, err := migrateConfigFiles(srcDir, cfg)
	if err != nil {
		t.Fatalf("Unexpected error with empty source: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 files from empty source, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// migrateMemories
// ---------------------------------------------------------------------------

func TestMigrateMemories_NoMemoryDir(t *testing.T) {
	srcDir := t.TempDir() // No "memory/" subdirectory

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateMemories(srcDir, cfg)
	if err != nil {
		t.Fatalf("Expected no error when memory/ missing, got: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count=0, got %d", count)
	}
}

func TestMigrateMemories_DryRun_WithJSONFiles(t *testing.T) {
	srcDir := t.TempDir()
	memDir := filepath.Join(srcDir, "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatalf("Failed to create memory dir: %v", err)
	}

	// Write sample memory files
	memories := map[string]string{
		"observation_1.json": `{"content":"Go is awesome","session_id":"s1"}`,
		"observation_2.json": `{"content":"TDD saves time","session_id":"s2"}`,
		"ignore.txt":         "not a memory file",
	}
	for name, data := range memories {
		os.WriteFile(filepath.Join(memDir, name), []byte(data), 0644)
	}

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateMemories(srcDir, cfg)
	if err != nil {
		t.Fatalf("migrateMemories failed: %v", err)
	}

	// Should count .json files (2), skip .txt (1)
	if count != 2 {
		t.Errorf("Expected 2 memories, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// migrateSkills
// ---------------------------------------------------------------------------

func TestMigrateSkills_NoSkillsDir(t *testing.T) {
	srcDir := t.TempDir()

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateSkills(srcDir, cfg)
	if err != nil {
		t.Fatalf("Expected no error when skills/ missing, got: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 skills, got %d", count)
	}
}

func TestMigrateSkills_DryRun_WithYAMLFiles(t *testing.T) {
	srcDir := t.TempDir()
	skillsDir := filepath.Join(srcDir, "skills")
	os.MkdirAll(skillsDir, 0755)

	// Write sample skill files
	skills := map[string]string{
		"go-testing.yaml": "name: go-testing\nversion: 1.0.0\nprompt: Run tests with go test",
		"sdd-apply.md":    "# SDD Apply\n## Purpose\nApply SDD specs",
		"ignore.sh":       "#!/bin/bash\necho hello",
	}
	for name, data := range skills {
		os.WriteFile(filepath.Join(skillsDir, name), []byte(data), 0644)
	}

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateSkills(srcDir, cfg)
	if err != nil {
		t.Fatalf("migrateSkills failed: %v", err)
	}

	// Should count .yaml and .md files (2), skip .sh (1)
	if count != 2 {
		t.Errorf("Expected 2 skills, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// convertMemoryFormat
// ---------------------------------------------------------------------------

func TestConvertMemoryFormat_ValidJSON(t *testing.T) {
	input := []byte(`{"content":"Test memory","session_id":"sess-1","tags":["arch"]}`)

	output, err := convertMemoryFormat(input, "test.json")
	if err != nil {
		t.Fatalf("convertMemoryFormat failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(output, &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Must have the ODIN memory fields
	if _, ok := decoded["content"]; !ok {
		t.Error("Output should have 'content' field")
	}
	if _, ok := decoded["tags"]; !ok {
		t.Error("Output should have 'tags' field")
	}
	if _, ok := decoded["created_at"]; !ok {
		t.Error("Output should have 'created_at' field")
	}

	// Tags should include migration marker
	tags, ok := decoded["tags"].([]interface{})
	if !ok {
		t.Fatal("tags should be an array")
	}
	hasMigratedTag := false
	for _, tag := range tags {
		if tag == "migrated" {
			hasMigratedTag = true
		}
	}
	if !hasMigratedTag {
		t.Error("Output should include 'migrated' tag")
	}
}

func TestConvertMemoryFormat_PlainText(t *testing.T) {
	// Non-JSON input should be wrapped as text memory
	input := []byte("This is a plain text memory observation")

	output, err := convertMemoryFormat(input, "plain.md")
	if err != nil {
		t.Fatalf("convertMemoryFormat with plain text failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(output, &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	content, ok := decoded["content"].(string)
	if !ok {
		t.Fatal("content should be a string")
	}
	if !strings.Contains(content, "plain text memory") {
		t.Errorf("Content should preserve original text, got: %s", content)
	}
}

// ---------------------------------------------------------------------------
// convertSkillFormat
// ---------------------------------------------------------------------------

func TestConvertSkillFormat_ValidJSON(t *testing.T) {
	input := []byte(`{"name":"go-testing","version":"1.0.0","prompt":"Run go test ./..."}`)

	output, err := convertSkillFormat(input, "go-testing.json")
	if err != nil {
		t.Fatalf("convertSkillFormat failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(output, &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if decoded["name"] != "go-testing" {
		t.Errorf("Expected name='go-testing', got '%v'", decoded["name"])
	}
	if decoded["migrated"] != true {
		t.Error("Output should have migrated=true")
	}
	if decoded["source"] != "gentle-ai" {
		t.Errorf("Expected source='gentle-ai', got '%v'", decoded["source"])
	}
}

func TestConvertSkillFormat_PlainText(t *testing.T) {
	// Non-JSON skill (e.g., a markdown skill) should be wrapped
	input := []byte("# Go Testing\n## Purpose\nRun tests efficiently")

	output, err := convertSkillFormat(input, "go-testing.md")
	if err != nil {
		t.Fatalf("convertSkillFormat with plain text failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(output, &decoded); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Name should be derived from filename (without extension)
	if decoded["name"] != "go-testing" {
		t.Errorf("Expected name='go-testing', got '%v'", decoded["name"])
	}
	if decoded["type"] != "prompt" {
		t.Errorf("Expected type='prompt', got '%v'", decoded["type"])
	}
}

// ---------------------------------------------------------------------------
// convertRulesToRego
// ---------------------------------------------------------------------------

func TestConvertRulesToRego_GeneratesPolicy(t *testing.T) {
	agentsMd := `# AGENTS.md
## Rules
- Never add Co-Authored-By
- Always use conventional commits
`

	policies, err := convertRulesToRego(agentsMd)
	if err != nil {
		t.Fatalf("convertRulesToRego failed: %v", err)
	}

	if len(policies) == 0 {
		t.Fatal("Expected at least one policy to be generated")
	}

	policyContent, ok := policies["migrated_from_gentle_ai.rego"]
	if !ok {
		t.Fatal("Expected 'migrated_from_gentle_ai.rego' key in policies")
	}

	// Must be valid Rego (contains package declaration)
	if !strings.Contains(policyContent, "package") {
		t.Error("Policy should contain a 'package' declaration")
	}

	// Must have a default allow rule
	if !strings.Contains(policyContent, "default allow") {
		t.Error("Policy should have a 'default allow' rule")
	}
}

func TestConvertRulesToRego_EmptyInput(t *testing.T) {
	policies, err := convertRulesToRego("")
	if err != nil {
		t.Fatalf("convertRulesToRego with empty input failed: %v", err)
	}

	// Should still generate a base policy
	if len(policies) == 0 {
		t.Error("Should generate at least a base policy even with empty AGENTS.md")
	}
}

// ---------------------------------------------------------------------------
// migrateRulesPolicies
// ---------------------------------------------------------------------------

func TestMigrateRulesPolicies_NoAGENTSMD(t *testing.T) {
	srcDir := t.TempDir() // No AGENTS.md

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateRulesPolicies(srcDir, cfg)
	if err != nil {
		t.Fatalf("Expected graceful handling when AGENTS.md missing: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 policies when AGENTS.md not found, got %d", count)
	}
}

func TestMigrateRulesPolicies_DryRun_WithAGENTSMD(t *testing.T) {
	srcDir := t.TempDir()

	// Write a sample AGENTS.md
	agentsMd := `# AGENTS.md
## Rules
- Use conventional commits
- Run tests before pushing
`
	if err := os.WriteFile(filepath.Join(srcDir, "AGENTS.md"), []byte(agentsMd), 0644); err != nil {
		t.Fatalf("Failed to write AGENTS.md: %v", err)
	}

	cfg := MigrationConfig{DryRun: true}

	count, err := migrateRulesPolicies(srcDir, cfg)
	if err != nil {
		t.Fatalf("migrateRulesPolicies failed: %v", err)
	}

	// DryRun: should NOT write any files — returns 0
	if count != 0 {
		t.Errorf("DryRun should not write policies, expected count=0, got %d", count)
	}
}

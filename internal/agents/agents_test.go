package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeInstaller(t *testing.T) {
	installer := &ClaudeInstaller{}

	if installer.ID() != "claude-code" {
		t.Errorf("expected ID 'claude-code', got '%s'", installer.ID())
	}

	if installer.Name() != "Claude Code" {
		t.Errorf("expected Name 'Claude Code', got '%s'", installer.Name())
	}
}

func TestGeminiInstaller(t *testing.T) {
	installer := &GeminiInstaller{}

	if installer.ID() != "gemini-cli" {
		t.Errorf("expected ID 'gemini-cli', got '%s'", installer.ID())
	}

	if installer.Name() != "Gemini CLI" {
		t.Errorf("expected Name 'Gemini CLI', got '%s'", installer.Name())
	}
}

func TestCursorInstaller(t *testing.T) {
	installer := &CursorInstaller{}

	if installer.ID() != "cursor" {
		t.Errorf("expected ID 'cursor', got '%s'", installer.ID())
	}

	if installer.Name() != "Cursor AI" {
		t.Errorf("expected Name 'Cursor AI', got '%s'", installer.Name())
	}
}

func TestWindsurfInstaller(t *testing.T) {
	installer := &WindsurfInstaller{}

	if installer.ID() != "windsurf" {
		t.Errorf("expected ID 'windsurf', got '%s'", installer.ID())
	}

	if installer.Name() != "Windsurf AI" {
		t.Errorf("expected Name 'Windsurf AI', got '%s'", installer.Name())
	}
}

func TestOpenCodeInstaller(t *testing.T) {
	installer := &OpenCodeInstaller{}

	if installer.ID() != "opencode" {
		t.Errorf("expected ID 'opencode', got '%s'", installer.ID())
	}

	if installer.Name() != "OpenCode" {
		t.Errorf("expected Name 'OpenCode', got '%s'", installer.Name())
	}
}

func TestCodexInstaller(t *testing.T) {
	installer := &CodexInstaller{}

	if installer.ID() != "codex" {
		t.Errorf("expected ID 'codex', got '%s'", installer.ID())
	}

	if installer.Name() != "OpenAI Codex" {
		t.Errorf("expected Name 'OpenAI Codex', got '%s'", installer.Name())
	}
}

func TestDetectCLI(t *testing.T) {
	// Test that DetectCLI returns false for non-existent commands
	if DetectCLI("nonexistent-command-xyz") {
		t.Error("expected DetectCLI to return false for non-existent command")
	}
}

func TestAgentConfig(t *testing.T) {
	cfg := &AgentConfig{
		Model:      "claude-3-5-sonnet",
		RulesPath:  "/path/to/rules",
		ConfigPath: "/path/to/config",
	}

	if cfg.Model != "claude-3-5-sonnet" {
		t.Errorf("expected Model 'claude-3-5-sonnet', got '%s'", cfg.Model)
	}
}

func TestDefaultAgentConfig(t *testing.T) {
	cfg := DefaultAgentConfig()

	if cfg.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("unexpected default model: %s", cfg.Model)
	}

	if cfg.RulesPath == "" {
		t.Error("expected non-empty RulesPath")
	}

	if cfg.ConfigPath == "" {
		t.Error("expected non-empty ConfigPath")
	}
}

func TestDetectAgents(t *testing.T) {
	agents := DetectAgents()

	// Should return only agents that are available
	for _, agent := range agents {
		if !agent.Available() {
			t.Errorf("agent %s is marked available but Available() returns false", agent.Name())
		}
	}
}

func TestListAgents(t *testing.T) {
	agents := ListAgents()

	if len(agents) != 6 {
		t.Errorf("expected 6 agents, got %d", len(agents))
	}

	// Check all agents are present
	agentIDs := make(map[AgentID]bool)
	for _, agent := range agents {
		agentIDs[agent.ID()] = true
	}

	expectedIDs := []AgentID{"claude-code", "gemini-cli", "cursor", "windsurf", "opencode", "codex"}
	for _, id := range expectedIDs {
		if !agentIDs[id] {
			t.Errorf("expected agent %s not found in list", id)
		}
	}
}

func TestDetectCLIByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected AgentID
	}{
		{"claude-code", "claude-code", "claude-code"},
		{"Claude Code", "Claude Code", "claude-code"},
		{"gemini-cli", "gemini-cli", "gemini-cli"},
		{"Gemini CLI", "Gemini CLI", "gemini-cli"},
		{"cursor", "cursor", "cursor"},
		{"Cursor AI", "Cursor AI", "cursor"},
		{"invalid", "invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := DetectCLIByName(tt.input)
			if tt.expected == "" {
				if agent != nil {
					t.Errorf("expected nil for '%s', got %s", tt.input, agent.ID())
				}
			} else {
				if agent == nil {
					t.Errorf("expected agent for '%s', got nil", tt.input)
				} else if agent.ID() != tt.expected {
					t.Errorf("expected ID '%s' for '%s', got '%s'", tt.expected, tt.input, agent.ID())
				}
			}
		})
	}
}

func TestGetAgentInfo(t *testing.T) {
	info := GetAgentInfo("claude-code")

	if info == nil {
		t.Fatal("expected info for claude-code")
	}

	if info.Name != "Claude Code" {
		t.Errorf("expected Name 'Claude Code', got '%s'", info.Name)
	}

	if info.RulesPath != "~/.claude/rules/" {
		t.Errorf("unexpected RulesPath: %s", info.RulesPath)
	}

	if info.ConfigPath != "~/.claude/CLAUDE.md" {
		t.Errorf("unexpected ConfigPath: %s", info.ConfigPath)
	}

	// Test unknown agent
	nilInfo := GetAgentInfo("unknown-agent")
	if nilInfo != nil {
		t.Error("expected nil for unknown agent")
	}
}

func TestEnsureDir(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "odin-agents-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating a file path
	testPath := filepath.Join(tempDir, "subdir", "file.txt")
	if err := EnsureDir(testPath); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(testPath)); os.IsNotExist(err) {
		t.Error("expected directory to exist")
	}

	// Test empty path returns error
	if err := EnsureDir(""); err == nil {
		t.Error("expected error for empty path")
	}
}

func TestWriteFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "odin-agents-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test writing a file
	testPath := filepath.Join(tempDir, "subdir", "test.txt")
	content := []byte("test content")

	if err := WriteFile(testPath, content); err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Errorf("failed to read file: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(data))
	}
}

func TestDetector(t *testing.T) {
	detector := NewDetector()
	if detector == nil {
		t.Fatal("expected non-nil detector")
	}

	// Test DetectAgentsFast
	results := detector.DetectAgentsFast()
	if len(results) == 0 {
		t.Error("expected at least one detection result")
	}

	// Check result structure
	for _, r := range results {
		if r.AgentID == "" {
			t.Error("expected non-empty AgentID")
		}
		if r.Name == "" {
			t.Error("expected non-empty Name")
		}
	}
}

func TestDetectionResult(t *testing.T) {
	result := DetectionResult{
		AgentID:      "claude-code",
		Name:         "Claude Code",
		Available:    true,
		CLIInstalled: true,
		ConfigExists: false,
		Version:      "1.0.0",
	}

	if result.AgentID != "claude-code" {
		t.Errorf("unexpected AgentID: %s", result.AgentID)
	}
	if !result.Available {
		t.Error("expected Available to be true")
	}
	if result.Version != "1.0.0" {
		t.Errorf("unexpected Version: %s", result.Version)
	}
}

func TestGetInstalledAgents(t *testing.T) {
	agents := GetInstalledAgents()

	// All returned agents should have Available() return true
	for _, agent := range agents {
		if !agent.Available() {
			t.Errorf("agent %s is in installed list but not available", agent.Name())
		}
	}
}

func TestGetConfiguredAgents(t *testing.T) {
	agents := GetConfiguredAgents()

	// All returned agents should have CLI and config
	for _, agent := range agents {
		if !agent.Available() {
			t.Errorf("agent %s is in configured list but CLI not installed", agent.Name())
		}
		if err := agent.Verify(); err != nil {
			t.Errorf("agent %s is in configured list but verify fails: %v", agent.Name(), err)
		}
	}
}

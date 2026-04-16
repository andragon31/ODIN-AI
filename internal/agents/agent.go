// Package agents provides multi-agent configuration installer for ODIN
// It can detect and configure AI agents like Claude Code, Cursor, Gemini CLI, etc.
package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/odin-ai/odin/pkg/logger"
)

// AgentID is a unique identifier for an agent
type AgentID string

// AgentInstaller is the interface for agent configuration installers
type AgentInstaller interface {
	// ID returns the unique identifier for this agent
	ID() AgentID
	// Name returns a human-readable name for the agent
	Name() string
	// Available checks if the agent CLI is installed
	Available() bool
	// Install installs the agent configuration
	Install(cfg *AgentConfig) error
	// Verify checks if the configuration is properly installed
	Verify() error
	// Uninstall removes the agent configuration
	Uninstall() error
}

// AgentConfig holds configuration for an agent installer
type AgentConfig struct {
	Model      string // Model to use, e.g., "claude-3-5-sonnet"
	RulesPath  string // Path to rules directory
	ConfigPath string // Path to config file
}

// DetectAgents returns a list of agents that are available on the system
func DetectAgents() []AgentInstaller {
	allAgents := []AgentInstaller{
		&ClaudeInstaller{},
		&GeminiInstaller{},
		&CursorInstaller{},
		&WindsurfInstaller{},
		&OpenCodeInstaller{},
		&CodexInstaller{},
	}

	detected := make([]AgentInstaller, 0)
	for _, agent := range allAgents {
		if agent.Available() {
			detected = append(detected, agent)
			logger.Debug("Detected agent", "agent", agent.Name())
		}
	}

	return detected
}

// ListAgents returns all supported agents (not just detected ones)
func ListAgents() []AgentInstaller {
	return []AgentInstaller{
		&ClaudeInstaller{},
		&GeminiInstaller{},
		&CursorInstaller{},
		&WindsurfInstaller{},
		&OpenCodeInstaller{},
		&CodexInstaller{},
	}
}

// DetectCLI checks if a CLI command exists in PATH
func DetectCLI(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// WriteFile writes content to a file, creating directories as needed
func WriteFile(path string, content []byte) error {
	if err := EnsureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

// DefaultAgentConfig returns a default configuration
func DefaultAgentConfig() *AgentConfig {
	homeDir, _ := os.UserHomeDir()
	return &AgentConfig{
		Model:      "claude-3-5-sonnet-20241022",
		RulesPath:  filepath.Join(homeDir, ".claude", "rules"),
		ConfigPath: filepath.Join(homeDir, ".claude"),
	}
}

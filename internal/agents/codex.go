package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/odin-ai/odin/pkg/logger"
)

// CodexInstaller installs Codex configuration
type CodexInstaller struct{}

// ID returns the agent ID
func (c *CodexInstaller) ID() AgentID {
	return "codex"
}

// Name returns the agent name
func (c *CodexInstaller) Name() string {
	return "OpenAI Codex"
}

// Available checks if Codex CLI is installed
func (c *CodexInstaller) Available() bool {
	return DetectCLI("codex")
}

// codexConfigTemplate is the codex config template
const codexConfigTemplate = `# Codex Configuration - ODIN Managed

## Overview
This is an ODIN-managed Codex configuration.

## Model
{{if .Model}}Model: {{.Model}}{{else}}Using default model{{end}}

## Project Guidelines

### SDD Workflow
ODIN Spec-Driven Development:
1. Explore → 2. Propose → 3. Spec → 4. Design → 5. Tasks → 6. Apply → 7. Verify → 8. Archive

### Code Standards
- Follow existing project patterns
- Write tests for new functionality
- Document non-obvious decisions

## Integration
Managed by ODIN AI - Do not edit manually.
`

// Install installs the Codex configuration
func (c *CodexInstaller) Install(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	codexPath := filepath.Join(homeDir, ".codex")

	// Create directory
	if err := os.MkdirAll(codexPath, 0755); err != nil {
		return fmt.Errorf("failed to create .codex directory: %w", err)
	}

	// Generate config file
	tmpl, err := template.New("codexConfig").Parse(codexConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	configPath := filepath.Join(codexPath, "config.md")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config.md: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write config.md: %w", err)
	}

	logger.Info("Codex configuration installed", "path", codexPath)
	return nil
}

// Verify checks if Codex configuration is properly installed
func (c *CodexInstaller) Verify() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".codex", "config.md")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config.md not found at %s", configPath)
	}

	return nil
}

// Uninstall removes the Codex configuration
func (c *CodexInstaller) Uninstall() error {
	homeDir, _ := os.UserHomeDir()
	codexPath := filepath.Join(homeDir, ".codex")

	if err := os.RemoveAll(codexPath); err != nil {
		return fmt.Errorf("failed to remove .codex directory: %w", err)
	}

	logger.Info("Codex configuration uninstalled")
	return nil
}

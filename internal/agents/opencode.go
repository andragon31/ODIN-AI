package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/odin-ai/odin/pkg/logger"
)

// OpenCodeInstaller installs OpenCode configuration
type OpenCodeInstaller struct{}

// ID returns the agent ID
func (o *OpenCodeInstaller) ID() AgentID {
	return "opencode"
}

// Name returns the agent name
func (o *OpenCodeInstaller) Name() string {
	return "OpenCode"
}

// Available checks if OpenCode is installed
func (o *OpenCodeInstaller) Available() bool {
	homeDir, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(homeDir, ".opencode"),
		filepath.Join(homeDir, ".config", "opencode"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return DetectCLI("opencode")
}

// opencodeConfigTemplate is the opencode config template
const opencodeConfigTemplate = `# OpenCode Configuration - ODIN Managed

## Overview
This is an ODIN-managed OpenCode configuration.

## Model
{{if .Model}}Model: {{.Model}}{{else}}Using default model{{end}}

## Project Guidelines

### SDD Workflow
ODIN Spec-Driven Development:
1. Explore → 2. Propose → 3. Spec → 4. Design → 5. Tasks → 6. Apply → 7. Verify → 8. Archive

### Standards
- Follow project conventions
- Write tests
- Document decisions

## Integration
Managed by ODIN AI.
`

// Install installs the OpenCode configuration
func (o *OpenCodeInstaller) Install(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	opencodePath := filepath.Join(homeDir, ".opencode")

	// Create directory
	if err := os.MkdirAll(opencodePath, 0755); err != nil {
		return fmt.Errorf("failed to create .opencode directory: %w", err)
	}

	// Generate config file
	tmpl, err := template.New("opencodeConfig").Parse(opencodeConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	configPath := filepath.Join(opencodePath, "config.md")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config.md: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write config.md: %w", err)
	}

	logger.Info("OpenCode configuration installed", "path", opencodePath)
	return nil
}

// Verify checks if OpenCode configuration is properly installed
func (o *OpenCodeInstaller) Verify() error {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".opencode", "config.md")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config.md not found at %s", configPath)
	}

	return nil
}

// Uninstall removes the OpenCode configuration
func (o *OpenCodeInstaller) Uninstall() error {
	homeDir, _ := os.UserHomeDir()
	opencodePath := filepath.Join(homeDir, ".opencode")

	if err := os.RemoveAll(opencodePath); err != nil {
		return fmt.Errorf("failed to remove .opencode directory: %w", err)
	}

	logger.Info("OpenCode configuration uninstalled")
	return nil
}

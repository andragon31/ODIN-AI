package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/odin-ai/odin/pkg/logger"
)

// WindsurfInstaller installs Windsurf AI configuration
type WindsurfInstaller struct{}

// ID returns the agent ID
func (w *WindsurfInstaller) ID() AgentID {
	return "windsurf"
}

// Name returns the agent name
func (w *WindsurfInstaller) Name() string {
	return "Windsurf AI"
}

// Available checks if Windsurf is installed
func (w *WindsurfInstaller) Available() bool {
	homeDir, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(homeDir, ".windsurf"),
		"/Applications/Windsurf.app",
		filepath.Join(homeDir, "AppData", "Local", "Windsurf"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return DetectCLI("windsurf")
}

// windsurfRulesTemplate is the windsurf rules template
const windsurfRulesTemplate = `# Windsurf AI Rules - ODIN Managed

## Overview
This is an ODIN-managed Windsurf configuration.

## Model
{{if .Model}}Model: {{.Model}}{{else}}Using default model{{end}}

## Project Guidelines

### SDD Workflow
ODIN Spec-Driven Development workflow:
1. Explore - Investigate before committing
2. Propose - Create change proposals
3. Spec - Write specifications
4. Design - Document architecture
5. Tasks - Implementation breakdown
6. Apply - Code implementation
7. Verify - Validate against specs
8. Archive - Sync and complete

### Code Standards
- Follow existing project patterns
- Write tests for new functionality
- Document complex decisions

## Integration
Managed by ODIN AI - Do not edit manually.
`

// Install installs the Windsurf configuration
func (w *WindsurfInstaller) Install(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	windsurfPath := filepath.Join(homeDir, ".windsurf", "rules")

	// Create directory
	if err := os.MkdirAll(windsurfPath, 0755); err != nil {
		return fmt.Errorf("failed to create .windsurf/rules directory: %w", err)
	}

	// Generate windsurf.md
	tmpl, err := template.New("windsurfRules").Parse(windsurfRulesTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	rulesPath := filepath.Join(windsurfPath, "windsurf.md")
	f, err := os.Create(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to create windsurf.md: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write windsurf.md: %w", err)
	}

	logger.Info("Windsurf configuration installed", "path", windsurfPath)
	return nil
}

// Verify checks if Windsurf configuration is properly installed
func (w *WindsurfInstaller) Verify() error {
	homeDir, _ := os.UserHomeDir()
	windsurfMdPath := filepath.Join(homeDir, ".windsurf", "rules", "windsurf.md")

	if _, err := os.Stat(windsurfMdPath); os.IsNotExist(err) {
		return fmt.Errorf("windsurf.md not found at %s", windsurfMdPath)
	}

	return nil
}

// Uninstall removes the Windsurf configuration
func (w *WindsurfInstaller) Uninstall() error {
	homeDir, _ := os.UserHomeDir()
	windsurfPath := filepath.Join(homeDir, ".windsurf")

	if err := os.RemoveAll(windsurfPath); err != nil {
		return fmt.Errorf("failed to remove .windsurf directory: %w", err)
	}

	logger.Info("Windsurf configuration uninstalled")
	return nil
}

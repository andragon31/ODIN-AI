package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/odin-ai/odin/pkg/logger"
)

// CursorInstaller installs Cursor AI configuration
type CursorInstaller struct{}

// ID returns the agent ID
func (c *CursorInstaller) ID() AgentID {
	return "cursor"
}

// Name returns the agent name
func (c *CursorInstaller) Name() string {
	return "Cursor AI"
}

// Available checks if Cursor is installed (checks for cursor binary or app)
func (c *CursorInstaller) Available() bool {
	// Check common Cursor installation paths
	homeDir, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(homeDir, ".cursor"),
		"/Applications/Cursor.app",
		filepath.Join(homeDir, "AppData", "Local", "Cursor"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Also check if cursor CLI is available
	return DetectCLI("cursor")
}

// cursorRulesTemplate is the cursor rules template
const cursorRulesTemplate = `# Cursor AI Rules - ODIN Managed

## Overview
This is an ODIN-managed Cursor configuration.

## Model
{{if .Model}}Model: {{.Model}}{{else}}Using default model{{end}}

## Project Guidelines

### SDD Workflow
ODIN Spec-Driven Development:
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
- Document complex logic

## Integration
Managed by ODIN AI - Do not edit manually.
`

// Install installs the Cursor configuration
func (c *CursorInstaller) Install(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	cursorPath := filepath.Join(homeDir, ".cursor", "rules")

	// Create directory
	if err := os.MkdirAll(cursorPath, 0755); err != nil {
		return fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	// Generate cursor.md
	tmpl, err := template.New("cursorRules").Parse(cursorRulesTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	rulesPath := filepath.Join(cursorPath, "cursor.md")
	f, err := os.Create(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to create cursor.md: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write cursor.md: %w", err)
	}

	logger.Info("Cursor configuration installed", "path", cursorPath)
	return nil
}

// Verify checks if Cursor configuration is properly installed
func (c *CursorInstaller) Verify() error {
	homeDir, _ := os.UserHomeDir()
	cursorMdPath := filepath.Join(homeDir, ".cursor", "rules", "cursor.md")

	if _, err := os.Stat(cursorMdPath); os.IsNotExist(err) {
		return fmt.Errorf("cursor.md not found at %s", cursorMdPath)
	}

	return nil
}

// Uninstall removes the Cursor configuration
func (c *CursorInstaller) Uninstall() error {
	homeDir, _ := os.UserHomeDir()
	cursorPath := filepath.Join(homeDir, ".cursor")

	if err := os.RemoveAll(cursorPath); err != nil {
		return fmt.Errorf("failed to remove .cursor directory: %w", err)
	}

	logger.Info("Cursor configuration uninstalled")
	return nil
}

package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/odin-ai/odin/pkg/logger"
)

// GeminiInstaller installs Gemini CLI configuration
type GeminiInstaller struct{}

// ID returns the agent ID
func (g *GeminiInstaller) ID() AgentID {
	return "gemini-cli"
}

// Name returns the agent name
func (g *GeminiInstaller) Name() string {
	return "Gemini CLI"
}

// Available checks if Gemini CLI is installed
func (g *GeminiInstaller) Available() bool {
	return DetectCLI("gemini")
}

// geminiMdTemplate is the GEMINI.md template
const geminiMdTemplate = `# Gemini CLI - ODIN Managed Configuration

## Project Overview
This is an ODIN-managed Gemini CLI configuration.

## Model Configuration
{{if .Model}}Model: {{.Model}}{{else}}Using default model{{end}}

## Project Rules

### SDD Workflow
Follow the ODIN Spec-Driven Development workflow:
1. Explore → 2. Propose → 3. Spec → 4. Design → 5. Tasks → 6. Apply → 7. Verify → 8. Archive

### Code Guidelines
- Write clean, maintainable code
- Follow project conventions
- Document non-obvious decisions

## Integration
This configuration is managed by ODIN AI.
Do not edit manually - changes will be overwritten.
`

// Install installs the Gemini CLI configuration
func (g *GeminiInstaller) Install(cfg *AgentConfig) error {
	homeDir, _ := os.UserHomeDir()
	geminiPath := filepath.Join(homeDir, ".gemini")

	// Create directory
	if err := os.MkdirAll(geminiPath, 0755); err != nil {
		return fmt.Errorf("failed to create .gemini directory: %w", err)
	}

	// Generate GEMINI.md
	tmpl, err := template.New("geminiMd").Parse(geminiMdTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	geminiMdPath := filepath.Join(geminiPath, "GEMINI.md")
	f, err := os.Create(geminiMdPath)
	if err != nil {
		return fmt.Errorf("failed to create GEMINI.md: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write GEMINI.md: %w", err)
	}

	logger.Info("Gemini CLI configuration installed", "path", geminiPath)
	return nil
}

// Verify checks if Gemini CLI configuration is properly installed
func (g *GeminiInstaller) Verify() error {
	homeDir, _ := os.UserHomeDir()
	geminiMdPath := filepath.Join(homeDir, ".gemini", "GEMINI.md")

	if _, err := os.Stat(geminiMdPath); os.IsNotExist(err) {
		return fmt.Errorf("GEMINI.md not found at %s", geminiMdPath)
	}

	return nil
}

// Uninstall removes the Gemini CLI configuration
func (g *GeminiInstaller) Uninstall() error {
	homeDir, _ := os.UserHomeDir()
	geminiPath := filepath.Join(homeDir, ".gemini")

	if err := os.RemoveAll(geminiPath); err != nil {
		return fmt.Errorf("failed to remove .gemini directory: %w", err)
	}

	logger.Info("Gemini CLI configuration uninstalled")
	return nil
}

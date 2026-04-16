package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Detector handles detection of AI agents on the system
type Detector struct{}

// NewDetector creates a new agent detector
func NewDetector() *Detector {
	return &Detector{}
}

// DetectionResult represents the result of agent detection
type DetectionResult struct {
	AgentID      AgentID   `json:"id"`
	Name         string    `json:"name"`
	Available    bool      `json:"available"`
	CLIInstalled bool      `json:"cli_installed"`
	ConfigExists bool      `json:"config_exists"`
	Version      string    `json:"version,omitempty"`
	DetectedAt   time.Time `json:"detected_at"`
}

// DetectAll performs detection of all known agents
func (d *Detector) DetectAll() []DetectionResult {
	agents := ListAgents()
	results := make([]DetectionResult, 0, len(agents))

	for _, agent := range agents {
		result := d.Detect(agent)
		results = append(results, result)
	}

	return results
}

// Detect checks a single agent for availability and existing configuration
func (d *Detector) Detect(agent AgentInstaller) DetectionResult {
	result := DetectionResult{
		AgentID:    agent.ID(),
		Name:       agent.Name(),
		DetectedAt: time.Now(),
	}

	// Check if CLI is installed
	result.CLIInstalled = agent.Available()
	result.Available = result.CLIInstalled // Available means CLI is installed

	// Check if config already exists by verifying
	if err := agent.Verify(); err == nil {
		result.ConfigExists = true
	}

	// Try to get version if CLI is installed
	if result.CLIInstalled {
		version := d.getVersion(agent.ID())
		result.Version = version
	}

	return result
}

// DetectWithTimeout performs detection with a timeout for version checks
func (d *Detector) DetectWithTimeout(agent AgentInstaller, timeout time.Duration) DetectionResult {
	result := DetectionResult{
		AgentID:    agent.ID(),
		Name:       agent.Name(),
		DetectedAt: time.Now(),
	}

	// Check if CLI is installed
	result.CLIInstalled = agent.Available()
	result.Available = result.CLIInstalled

	// Check if config exists
	if err := agent.Verify(); err == nil {
		result.ConfigExists = true
	}

	// Get version with timeout
	if result.CLIInstalled {
		versionCh := make(chan string, 1)
		go func() {
			versionCh <- d.getVersion(agent.ID())
		}()

		select {
		case result.Version = <-versionCh:
		case <-time.After(timeout):
			result.Version = "unknown (timeout)"
		}
	}

	return result
}

// DetectAgentsFast detects all agents without version checks (fast)
func (d *Detector) DetectAgentsFast() []DetectionResult {
	agents := ListAgents()
	results := make([]DetectionResult, 0, len(agents))

	for _, agent := range agents {
		result := DetectionResult{
			AgentID:    agent.ID(),
			Name:       agent.Name(),
			Available:  agent.Available(),
			DetectedAt: time.Now(),
		}

		if err := agent.Verify(); err == nil {
			result.ConfigExists = true
		}

		results = append(results, result)
	}

	return results
}

// getVersion attempts to get the version of an agent CLI
func (d *Detector) getVersion(id AgentID) string {
	commands := map[AgentID][]string{
		"claude-code": {"claude", "--version"},
		"gemini-cli":  {"gemini", "--version"},
		"cursor":      {"cursor", "--version"},
		"windsurf":    {"windsurf", "--version"},
		"opencode":    {"opencode", "--version"},
		"codex":       {"codex", "--version"},
	}

	cmdParts, ok := commands[id]
	if !ok {
		return ""
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// DetectCLIByName detects a specific agent by name string
func DetectCLIByName(name string) AgentInstaller {
	agents := ListAgents()
	for _, agent := range agents {
		if strings.EqualFold(string(agent.ID()), name) ||
			strings.EqualFold(agent.Name(), name) {
			return agent
		}
	}
	return nil
}

// GetInstalledAgents returns only agents that have their CLI installed
func GetInstalledAgents() []AgentInstaller {
	detected := DetectAgents()
	return detected
}

// GetConfiguredAgents returns agents that have both CLI and config
func GetConfiguredAgents() []AgentInstaller {
	detector := NewDetector()
	agents := ListAgents()
	configured := make([]AgentInstaller, 0)

	for _, agent := range agents {
		result := detector.Detect(agent)
		if result.CLIInstalled && result.ConfigExists {
			configured = append(configured, agent)
		}
	}

	return configured
}

// PrintDetectionReport prints a formatted detection report
func PrintDetectionReport(results []DetectionResult) {
	fmt.Println("\n=== Agent Detection Report ===")
	fmt.Printf("%-15s %-20s %-10s %-10s %s\n",
		"ID", "Name", "CLI", "Config", "Version")
	fmt.Println(strings.Repeat("-", 70))

	for _, r := range results {
		cliStatus := "✗"
		if r.CLIInstalled {
			cliStatus = "✓"
		}
		configStatus := "✗"
		if r.ConfigExists {
			configStatus = "✓"
		}

		fmt.Printf("%-15s %-20s %-10s %-10s %s\n",
			string(r.AgentID),
			r.Name,
			cliStatus,
			configStatus,
			r.Version)
	}
	fmt.Println()
}

// AgentInfo holds information about a specific agent for CLI display
type AgentInfo struct {
	ID          AgentID
	Name        string
	Description string
	RulesPath   string
	ConfigPath  string
}

// GetAgentInfo returns detailed info about an agent
func GetAgentInfo(id AgentID) *AgentInfo {
	infoMap := map[AgentID]AgentInfo{
		"claude-code": {
			ID:          "claude-code",
			Name:        "Claude Code",
			Description: "Anthropic's Claude CLI with SDD workflow support",
			RulesPath:   "~/.claude/rules/",
			ConfigPath:  "~/.claude/CLAUDE.md",
		},
		"gemini-cli": {
			ID:          "gemini-cli",
			Name:        "Gemini CLI",
			Description: "Google's Gemini CLI integration",
			RulesPath:   "~/.gemini/",
			ConfigPath:  "~/.gemini/GEMINI.md",
		},
		"cursor": {
			ID:          "cursor",
			Name:        "Cursor AI",
			Description: "Cursor IDE with AI assistance",
			RulesPath:   "~/.cursor/rules/",
			ConfigPath:  "~/.cursor/rules/cursor.md",
		},
		"windsurf": {
			ID:          "windsurf",
			Name:        "Windsurf AI",
			Description: "Windsurf IDE with AI assistance",
			RulesPath:   "~/.windsurf/rules/",
			ConfigPath:  "~/.windsurf/rules/windsurf.md",
		},
		"opencode": {
			ID:          "opencode",
			Name:        "OpenCode",
			Description: "OpenCode AI CLI",
			RulesPath:   "~/.opencode/",
			ConfigPath:  "~/.opencode/config.md",
		},
		"codex": {
			ID:          "codex",
			Name:        "OpenAI Codex",
			Description: "OpenAI's Codex CLI",
			RulesPath:   "~/.codex/",
			ConfigPath:  "~/.codex/config.md",
		},
	}

	if info, ok := infoMap[id]; ok {
		return &info
	}
	return nil
}

// EnsureAgentConfig ensures the agent configuration directory exists
func EnsureAgentConfig(id AgentID) (string, error) {
	info := GetAgentInfo(id)
	if info == nil {
		return "", fmt.Errorf("unknown agent: %s", id)
	}

	homeDir, _ := os.UserHomeDir()
	configPath := info.ConfigPath
	if strings.HasPrefix(configPath, "~/") {
		configPath = filepath.Join(homeDir, configPath[2:])
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return configPath, nil
}

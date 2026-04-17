// Package catalog provides the component catalog system for ODIN
// Catalog lists available AI agents, components, and runes that can be installed
package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/odin-ai/odin/pkg/logger"
)

// AgentID represents a known AI agent
type AgentID string

const (
	AgentClaudeCode AgentID = "claude-code"
	AgentGeminiCLI  AgentID = "gemini-cli"
	AgentOpenCode   AgentID = "opencode"
	AgentCodex      AgentID = "codex"
	AgentCursor     AgentID = "cursor"
	AgentWindsurf   AgentID = "windsurf"
)

// Agent represents a known AI agent that can integrate with ODIN
type Agent struct {
	ID          AgentID  `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Website     string   `json:"website"`
	SupportedOS []string `json:"supported_os"`
	BinaryName  string   `json:"binary_name"`
	IsInstalled bool     `json:"is_installed"`
	Version     string   `json:"version"`
}

// Component represents an installable component in the ODIN ecosystem
type Component struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DependsOn   []string `json:"depends_on"`
	Runes       []string `json:"runes"`
	Version     string   `json:"version"`
}

// RuneInfo represents a rune available in the catalog
type RuneInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// CatalogType represents the type of catalog item
type CatalogType string

const (
	TypeAgent     CatalogType = "agents"
	TypeComponent CatalogType = "components"
	TypeRune      CatalogType = "runes"
)

// KnownAgents returns all known AI agents
func KnownAgents() []Agent {
	return []Agent{
		{
			ID:          AgentClaudeCode,
			Name:        "Claude Code",
			Description: "Anthropic's CLI tool for Claude AI coding assistance",
			Website:     "https://docs.anthropic.com/en/docs/claude-code",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "claude",
		},
		{
			ID:          AgentGeminiCLI,
			Name:        "Google Gemini CLI",
			Description: "Google's CLI tool for Gemini AI assistance",
			Website:     "https://ai.google.dev/gemini-api/docs",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "gemini",
		},
		{
			ID:          AgentOpenCode,
			Name:        "OpenCode",
			Description: "Open source AI coding agent",
			Website:     "https://github.com/opencode-ai/opencode",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "opencode",
		},
		{
			ID:          AgentCodex,
			Name:        "OpenAI Codex",
			Description: "OpenAI's AI coding assistant",
			Website:     "https://platform.openai.com/docs/guides/code",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "codex",
		},
		{
			ID:          AgentCursor,
			Name:        "Cursor",
			Description: "AI-first code editor",
			Website:     "https://cursor.sh",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "cursor",
		},
		{
			ID:          AgentWindsurf,
			Name:        "Windsurf",
			Description: "Codeium's AI coding agent",
			Website:     "https://codeium.com/windsurf",
			SupportedOS: []string{"linux", "darwin", "windows"},
			BinaryName:  "windsurf",
		},
	}
}

// KnownComponents returns all known installable components
func KnownComponents() []Component {
	return []Component{
		{
			ID:          "sdd",
			Name:        "Spec-Driven Development",
			Description: "SDD workflow system for ODIN with 9-phase lifecycle",
			DependsOn:   []string{},
			Runes:       []string{"sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive"},
			Version:     "1.0.0",
		},
		{
			ID:          "mimir",
			Name:        "Mimir Memory",
			Description: "Vector search memory with SQLite persistence",
			DependsOn:   []string{},
			Runes:       []string{},
			Version:     "1.0.0",
		},
		{
			ID:          "heimdall",
			Name:        "Heimdall Guardian",
			Description: "Security layer with SAST and OPA policy enforcement",
			DependsOn:   []string{},
			Runes:       []string{"guardian"},
			Version:     "1.0.0",
		},
		{
			ID:          "bifrost",
			Name:        "Bifrost Sync",
			Description: "CRDT-based sync with Git-backed versioning",
			DependsOn:   []string{},
			Runes:       []string{"sync"},
			Version:     "1.0.0",
		},
		{
			ID:          "nornir",
			Name:        "Nornir Testing",
			Description: "E2E testing and verification framework",
			DependsOn:   []string{},
			Runes:       []string{"go-testing"},
			Version:     "1.0.0",
		},
		{
			ID:          "volva",
			Name:        "Volva UI",
			Description: "Terminal UI with themes and accessibility",
			DependsOn:   []string{},
			Runes:       []string{},
			Version:     "1.0.0",
		},
		{
			ID:          "dvergar",
			Name:        "Dvergar Deploy",
			Description: "Install, upgrade, rollback system",
			DependsOn:   []string{},
			Runes:       []string{"deploy"},
			Version:     "1.0.0",
		},
	}
}

// AvailableRunes returns available runes in the catalog
func AvailableRunes() []RuneInfo {
	return []RuneInfo{
		{Name: "sdd-propose", Description: "Create change proposals with intent and scope", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-spec", Description: "Write specifications with requirements and scenarios", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-design", Description: "Create technical design documents", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-tasks", Description: "Break down changes into implementation tasks", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-apply", Description: "Implement code following specs", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-verify", Description: "Validate implementation matches specs", Tags: []string{"sdd", "workflow"}},
		{Name: "sdd-archive", Description: "Archive completed changes", Tags: []string{"sdd", "workflow"}},
		{Name: "go-testing", Description: "Go testing patterns with Bubbletea TUI testing", Tags: []string{"testing", "go"}},
		{Name: "guardian", Description: "Heimdall security policies", Tags: []string{"security"}},
		{Name: "sync", Description: "Bifrost CRDT sync operations", Tags: []string{"sync"}},
		{Name: "branch-pr", Description: "PR creation workflow following issue-first system", Tags: []string{"workflow", "github"}},
		{Name: "issue-creation", Description: "Issue creation workflow", Tags: []string{"workflow", "github"}},
		{Name: "skill-creator", Description: "Create new AI agent skills", Tags: []string{"skills"}},
	}
}

// CatalogManager manages the component catalog
type CatalogManager struct {
	odinPath string
}

// NewCatalogManager creates a new catalog manager
func NewCatalogManager() *CatalogManager {
	homeDir, _ := os.UserHomeDir()
	return &CatalogManager{
		odinPath: filepath.Join(homeDir, ".odin"),
	}
}

// DefaultCatalogManager returns a catalog manager with default paths
func DefaultCatalogManager() *CatalogManager {
	return NewCatalogManager()
}

// ListAgents returns all known agents
func (c *CatalogManager) ListAgents() []Agent {
	return KnownAgents()
}

// ListComponents returns all known components
func (c *CatalogManager) ListComponents() []Component {
	return KnownComponents()
}

// ListRunes returns all available runes
func (c *CatalogManager) ListRunes() []RuneInfo {
	return AvailableRunes()
}

// GetAgent returns an agent by ID
func (c *CatalogManager) GetAgent(id AgentID) *Agent {
	for _, agent := range KnownAgents() {
		if agent.ID == id {
			return &agent
		}
	}
	return nil
}

// GetComponent returns a component by ID
func (c *CatalogManager) GetComponent(id string) *Component {
	for _, comp := range KnownComponents() {
		if comp.ID == id {
			return &comp
		}
	}
	return nil
}

// GetRune returns a rune by name
func (c *CatalogManager) GetRune(name string) *RuneInfo {
	for _, rune := range AvailableRunes() {
		if rune.Name == name {
			return &rune
		}
	}
	return nil
}

// ListByType returns catalog items filtered by type
func (c *CatalogManager) ListByType(itemType CatalogType) interface{} {
	switch itemType {
	case TypeAgent:
		return c.ListAgents()
	case TypeComponent:
		return c.ListComponents()
	case TypeRune:
		return c.ListRunes()
	default:
		return nil
	}
}

// DetectInstalledAgents detects which AI agents are installed on the system and returns full Agent info
func (c *CatalogManager) DetectInstalledAgents() []Agent {
	installed := []Agent{}
	known := KnownAgents()

	path := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(path, pathSeparator)

	for _, agent := range known {
		cmd := agent.BinaryName
		found := false
		for _, p := range paths {
			executable := filepath.Join(p, cmd)
			if runtime.GOOS == "windows" {
				executable = filepath.Join(p, cmd+".exe")
			}
			if _, err := os.Stat(executable); err == nil {
				agent.IsInstalled = true
				// In a real scenario, we would run 'cmd --version' here
				// For now, we'll mark it as found
				found = true
				break
			}
		}
		if found {
			installed = append(installed, agent)
		}
	}

	return installed
}

// IsComponentInstalled checks if a component is installed
func (c *CatalogManager) IsComponentInstalled(id string) bool {
	componentPath := filepath.Join(c.odinPath, id)
	if _, err := os.Stat(componentPath); err == nil {
		return true
	}
	return false
}

// IsRuneInstalled checks if a rune is installed
func (c *CatalogManager) IsRuneInstalled(name string) bool {
	runePath := filepath.Join(c.odinPath, "runes", name)
	if _, err := os.Stat(runePath); err == nil {
		return true
	}
	return false
}

// InstallComponent marks a component as installed (actual installation is done by pipeline)
func (c *CatalogManager) InstallComponent(id string) error {
	comp := c.GetComponent(id)
	if comp == nil {
		return fmt.Errorf("component %s not found in catalog", id)
	}

	componentPath := filepath.Join(c.odinPath, id)
	if err := os.MkdirAll(componentPath, 0755); err != nil {
		return fmt.Errorf("failed to create component directory: %w", err)
	}

	logger.Info("Component marked as installed", "id", id, "path", componentPath)
	return nil
}

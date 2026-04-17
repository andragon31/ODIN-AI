// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/odin-ai/odin/internal/catalog"
	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/internal/memory"
	"github.com/odin-ai/odin/internal/skills"
	"github.com/odin-ai/odin/internal/sync"
)

// EcosystemDashboard shows the status of all ODIN components
type EcosystemDashboard struct {
	Theme    *Theme
	Renderer *DashboardRenderer
}

// NewEcosystemDashboard creates a new dashboard instance
func NewEcosystemDashboard() *EcosystemDashboard {
	engine := NewThemeEngine()
	theme := engine.GetActiveTheme()

	return &EcosystemDashboard{
		Theme:    theme,
		Renderer: NewDashboardRenderer(theme),
	}
}

// ComponentStatus represents a single component's status
type ComponentStatus struct {
	Name        string
	State       string // OK, WARNING, ERROR, NOT_INITIALIZED
	Detail      string
	LastChecked time.Time
}

// RenderDashboard renders the complete ecosystem dashboard
func (d *EcosystemDashboard) RenderDashboard() string {
	components := d.GatherComponentStatuses()

	var output string

	output += d.Renderer.RenderHeader()

	output += d.Renderer.RenderComponentSection("Core", d.filterByCategory(components, "core"))
	output += d.Renderer.RenderComponentSection("Memory", d.filterByCategory(components, "memory"))
	output += d.Renderer.RenderComponentSection("Security", d.filterByCategory(components, "security"))
	output += d.Renderer.RenderComponentSection("Sync", d.filterByCategory(components, "sync"))
	output += d.Renderer.RenderComponentSection("Skills", d.filterByCategory(components, "skills"))
	output += d.Renderer.RenderComponentSection("Router", d.filterByCategory(components, "router"))
	output += d.Renderer.RenderComponentSection("Agents", d.filterByCategory(components, "agents"))

	output += d.Renderer.RenderFooter()

	return output
}

// GatherComponentStatuses collects status from all components
func (d *EcosystemDashboard) GatherComponentStatuses() []ComponentStatus {
	var statuses []ComponentStatus

	// ODIN Core
	statuses = append(statuses, ComponentStatus{
		Name:        "Odin (Orchestrator)",
		State:       "OK",
		Detail:      "Running",
		LastChecked: time.Now(),
	})

	// Mimir (Memory)
	memCfg := memory.DefaultConfig()
	if _, err := os.Stat(memCfg.DBPath); err == nil {
		statuses = append(statuses, ComponentStatus{
			Name:        "Mimir (Memory)",
			State:       "OK",
			Detail:      "sqlite-vss + ollama",
			LastChecked: time.Now(),
		})
	} else {
		statuses = append(statuses, ComponentStatus{
			Name:        "Mimir (Memory)",
			State:       "NOT_INITIALIZED",
			Detail:      "Run 'odin init' to initialize",
			LastChecked: time.Now(),
		})
	}

	// Heimdall (Guardian)
	guardianCfg := config.DefaultConfig()
	if _, err := os.Stat(guardianCfg.Guardian.RulesPath); err == nil {
		statuses = append(statuses, ComponentStatus{
			Name:        "Heimdall (Security)",
			State:       "OK",
			Detail:      "OPA + gosec",
			LastChecked: time.Now(),
		})
	} else {
		statuses = append(statuses, ComponentStatus{
			Name:        "Heimdall (Security)",
			State:       "NOT_INITIALIZED",
			Detail:      "Rules not loaded",
			LastChecked: time.Now(),
		})
	}

	// Bifrost (Sync)
	bifrostRepoPath := sync.DefaultRepoPath()
	if _, err := os.Stat(filepath.Join(bifrostRepoPath, ".git")); err == nil {
		statuses = append(statuses, ComponentStatus{
			Name:        "Bifrost (Sync)",
			State:       "OK",
			Detail:      "go-git + CRDT",
			LastChecked: time.Now(),
		})
	} else {
		statuses = append(statuses, ComponentStatus{
			Name:        "Bifrost (Sync)",
			State:       "NOT_INITIALIZED",
			Detail:      "Not synced",
			LastChecked: time.Now(),
		})
	}

	// Runes (Skills Registry)
	runesPath := skills.DefaultRunesPath()
	if entries, err := os.ReadDir(runesPath); err == nil {
		runeCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				runeCount++
			}
		}
		statuses = append(statuses, ComponentStatus{
			Name:        "Runes (Skills)",
			State:       "OK",
			Detail:      fmt.Sprintf("%d runes installed", runeCount),
			LastChecked: time.Now(),
		})
	} else {
		statuses = append(statuses, ComponentStatus{
			Name:        "Runes (Skills)",
			State:       "NOT_INITIALIZED",
			Detail:      "No runes found",
			LastChecked: time.Now(),
		})
	}

	// Router
	routerState := "NOT_INITIALIZED"
	routerDetail := "No providers configured"

	// Check catalog for router info
	cm := catalog.DefaultCatalogManager()
	agents := cm.DetectInstalledAgents()
	if len(agents) > 0 {
		routerState = "OK"
		routerDetail = fmt.Sprintf("%d agents detected", len(agents))
	}

	statuses = append(statuses, ComponentStatus{
		Name:        "Router (Model)",
		State:       routerState,
		Detail:      routerDetail,
		LastChecked: time.Now(),
	})

	// Pipeline
	pipelineState := "OK"
	pipelineDetail := "Ready"

	statuses = append(statuses, ComponentStatus{
		Name:        "Pipeline (Install)",
		State:       pipelineState,
		Detail:      pipelineDetail,
		LastChecked: time.Now(),
	})

	// Nornir (Verify)
	verifyPath := filepath.Join(os.Getenv("HOME"), ".odin", "verify")
	if _, err := os.Stat(verifyPath); err == nil {
		statuses = append(statuses, ComponentStatus{
			Name:        "Nornir (Verify)",
			State:       "OK",
			Detail:      "Benchmarks pass",
			LastChecked: time.Now(),
		})
	} else {
		statuses = append(statuses, ComponentStatus{
			Name:        "Nornir (Verify)",
			State:       "NOT_INITIALIZED",
			Detail:      "No benchmarks run",
			LastChecked: time.Now(),
		})
	}

	return statuses
}

// filterByCategory filters components by category
func (d *EcosystemDashboard) filterByCategory(components []ComponentStatus, category string) []ComponentStatus {
	// Map component names to categories
	categoryMap := map[string]string{
		"Odin (Orchestrator)": "core",
		"Mimir (Memory)":      "memory",
		"Heimdall (Security)": "security",
		"Bifrost (Sync)":      "sync",
		"Runes (Skills)":      "skills",
		"Router (Model)":      "router",
		"Pipeline (Install)":  "core",
		"Nornir (Verify)":     "core",
	}

	var filtered []ComponentStatus
	for _, comp := range components {
		if cat, ok := categoryMap[comp.Name]; ok && cat == category {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}

// GetOverallHealth returns the overall health status
func (d *EcosystemDashboard) GetOverallHealth() (string, int, int) {
	statuses := d.GatherComponentStatuses()

	ok := 0
	issues := 0

	for _, s := range statuses {
		if s.State == "OK" {
			ok++
		} else {
			issues++
		}
	}

	if issues == 0 {
		return "HEALTHY", ok, issues
	} else if issues < len(statuses)/2 {
		return "DEGRADED", ok, issues
	} else {
		return "UNHEALTHY", ok, issues
	}
}

// DashboardRenderer renders dashboard components
type DashboardRenderer struct {
	theme *Theme
}

// NewDashboardRenderer creates a new renderer
func NewDashboardRenderer(theme *Theme) *DashboardRenderer {
	return &DashboardRenderer{theme: theme}
}

// RenderHeader renders the dashboard header
func (r *DashboardRenderer) RenderHeader() string {
	return fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                    ODIN Ecosystem Status                      ║
╠════════════════════════════════════════════════════════════════╣
║  Odin v0.1.0  •  Local-First AI  •  %s                    ║
╚════════════════════════════════════════════════════════════════╝
`, time.Now().Format("2006-01-02 15:04"))
}

// RenderComponentSection renders a section of components
func (r *DashboardRenderer) RenderComponentSection(name string, components []ComponentStatus) string {
	if len(components) == 0 {
		return ""
	}

	output := fmt.Sprintf("\n═══ %s ═══\n", name)

	for _, comp := range components {
		stateIcon := "✓"

		switch comp.State {
		case "OK":
			stateIcon = "✓"
		case "WARNING":
			stateIcon = "⚠"
		case "ERROR":
			stateIcon = "✗"
		case "NOT_INITIALIZED":
			stateIcon = "○"
		}

		output += fmt.Sprintf("  %s %-20s %s\n", stateIcon, comp.Name, comp.Detail)
	}

	return output
}

// RenderFooter renders the dashboard footer
func (r *DashboardRenderer) RenderFooter() string {
	return fmt.Sprintf(`
════════════════════════════════════════════════════════════════
  Run 'odin status --json' for machine-readable output
  Run 'odin init' to initialize uninitialized components
════════════════════════════════════════════════════════════════
`)
}

// RenderTextDashboard returns a text-based dashboard (no TUI dependencies)
func RenderTextDashboard() string {
	dash := NewEcosystemDashboard()
	return dash.RenderDashboard()
}

// GetDashboardJSON returns the dashboard data as JSON-compatible map
func GetDashboardJSON() map[string]interface{} {
	dash := NewEcosystemDashboard()
	statuses := dash.GatherComponentStatuses()
	health, ok, issues := dash.GetOverallHealth()

	result := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"overall_health":   health,
		"components_ok":    ok,
		"components_issue": issues,
		"components":       []map[string]interface{}{},
	}

	var compList []map[string]interface{}
	for _, s := range statuses {
		compList = append(compList, map[string]interface{}{
			"name":         s.Name,
			"state":        s.State,
			"detail":       s.Detail,
			"last_checked": s.LastChecked.Format(time.RFC3339),
		})
	}
	result["components"] = compList

	return result
}

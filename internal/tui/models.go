package tui

import (
	"fmt"
	"github.com/odin-ai/odin/internal/catalog"
	"github.com/odin-ai/odin/internal/config"
	"strings"
)

// ModelDashboard represents the UI for mapping models to phases/runes
type ModelDashboard struct {
	Renderer *DashboardRenderer
	Config   *config.Config
	Catalog  *catalog.CatalogManager
}

// NewModelDashboard creates a new model mapping dashboard
func NewModelDashboard(cfg *config.Config) *ModelDashboard {
	engine := NewThemeEngine()
	theme := engine.GetActiveTheme()
	return &ModelDashboard{
		Renderer: NewDashboardRenderer(theme),
		Config:   cfg,
		Catalog:  catalog.DefaultCatalogManager(),
	}
}

// RenderModelMapping returns the rendered model mapping view
func (d *ModelDashboard) RenderModelMapping() string {
	var sb strings.Builder

	sb.WriteString(d.Renderer.RenderHeader())
	sb.WriteString("\n═══ MANEJO DE MODELOS (Detección Automática) ═══\n\n")

	// Detect installed agents
	agents := d.Catalog.DetectInstalledAgents()
	if len(agents) == 0 {
		sb.WriteString("  [!] No se detectaron agentes de IA locales (Claude, Gemini, etc.)\n")
		sb.WriteString("      Asegúrate de que estén en tu PATH.\n")
	} else {
		for _, agent := range agents {
			status := "✓"
			sb.WriteString(fmt.Sprintf("  %s %-15s [%s] %s\n", status, agent.Name, agent.BinaryName, agent.Description))
		}
	}

	sb.WriteString("\n═══ ASIGNACIÓN MANUAL (Fases PENTAKILL) ═══\n\n")

	phases := []string{"proposal", "domain", "spec", "design", "tasks", "apply", "verify", "deploy", "archive"}
	for _, phase := range phases {
		mapping := "AUTO (Default)"
		if m, ok := d.Config.Router.Mapping.PhaseMappings[phase]; ok {
			mapping = m
		}
		sb.WriteString(fmt.Sprintf("  %-15s → %s\n", strings.Title(phase), mapping))
	}

	sb.WriteString("\n═══ ASIGNACIÓN MANUAL (Runes/Skills) ═══\n\n")
	if len(d.Config.Router.Mapping.RuneMappings) == 0 {
		sb.WriteString("  [No hay mapeos específicos para Runes]\n")
	} else {
		for runeName, agentID := range d.Config.Router.Mapping.RuneMappings {
			sb.WriteString(fmt.Sprintf("  %-15s → %s\n", runeName, agentID))
		}
	}

	sb.WriteString("\n" + d.Renderer.RenderFooter())

	return sb.String()
}

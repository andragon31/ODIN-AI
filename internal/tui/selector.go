// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/odin-ai/odin/internal/config"
)

// SelectorPanel represents the different focusable panels
type SelectorPanel int

const (
	ToolsPanel SelectorPanel = iota
	ModelsPanel
	TargetsPanel
)

// ModelSelector is the TUI component for manual model selection
type ModelSelector struct {
	Config   *config.Config
	Styles   *Styles
	Active   SelectorPanel
	
	// Tool selection
	ToolList   []string
	ToolCursor int
	
	// Model selection
	Models      []config.ToolModel
	ModelCursor int
	
	// Target selection (Phases/Runes)
	Targets      []string
	TargetCursor int
	
	LastMessage string
}

// NewModelSelector creates a new model selector
func NewModelSelector(cfg *config.Config) *ModelSelector {
	engine := NewThemeEngine()
	styles := NewStyles(engine.GetActiveTheme())
	
	// Initial tool list
	tools := make([]string, 0)
	for name := range cfg.Discovery.Tools {
		tools = append(tools, name)
	}
	sort.Strings(tools)
	
	// Targets (Phases + Runes)
	targets := append([]string{}, config.SDDPhases...)
	targets = append(targets, config.CommonRunes...)

	s := &ModelSelector{
		Config:  cfg,
		Styles:  styles,
		Active:  ToolsPanel,
		ToolList: tools,
		Targets:  targets,
	}
	
	s.updateModels()
	return s
}

// Init initializes the component
func (s *ModelSelector) Init() tea.Cmd {
	return nil
}

// updateModels updates the model list based on selected tool
func (s *ModelSelector) updateModels() {
	if len(s.ToolList) == 0 {
		return
	}
	
	toolName := s.ToolList[s.ToolCursor]
	if tool, ok := s.Config.Discovery.Tools[toolName]; ok {
		s.Models = tool.Models
		if s.ModelCursor >= len(s.Models) {
			s.ModelCursor = 0
		}
	}
}

// Update handles Bubbletea messages
func (s *ModelSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			s.cursorUp()
		case "down", "j":
			s.cursorDown()
		case "left", "h":
			s.panelLeft()
		case "right", "l", "tab":
			s.panelRight()
		case "enter", " ":
			return s, s.assignModel()
		case "s":
			// Save config
			homeDir := s.Config.HomeDir
			configPath := fmt.Sprintf("%s/.odin/config.yaml", homeDir)
			if err := s.Config.Save(configPath); err != nil {
				s.LastMessage = fmt.Sprintf("Error al guardar: %v", err)
			} else {
				s.LastMessage = "¡Configuración guardada exitosamente! 🧠"
			}
		}
	}
	return s, nil
}

func (s *ModelSelector) cursorUp() {
	switch s.Active {
	case ToolsPanel:
		if s.ToolCursor > 0 {
			s.ToolCursor--
			s.updateModels()
		}
	case ModelsPanel:
		if s.ModelCursor > 0 {
			s.ModelCursor--
		}
	case TargetsPanel:
		if s.TargetCursor > 0 {
			s.TargetCursor--
		}
	}
}

func (s *ModelSelector) cursorDown() {
	switch s.Active {
	case ToolsPanel:
		if s.ToolCursor < len(s.ToolList)-1 {
			s.ToolCursor++
			s.updateModels()
		}
	case ModelsPanel:
		if s.ModelCursor < len(s.Models)-1 {
			s.ModelCursor++
		}
	case TargetsPanel:
		if s.TargetCursor < len(s.Targets)-1 {
			s.TargetCursor++
		}
	}
}

func (s *ModelSelector) panelLeft() {
	if s.Active > ToolsPanel {
		s.Active--
	}
}

func (s *ModelSelector) panelRight() {
	if s.Active < TargetsPanel {
		s.Active++
	}
}

func (s *ModelSelector) assignModel() tea.Cmd {
	if len(s.Models) == 0 || len(s.ToolList) == 0 {
		return nil
	}
	
	model := s.Models[s.ModelCursor]
	target := s.Targets[s.TargetCursor]
	
	// Check if target is a phase or a rune
	isPhase := false
	for _, p := range config.SDDPhases {
		if p == target {
			isPhase = true
			break
		}
	}
	
	// Unique ID for the selected model (provider:model)
	modelID := fmt.Sprintf("%s:%s", model.Provider, model.Name)
	
	if isPhase {
		s.Config.Router.Mapping.PhaseMappings[target] = modelID
	} else {
		s.Config.Router.Mapping.RuneMappings[target] = modelID
	}
	
	s.LastMessage = fmt.Sprintf("Asignado %s a %s 🧠", model.DisplayName, target)
	return nil
}

// View renders the selector
func (s *ModelSelector) View() string {
	var builder strings.Builder
	
	builder.WriteString(s.Styles.RenderHeader("Configuración de Modelos AI"))
	builder.WriteString("\n")
	
	// Helper to render panel with title and content
	renderPanel := func(title string, items []string, cursor int, focused bool) string {
		var content strings.Builder
		for i, item := range items {
			prefix := "  "
			style := s.Styles.Muted
			if i == cursor {
				prefix = " ● "
				if focused {
					style = s.Styles.Accent
				} else {
					style = s.Styles.Secondary
				}
			}
			content.WriteString(style.Render(prefix + item + "\n"))
		}
		
		panelStyle := s.Styles.Border
		if focused {
			panelStyle = s.Styles.BorderFocus
		}
		
		return panelStyle.Width(30).Height(15).Render(
			s.Styles.Title.Render(title) + "\n" + content.String(),
		)
	}

	// Prepare data for panels
	modelNames := make([]string, len(s.Models))
	for i, m := range s.Models {
		modelNames[i] = m.DisplayName
		// Highlight if already assigned to current target
		target := s.Targets[s.TargetCursor]
		modelID := fmt.Sprintf("%s:%s", m.Provider, m.Name)
		
		var currentMapping string
		if val, ok := s.Config.Router.Mapping.PhaseMappings[target]; ok {
			currentMapping = val
		} else if val, ok := s.Config.Router.Mapping.RuneMappings[target]; ok {
			currentMapping = val
		}
		
		if currentMapping == modelID {
			modelNames[i] += " [ACTIVO]"
		}
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		renderPanel("🧠 Herramientas", s.ToolList, s.ToolCursor, s.Active == ToolsPanel),
		renderPanel("🤖 Modelos", modelNames, s.ModelCursor, s.Active == ModelsPanel),
		renderPanel("🎯 Destino", s.Targets, s.TargetCursor, s.Active == TargetsPanel),
	)
	
	builder.WriteString(panes)
	builder.WriteString("\n\n")
	
	// Help/Status footer
	if s.LastMessage != "" {
		builder.WriteString(s.Styles.Accent.Render(s.LastMessage) + "\n")
	}
	
	help := "Navegar: ← ↓ ↑ → • Seleccionar: ENTER • Guardar: 's' • Volver: ESC"
	builder.WriteString(s.Styles.Muted.Render(help))
	
	return builder.String()
}

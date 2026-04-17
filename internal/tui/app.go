// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/odin-ai/odin/internal/config"
)

// ViewID represents a specific screen in the TUI
type ViewID int

const (
	DashboardView ViewID = iota
	ModelSelectionView
)

// App is the main TUI application model
type App struct {
	ActiveView ViewID
	Dashboard  *EcosystemDashboard
	Selector   *ModelSelector
	Config     *config.Config
	Quitting   bool
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	return &App{
		ActiveView: DashboardView,
		Dashboard:  NewEcosystemDashboard(),
		Selector:   NewModelSelector(cfg),
		Config:     cfg,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			a.Quitting = true
			return a, tea.Quit
		case "m":
			if a.ActiveView == DashboardView {
				a.ActiveView = ModelSelectionView
				return a, nil
			}
		case "esc":
			if a.ActiveView == ModelSelectionView {
				a.ActiveView = DashboardView
				return a, nil
			}
		}
	}

	// Delegate to active view
	var cmd tea.Cmd
	switch a.ActiveView {
	case DashboardView:
		// EcosystemDashboard current implementation is static, 
		// but we can make it a tea.Model later if needed.
	case ModelSelectionView:
		_, cmd = a.Selector.Update(msg)
	}

	return a, cmd
}

// View renders the application
func (a *App) View() string {
	if a.Quitting {
		return "Cerrando ODIN... ¡Hasta pronto! 🧠\n"
	}

	switch a.ActiveView {
	case DashboardView:
		return a.Dashboard.RenderDashboard() + "\n Presiona 'm' para configurar modelos • 'q' para salir"
	case ModelSelectionView:
		return a.Selector.View()
	default:
		return "Error: Vista desconocida"
	}
}

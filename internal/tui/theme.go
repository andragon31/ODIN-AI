// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme
type Theme struct {
	Name       string         `json:"name"`
	Background lipgloss.Color `json:"background"`
	Card       lipgloss.Color `json:"card"`
	Text       lipgloss.Color `json:"text"`
	Accent     lipgloss.Color `json:"accent"`
	Secondary  lipgloss.Color `json:"secondary"`
	Error      lipgloss.Color `json:"error"`
	Muted      lipgloss.Color `json:"muted"`
	Success    lipgloss.Color `json:"success"`
}

// ThemeEngine manages themes and styling
type ThemeEngine struct {
	themes map[string]*Theme
	active *Theme
}

// NewThemeEngine creates a new theme engine
func NewThemeEngine() *ThemeEngine {
	engine := &ThemeEngine{
		themes: make(map[string]*Theme),
	}

	// Register built-in themes
	engine.themes["rose-pine"] = RosePineTheme()
	engine.themes["nord"] = NordTheme()
	engine.themes["catppuccin-mocha"] = CatppuccinMochaTheme()
	engine.themes["dracula"] = DraculaTheme()

	// Set default theme
	engine.active = engine.themes["rose-pine"]

	return engine
}

// GetTheme returns a theme by name
func (te *ThemeEngine) GetTheme(name string) *Theme {
	return te.themes[name]
}

// SetActiveTheme sets the active theme
func (te *ThemeEngine) SetActiveTheme(name string) error {
	theme, exists := te.themes[name]
	if !exists {
		return &ThemeNotFoundError{Name: name}
	}
	te.active = theme
	return nil
}

// GetActiveTheme returns the active theme
func (te *ThemeEngine) GetActiveTheme() *Theme {
	return te.active
}

// ListThemes returns all available themes
func (te *ThemeEngine) ListThemes() []*Theme {
	themes := make([]*Theme, 0, len(te.themes))
	for _, theme := range te.themes {
		themes = append(themes, theme)
	}
	return themes
}

// RegisterTheme registers a new theme
func (te *ThemeEngine) RegisterTheme(theme *Theme) {
	te.themes[theme.Name] = theme
}

// ThemeNotFoundError represents a theme not found error
type ThemeNotFoundError struct {
	Name string
}

func (e *ThemeNotFoundError) Error() string {
	return "theme not found: " + e.Name
}

// Built-in Themes

// RosePineTheme returns the Rose Pine theme
func RosePineTheme() *Theme {
	return &Theme{
		Name:       "rose-pine",
		Background: lipgloss.Color("#191724"),
		Card:       lipgloss.Color("#26233a"),
		Text:       lipgloss.Color("#e0def4"),
		Accent:     lipgloss.Color("#c4a6e8"),
		Secondary:  lipgloss.Color("#908cba"),
		Error:      lipgloss.Color("#eb6f92"),
		Muted:      lipgloss.Color("#6e6a86"),
		Success:    lipgloss.Color("#9ccfd8"),
	}
}

// NordTheme returns the Nord theme
func NordTheme() *Theme {
	return &Theme{
		Name:       "nord",
		Background: lipgloss.Color("#2e3440"),
		Card:       lipgloss.Color("#3b4252"),
		Text:       lipgloss.Color("#eceff4"),
		Accent:     lipgloss.Color("#88c0d0"),
		Secondary:  lipgloss.Color("#81a1c1"),
		Error:      lipgloss.Color("#bf616a"),
		Muted:      lipgloss.Color("#4c566a"),
		Success:    lipgloss.Color("#a3be8c"),
	}
}

// CatppuccinMochaTheme returns the Catppuccin Mocha theme
func CatppuccinMochaTheme() *Theme {
	return &Theme{
		Name:       "catppuccin-mocha",
		Background: lipgloss.Color("#1e1e2e"),
		Card:       lipgloss.Color("#313244"),
		Text:       lipgloss.Color("#cdd6f4"),
		Accent:     lipgloss.Color("#cba6f7"),
		Secondary:  lipgloss.Color("#a6adc8"),
		Error:      lipgloss.Color("#f38ba8"),
		Muted:      lipgloss.Color("#6c7086"),
		Success:    lipgloss.Color("#a6e3a1"),
	}
}

// DraculaTheme returns the Dracula theme
func DraculaTheme() *Theme {
	return &Theme{
		Name:       "dracula",
		Background: lipgloss.Color("#282a36"),
		Card:       lipgloss.Color("#44475a"),
		Text:       lipgloss.Color("#f8f8f2"),
		Accent:     lipgloss.Color("#bd93f9"),
		Secondary:  lipgloss.Color("#6272a4"),
		Error:      lipgloss.Color("#ff5555"),
		Muted:      lipgloss.Color("#6272a4"),
		Success:    lipgloss.Color("#50fa7b"),
	}
}

// GetBuiltInThemes returns all built-in themes
func GetBuiltInThemes() []*Theme {
	return []*Theme{
		RosePineTheme(),
		NordTheme(),
		CatppuccinMochaTheme(),
		DraculaTheme(),
	}
}

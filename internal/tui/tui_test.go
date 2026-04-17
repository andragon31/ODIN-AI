// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"testing"
	"github.com/charmbracelet/lipgloss"
)

func TestNewThemeEngine(t *testing.T) {
	engine := NewThemeEngine()

	if engine == nil {
		t.Error("NewThemeEngine() should not return nil")
	}

	if len(engine.themes) < 4 {
		t.Errorf("expected at least 4 built-in themes, got %d", len(engine.themes))
	}
}

func TestThemeEngine_GetTheme(t *testing.T) {
	engine := NewThemeEngine()

	tests := []struct {
		name     string
		expected bool
	}{
		{"rose-pine", true},
		{"nord", true},
		{"catppuccin-mocha", true},
		{"dracula", true},
		{"nonexistent", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			theme := engine.GetTheme(tc.name)
			if tc.expected && theme == nil {
				t.Errorf("expected theme %s to exist", tc.name)
			}
			if !tc.expected && theme != nil {
				t.Errorf("expected theme %s to not exist", tc.name)
			}
		})
	}
}

func TestThemeEngine_SetActiveTheme(t *testing.T) {
	engine := NewThemeEngine()

	err := engine.SetActiveTheme("nord")
	if err != nil {
		t.Errorf("SetActiveTheme() failed: %v", err)
	}

	if engine.GetActiveTheme().Name != "nord" {
		t.Errorf("expected active theme to be 'nord', got %s", engine.GetActiveTheme().Name)
	}

	// Test setting non-existent theme
	err = engine.SetActiveTheme("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent theme")
	}

	_, ok := err.(*ThemeNotFoundError)
	if !ok {
		t.Error("expected ThemeNotFoundError")
	}
}

func TestThemeEngine_ListThemes(t *testing.T) {
	engine := NewThemeEngine()
	themes := engine.ListThemes()

	if len(themes) != len(engine.themes) {
		t.Errorf("expected %d themes, got %d", len(engine.themes), len(themes))
	}
}

func TestThemeEngine_RegisterTheme(t *testing.T) {
	engine := NewThemeEngine()

	customTheme := &Theme{
		Name:       "custom",
		Background: lipgloss.Color("#000000"),
		Card:       lipgloss.Color("#111111"),
		Text:       lipgloss.Color("#ffffff"),
		Accent:     lipgloss.Color("#ff0000"),
		Secondary:  lipgloss.Color("#00ff00"),
		Error:      lipgloss.Color("#0000ff"),
		Muted:      lipgloss.Color("#888888"),
		Success:    lipgloss.Color("#00ff00"),
	}

	engine.RegisterTheme(customTheme)

	if engine.GetTheme("custom") == nil {
		t.Error("custom theme should be registered")
	}
}

func TestGetBuiltInThemes(t *testing.T) {
	themes := GetBuiltInThemes()

	if len(themes) != 4 {
		t.Errorf("expected 4 built-in themes, got %d", len(themes))
	}

	names := make(map[string]bool)
	for _, theme := range themes {
		names[theme.Name] = true
	}

	expectedNames := []string{"rose-pine", "nord", "catppuccin-mocha", "dracula"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("expected theme %s to be in built-in themes", name)
		}
	}
}

func TestNewStyles(t *testing.T) {
	theme := RosePineTheme()
	styles := NewStyles(theme)

	if styles == nil {
		t.Error("NewStyles() should not return nil")
	}

	if styles.Background.GetForeground() == nil {
		t.Error("Background style should have a horizontal foreground assigned")
	}

	if styles.Card.GetBackground() == nil {
		t.Error("Card style should have a background assigned")
	}
}

func TestListBox(t *testing.T) {
	theme := RosePineTheme()
	styles := NewStyles(theme)

	listBox := NewListBox("Test List", []string{"Item 1", "Item 2", "Item 3"}, styles)

	if listBox.Selected != 0 {
		t.Errorf("expected selected to be 0, got %d", listBox.Selected)
	}

	listBox.SelectNext()
	if listBox.Selected != 1 {
		t.Errorf("expected selected to be 1 after SelectNext(), got %d", listBox.Selected)
	}

	listBox.SelectPrevious()
	if listBox.Selected != 0 {
		t.Errorf("expected selected to be 0 after SelectPrevious(), got %d", listBox.Selected)
	}
}

func TestProgressBar(t *testing.T) {
	theme := RosePineTheme()
	styles := NewStyles(theme)

	progress := NewProgressBar("Progress", 50, 100, 20, styles)

	if progress.Current != 50 {
		t.Errorf("expected current to be 50, got %d", progress.Current)
	}

	if progress.Total != 100 {
		t.Errorf("expected total to be 100, got %d", progress.Total)
	}
}

func TestRepeatChar(t *testing.T) {
	result := RepeatChar('█', 5)
	if result != "█████" {
		t.Errorf("expected '█████', got '%s'", result)
	}

	result = RepeatChar('a', 0)
	if result != "" {
		t.Errorf("expected empty string for 0 count, got '%s'", result)
	}
}

func TestConfirmDialog(t *testing.T) {
	theme := RosePineTheme()
	styles := NewStyles(theme)

	dialog := NewConfirmDialog("Confirm?", "Are you sure?", styles)

	if dialog.Title != "Confirm?" {
		t.Errorf("expected title 'Confirm?', got '%s'", dialog.Title)
	}

	if dialog.Confirmed {
		t.Error("expected Confirmed to be false")
	}
}

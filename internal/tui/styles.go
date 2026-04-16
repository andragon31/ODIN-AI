// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all lipgloss styles for a theme
type Styles struct {
	Background  lipgloss.Style
	Card        lipgloss.Style
	Text        lipgloss.Style
	Accent      lipgloss.Style
	Secondary   lipgloss.Style
	Error       lipgloss.Style
	Muted       lipgloss.Style
	Success     lipgloss.Style
	Bold        lipgloss.Style
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Highlight   lipgloss.Style
	Border      lipgloss.Style
	BorderFocus lipgloss.Style
}

// NewStyles creates a new style set from a theme
func NewStyles(theme *Theme) *Styles {
	return &Styles{
		Background:  lipgloss.NewStyle().Background(theme.Background).Foreground(theme.Text),
		Card:        lipgloss.NewStyle().Background(theme.Card).Foreground(theme.Text).Padding(1, 2),
		Text:        lipgloss.NewStyle().Foreground(theme.Text),
		Accent:      lipgloss.NewStyle().Foreground(theme.Accent),
		Secondary:   lipgloss.NewStyle().Foreground(theme.Secondary),
		Error:       lipgloss.NewStyle().Foreground(theme.Error),
		Muted:       lipgloss.NewStyle().Foreground(theme.Muted),
		Success:     lipgloss.NewStyle().Foreground(theme.Success),
		Bold:        lipgloss.NewStyle().Bold(true),
		Title:       lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).MarginBottom(1),
		Subtitle:    lipgloss.NewStyle().Foreground(theme.Secondary),
		Highlight:   lipgloss.NewStyle().Foreground(theme.Accent).Background(theme.Card),
		Border:      lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.Muted),
		BorderFocus: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(theme.Accent),
	}
}

// RenderCard renders a card with the given content
func (s *Styles) RenderCard(title, content string) string {
	titleStyle := s.Title.Render(title)
	contentStyle := s.Text.Render(content)
	borderStyle := s.Border

	return borderStyle.Render(titleStyle + "\n" + contentStyle)
}

// RenderList renders a styled list
func (s *Styles) RenderList(items []string, numbered bool) string {
	result := ""
	for i, item := range items {
		prefix := "  • "
		if numbered {
			prefix = lipgloss.NewStyle().Faint(true).Render("%d. ", i+1)
		}
		result += s.Text.Render(prefix + item + "\n")
	}
	return result
}

// RenderStatusIndicator renders a status indicator
func (s *Styles) RenderStatusIndicator(status string) string {
	var color lipgloss.Color
	switch status {
	case "success", "pass", "healthy":
		color = s.Success.GetForeground()
	case "error", "fail", "broken":
		color = s.Error.GetForeground()
	case "warning", "warn":
		color = lipgloss.Color("#f9e2af") // Warning yellow
	default:
		color = s.Muted.GetForeground()
	}

	dot := lipgloss.NewStyle().
		Background(color).
		Width(1).
		Height(1).
		MarginRight(1).
		Render(" ")

	return dot + s.Text.Render(status)
}

// RenderProgressBar renders a progress bar
func (s *Styles) RenderProgressBar(current, total int, width int) string {
	if total <= 0 {
		total = 1
	}

	ratio := float64(current) / float64(total)
	filled := int(ratio * float64(width))
	empty := width - filled

	filledStr := lipgloss.NewStyle().
		Background(s.Accent.GetForeground()).
		Render(RepeatChar('█', filled))

	emptyStr := lipgloss.NewStyle().
		Foreground(s.Muted.GetForeground()).
		Render(RepeatChar('░', empty))

	percentage := lipgloss.NewStyle().
		Faint(true).
		Render(" %d%%", int(ratio*100))

	return filledStr + emptyStr + percentage
}

// RepeatChar creates a string of repeated characters
func RepeatChar(char rune, count int) string {
	if count <= 0 {
		return ""
	}
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return string(result)
}

// RenderHeader renders a header with decorative borders
func (s *Styles) RenderHeader(title string) string {
	width := 50
	border := RepeatChar('═', width)

	top := lipgloss.NewStyle().
		Foreground(s.Accent.GetForeground()).
		Render("╔" + border + "╗")

	bottom := lipgloss.NewStyle().
		Foreground(s.Accent.GetForeground()).
		Render("╚" + border + "╝")

	middle := lipgloss.NewStyle().
		Foreground(s.Accent.GetForeground()).
		Render("║")

	centeredTitle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(title)

	return top + "\n" + middle + centeredTitle + middle + "\n" + bottom
}

// RenderTableCell renders a table cell with alignment
func (s *Styles) RenderTableCell(content string, width int, align lipgloss.Alignment) string {
	style := lipgloss.NewStyle().
		Width(width).
		Align(align).
		Foreground(s.Text.GetForeground())

	return style.Render(content)
}

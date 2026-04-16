// Package tui provides Völva - the interface engine for ODIN
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Component represents a reusable TUI component
type Component interface {
	Render() string
}

// Panel is a container component
type Panel struct {
	Title   string
	Content string
	Width   int
	Height  int
	Styles  *Styles
	Focused bool
}

// NewPanel creates a new panel
func NewPanel(title string, content string, width int, height int, styles *Styles) *Panel {
	return &Panel{
		Title:   title,
		Content: content,
		Width:   width,
		Height:  height,
		Styles:  styles,
		Focused: false,
	}
}

// Render renders the panel
func (p *Panel) Render() string {
	border := p.Styles.Border
	if p.Focused {
		border = p.Styles.BorderFocus
	}

	content := border.Render(p.Content)
	return content
}

// ListBox is a selectable list component
type ListBox struct {
	Title    string
	Items    []string
	Selected int
	Styles   *Styles
}

// NewListBox creates a new list box
func NewListBox(title string, items []string, styles *Styles) *ListBox {
	return &ListBox{
		Title:    title,
		Items:    items,
		Selected: 0,
		Styles:   styles,
	}
}

// Render renders the list box
func (l *ListBox) Render() string {
	var builder strings.Builder

	builder.WriteString(l.Styles.Title.Render(l.Title))
	builder.WriteString("\n")

	for i, item := range l.Items {
		prefix := "  "
		if i == l.Selected {
			prefix = " ● "
			builder.WriteString(l.Styles.Accent.Render(prefix + item + "\n"))
		} else {
			builder.WriteString(l.Styles.Muted.Render(prefix + item + "\n"))
		}
	}

	return builder.String()
}

// SelectNext moves selection to next item
func (l *ListBox) SelectNext() {
	if l.Selected < len(l.Items)-1 {
		l.Selected++
	}
}

// SelectPrevious moves selection to previous item
func (l *ListBox) SelectPrevious() {
	if l.Selected > 0 {
		l.Selected--
	}
}

// ProgressBar is a component for displaying progress
type ProgressBar struct {
	Label   string
	Current int
	Total   int
	Width   int
	Styles  *Styles
}

// NewProgressBar creates a new progress bar
func NewProgressBar(label string, current int, total int, width int, styles *Styles) *ProgressBar {
	return &ProgressBar{
		Label:   label,
		Current: current,
		Total:   total,
		Width:   width,
		Styles:  styles,
	}
}

// Render renders the progress bar
func (p *ProgressBar) Render() string {
	label := p.Styles.Text.Render(p.Label + ": ")
	bar := p.Styles.RenderProgressBar(p.Current, p.Total, p.Width)
	return label + bar
}

// StatusBadge shows a status indicator
type StatusBadge struct {
	Label  string
	Status string
	Styles *Styles
}

// NewStatusBadge creates a new status badge
func NewStatusBadge(label string, status string, styles *Styles) *StatusBadge {
	return &StatusBadge{
		Label:  label,
		Status: status,
		Styles: styles,
	}
}

// Render renders the status badge
func (s *StatusBadge) Render() string {
	indicator := s.Styles.RenderStatusIndicator(s.Status)
	label := s.Styles.Text.Render(" " + s.Label + ": ")
	return label + indicator
}

// Header renders a page header
type Header struct {
	Title    string
	Subtitle string
	Styles   *Styles
}

// NewHeader creates a new header
func NewHeader(title string, subtitle string, styles *Styles) *Header {
	return &Header{
		Title:    title,
		Subtitle: subtitle,
		Styles:   styles,
	}
}

// Render renders the header
func (h *Header) Render() string {
	return h.Styles.RenderHeader(h.Title)
}

// ModelToString converts a tea.Model to string for rendering
func ModelToString(m tea.Model) string {
	return m.View()
}

// Spacer creates vertical spacing
func Spacer(height int) string {
	return strings.Repeat("\n", height)
}

// RenderThemePreview renders a theme preview card
func RenderThemePreview(theme *Theme) string {
	styles := NewStyles(theme)

	card := styles.Card.Copy().
		Width(40).
		Render(
			styles.Title.Render(theme.Name) + "\n" +
				styles.Text.Render("Background: ") + string(theme.Background) + "\n" +
				styles.Text.Render("Card:       ") + string(theme.Card) + "\n" +
				styles.Text.Render("Text:       ") + string(theme.Text) + "\n" +
				styles.Text.Render("Accent:     ") + string(theme.Accent) + "\n" +
				styles.Text.Render("Secondary:  ") + string(theme.Secondary) + "\n" +
				styles.Text.Render("Error:      ") + string(theme.Error) + "\n" +
				styles.Text.Render("Success:    ") + string(theme.Success),
		)

	return card
}

// RenderThemeComparison renders multiple themes side by side
func RenderThemeComparison(themes []*Theme) string {
	if len(themes) == 0 {
		return "No themes to display"
	}

	var results []string
	for _, theme := range themes {
		results = append(results, RenderThemePreview(theme))
	}

	// Stack themes horizontally using lipgloss
	return lipgloss.JoinHorizontal(lipgloss.Top, results...)
}

// ConfirmDialog is a simple confirmation dialog
type ConfirmDialog struct {
	Title     string
	Message   string
	Confirmed bool
	Styles    *Styles
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title string, message string, styles *Styles) *ConfirmDialog {
	return &ConfirmDialog{
		Title:     title,
		Message:   message,
		Confirmed: false,
		Styles:    styles,
	}
}

// Render renders the confirmation dialog
func (c *ConfirmDialog) Render() string {
	content := fmt.Sprintf("%s\n\n%s\n\n%s [%s] %s [%s]",
		c.Title,
		c.Message,
		c.Styles.Success.Render("[y] Yes"),
		c.Styles.Muted.Render("n"),
		c.Styles.Muted.Render("No"),
		c.Styles.Muted.Render("n"),
	)

	return c.Styles.Border.Render(content)
}

// InputPrompt is a simple input prompt
type InputPrompt struct {
	Label   string
	Value   string
	Styles  *Styles
	Focused bool
}

// NewInputPrompt creates a new input prompt
func NewInputPrompt(label string, styles *Styles) *InputPrompt {
	return &InputPrompt{
		Label:   label,
		Value:   "",
		Styles:  styles,
		Focused: true,
	}
}

// Render renders the input prompt
func (i *InputPrompt) Render() string {
	border := i.Styles.Border
	if i.Focused {
		border = i.Styles.BorderFocus
	}

	content := i.Styles.Text.Render(i.Label+": ") +
		i.Styles.Accent.Render(i.Value+"_")

	return border.Render(content)
}

// Table represents a data table component
type Table struct {
	Headers []string
	Rows    [][]string
	Styles  *Styles
	Widths  []int
}

// NewTable creates a new table
func NewTable(headers []string, styles *Styles) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h) + 2
	}

	return &Table{
		Headers: headers,
		Rows:    [][]string{},
		Styles:  styles,
		Widths:  widths,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

// Render renders the table
func (t *Table) Render() string {
	var builder strings.Builder

	// Render headers
	for i, header := range t.Headers {
		cell := t.Styles.RenderTableCell(header, t.Widths[i], lipgloss.Left)
		builder.WriteString(t.Styles.Accent.Render(cell))
	}
	builder.WriteString("\n")

	// Render separator
	for _, width := range t.Widths {
		builder.WriteString(strings.Repeat("─", width))
	}
	builder.WriteString("\n")

	// Render rows
	for _, row := range t.Rows {
		for i, cell := range row {
			width := t.Widths[i]
			if i >= len(width) {
				width = 10
			}
			styledCell := t.Styles.RenderTableCell(cell, width, lipgloss.Left)
			builder.WriteString(t.Styles.Text.Render(styledCell))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

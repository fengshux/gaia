// Package cli provides the CLI interface
package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Renderer handles output rendering
type Renderer struct {
	styles Styles
}

// Styles contains styling for different output types
type Styles struct {
	Header    lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Info      lipgloss.Style
	Highlight lipgloss.Style
	Code      lipgloss.Style
	Tool      lipgloss.Style
	User      lipgloss.Style
	Assistant lipgloss.Style
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		styles: Styles{
			Header: lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Bold(true),
			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")),
			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")),
			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")),
			Info: lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")),
			Highlight: lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true),
			Code: lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Background(lipgloss.Color("235")),
			Tool: lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")),
			User: lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")),
			Assistant: lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")),
		},
	}
}

// Header renders a header
func (r *Renderer) Header(text string) string {
	return r.styles.Header.Render(text)
}

// Success renders a success message
func (r *Renderer) Success(text string) string {
	return r.styles.Success.Render("✓ " + text)
}

// Error renders an error message
func (r *Renderer) Error(text string) string {
	return r.styles.Error.Render("✗ " + text)
}

// Warning renders a warning message
func (r *Renderer) Warning(text string) string {
	return r.styles.Warning.Render("⚠ " + text)
}

// Info renders an info message
func (r *Renderer) Info(text string) string {
	return r.styles.Info.Render("ℹ " + text)
}

// Highlight renders highlighted text
func (r *Renderer) Highlight(text string) string {
	return r.styles.Highlight.Render(text)
}

// Code renders code
func (r *Renderer) Code(code string) string {
	return r.styles.Code.Render(code)
}

// Tool renders a tool name
func (r *Renderer) Tool(name string) string {
	return r.styles.Tool.Render("[" + name + "]")
}

// User renders user message
func (r *Renderer) User(text string) string {
	return r.styles.User.Render("You: ") + text
}

// Assistant renders assistant message
func (r *Renderer) Assistant(text string) string {
	return r.styles.Assistant.Render("Assistant: ") + text
}

// Table renders a simple table
func (r *Renderer) Table(headers []string, rows [][]string) string {
	var sb strings.Builder

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		sb.WriteString(r.styles.Header.Render(pad(h, widths[i])))
		sb.WriteString("  ")
	}
	sb.WriteString("\n")

	// Print separator
	for i, w := range widths {
		sb.WriteString(strings.Repeat("-", w))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			sb.WriteString(pad(cell, widths[i]))
			sb.WriteString("  ")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Box renders text in a box
func (r *Renderer) Box(title, content string) string {
	width := 60
	if len(title)+4 > width {
		width = len(title) + 4
	}

	var sb strings.Builder
	sb.WriteString("┌" + strings.Repeat("─", width-2) + "┐\n")

	if title != "" {
		sb.WriteString("│ " + r.styles.Header.Render(pad(title, width-4)) + " │\n")
		sb.WriteString("├" + strings.Repeat("─", width-2) + "┤\n")
	}

	for _, line := range strings.Split(content, "\n") {
		sb.WriteString("│ " + pad(line, width-4) + " │\n")
	}

	sb.WriteString("└" + strings.Repeat("─", width-2) + "┘\n")

	return sb.String()
}

// ProgressBar renders a progress bar
func (r *Renderer) ProgressBar(current, total int, label string) string {
	width := 40
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s [%s] %.0f%%", label, bar, percent*100)
}

// Helper function to pad strings
func pad(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

package tui

import "github.com/charmbracelet/lipgloss"

func renderComponentHeader(title string, border lipgloss.Border) string {
	style := lipgloss.NewStyle().Border(border, false, true, true, false).Padding(0, 1)
	if title == "" {
		title = "Untitled"
	}
	return style.Render(title)
}


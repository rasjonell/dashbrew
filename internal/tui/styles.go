package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *model) getBorderStyle() (normal lipgloss.Style, focused lipgloss.Style, border lipgloss.Border) {
	border = lipgloss.RoundedBorder()
	normalColor := getColor(m.cfg.Style.BorderColor, lipgloss.Color("#d9ebf9"))
	focusedColor := getColor(m.cfg.Style.FocusedBorderColor, lipgloss.Color("#00ff00"))

	if m.cfg.Style == nil {
		style := lipgloss.NewStyle().Border(border)
		return style, style.BorderForeground(focusedColor), border
	}

	switch m.cfg.Style.BorderType {
	case "rounded":
		border = lipgloss.RoundedBorder()
	case "thicc":
		border = lipgloss.ThickBorder()
	case "double":
		border = lipgloss.DoubleBorder()
	case "hidden":
		border = lipgloss.HiddenBorder()
	case "normal":
		border = lipgloss.NormalBorder()
	case "md":
		border = lipgloss.MarkdownBorder()
	case "ascii":
		border = lipgloss.ASCIIBorder()
	case "block":
		border = lipgloss.BlockBorder()
	}

	configColor := m.cfg.Style.FocusedBorderColor
	if len(configColor) == 7 && configColor[0] == '#' {
		focusedColor = lipgloss.Color(configColor)
	}

	style := lipgloss.NewStyle().Border(border).BorderForeground(normalColor)
	return style, style.BorderForeground(focusedColor), border
}

func getColor(configColor string, defaultColor lipgloss.Color) lipgloss.Color {
	if len(configColor) == 7 && configColor[0] == '#' {
		return lipgloss.Color(configColor)
	}
	return defaultColor
}

func wrapContent(content string, width int) string {
	return lipgloss.
		NewStyle().
		Width(width).
		Render(content)
}

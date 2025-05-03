package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
)

func GetBorderStyle(styles *config.StyleConfig) (normal lipgloss.Style, focused lipgloss.Style, border lipgloss.Border) {
	border = lipgloss.RoundedBorder()
	normalColor := GetColor(styles.BorderColor, lipgloss.Color("#d9ebf9"))
	focusedColor := GetColor(styles.FocusedBorderColor, lipgloss.Color("#00ff00"))

	if styles == nil {
		style := lipgloss.NewStyle().Border(border)
		return style, style.BorderForeground(focusedColor), border
	}

	switch styles.BorderType {
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

	configColor := styles.FocusedBorderColor
	if len(configColor) == 7 && configColor[0] == '#' {
		focusedColor = lipgloss.Color(configColor)
	}

	style := lipgloss.NewStyle().Border(border).BorderForeground(normalColor)
	return style, style.BorderForeground(focusedColor), border
}

func GetColor(configColor string, defaultColor lipgloss.Color) lipgloss.Color {
	if len(configColor) == 7 && configColor[0] == '#' {
		return lipgloss.Color(configColor)
	}
	return defaultColor
}

func WrapContent(content string, width int) string {
	return lipgloss.
		NewStyle().
		Width(width).
		Render(content)
}

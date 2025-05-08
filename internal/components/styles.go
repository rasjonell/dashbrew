package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
)

func GetBorderStyle(borderStyles *config.BorderStyleConfig) (normal lipgloss.Style, focused lipgloss.Style, border lipgloss.Border) {
	border = lipgloss.ThickBorder()
	normalColor := GetColor(borderStyles.Color, BrightenColor(lipgloss.Color("default"), 100))
	focusedColor := GetColor(borderStyles.FocusedColor, lipgloss.Color("default"))

	if borderStyles == nil {
		style := lipgloss.NewStyle().Border(border)
		return style, style.BorderForeground(focusedColor), border
	}

	switch borderStyles.Type {
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

	configColor := borderStyles.FocusedColor
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

func BrightenColor(color lipgloss.Color, percent float64) lipgloss.Color {
	colorStr := string(color)

	if colorStr == "" {
		return lipgloss.Color("15")
	}

	if strings.HasPrefix(colorStr, "#") {
		if len(colorStr) == 7 {
			r, _ := strconv.ParseInt(colorStr[1:3], 16, 0)
			g, _ := strconv.ParseInt(colorStr[3:5], 16, 0)
			b, _ := strconv.ParseInt(colorStr[5:7], 16, 0)

			factor := 1.0 + (percent / 100.0)

			rBright := int(math.Min(float64(r)*factor, 255))
			gBright := int(math.Min(float64(g)*factor, 255))
			bBright := int(math.Min(float64(b)*factor, 255))

			return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", rBright, gBright, bBright))
		}

		if len(colorStr) == 4 {
			r, _ := strconv.ParseInt(string(colorStr[1])+string(colorStr[1]), 16, 0)
			g, _ := strconv.ParseInt(string(colorStr[2])+string(colorStr[2]), 16, 0)
			b, _ := strconv.ParseInt(string(colorStr[3])+string(colorStr[3]), 16, 0)

			factor := 1.0 + (percent / 100.0)

			rBright := int(math.Min(float64(r)*factor, 255))
			gBright := int(math.Min(float64(g)*factor, 255))
			bBright := int(math.Min(float64(b)*factor, 255))

			return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", rBright, gBright, bBright))
		}
	}

	if len(colorStr) == 1 {
		ansiNum, err := strconv.Atoi(colorStr)
		if err == nil && ansiNum >= 0 && ansiNum <= 7 {
			return lipgloss.Color(fmt.Sprintf("%d", ansiNum+8))
		}
	}

	brightColorMap := map[string]string{
		"black":   "8",
		"red":     "9",
		"green":   "10",
		"yellow":  "11",
		"blue":    "12",
		"magenta": "13",
		"cyan":    "14",
		"white":   "15",
	}

	if brighter, exists := brightColorMap[strings.ToLower(colorStr)]; exists {
		return lipgloss.Color(brighter)
	}

	return lipgloss.Color("15")
}

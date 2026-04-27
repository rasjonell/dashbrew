package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"
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

// blendHexColors linearly interpolates between two color strings (hex or
// any lipgloss.Color value) by `t` in [0, 1] using LAB blending for visually
// uniform gradients. Returns `fallback` if either color fails to parse.
func blendHexColors(a, b string, t float64, fallback lipgloss.Color) lipgloss.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	ca, errA := colorful.Hex(a)
	cb, errB := colorful.Hex(b)
	if errA != nil || errB != nil {
		return fallback
	}
	mixed := ca.BlendLab(cb, t)
	return lipgloss.Color(mixed.Hex())
}

// PaletteColor returns the lipgloss.Color for a named role, falling back to
// `fallback` when the role is unset or the active style has no palette.
func PaletteColor(styles *config.StyleConfig, role string, fallback lipgloss.Color) lipgloss.Color {
	if styles == nil || styles.Palette == nil {
		return fallback
	}
	hex := ""
	switch role {
	case "background":
		hex = styles.Palette.Background
	case "foreground":
		hex = styles.Palette.Foreground
	case "dim":
		hex = styles.Palette.Dim
	case "accent":
		hex = styles.Palette.Accent
	case "success":
		hex = styles.Palette.Success
	case "warning":
		hex = styles.Palette.Warning
	case "error":
		hex = styles.Palette.Error
	case "border":
		hex = styles.Palette.Border
	case "borderFocused":
		hex = styles.Palette.BorderFocused
	}
	return GetColor(hex, fallback)
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

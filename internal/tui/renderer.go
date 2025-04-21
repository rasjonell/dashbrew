package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
)

const (
	borderSize = 1
)

var (
	defaultBorder = lipgloss.RoundedBorder()
	defaultStyle  = lipgloss.NewStyle().Border(defaultBorder)

	focusedBorder = defaultBorder
	focusedStyle  = lipgloss.NewStyle().Border(focusedBorder).BorderForeground(lipgloss.Color("#00ff00"))
)

func renderNode(
	node *config.LayoutNode,
	width, height int,
	renderComponent func(*config.Component, int, int) string,
	focusedComponentId string,
) string {
	switch node.Type {
	case "component":
		if node.Component == nil {
			return ""
		}
		w, h := calcWidthHeight(width, height)

		style := defaultStyle
		if componentId(node.Component) == focusedComponentId {
			style = focusedStyle
		}

		return style.
			Width(w).
			Height(h).
			Render(renderComponent(node.Component, w, h))

	case "container":
		if len(node.Children) == 0 {
			return ""
		}

		totalFlex := 0
		for _, child := range node.Children {
			totalFlex += getFlex(child)
		}

		var rendered []string
		for _, child := range node.Children {
			flex := getFlex(child)
			var childWidth, childHeight int
			if node.Direction == "row" {
				childWidth = width * flex / totalFlex
				childHeight = height
			} else {
				childWidth = width
				childHeight = height * flex / totalFlex
			}

			rendered = append(rendered, renderNode(child, childWidth, childHeight, renderComponent, focusedComponentId))
		}

		if node.Direction == "row" {
			return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
		}

		return lipgloss.JoinVertical(lipgloss.Left, rendered...)

	default:
		return ""
	}
}

func (m *model) getComponentContent(component *config.Component, width, height int) string {
	switch component.Type {
	case "text":
		return m.renderTextComponent(component, width, height)
	default:
		// Unknown type
		title := component.Title
		if title == "" {
			title = "[Untitled]"
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("[unknown type: %s]\n%s", component.Type, title),
		)
	}
}

func calcWidthHeight(w, h int) (int, int) {
	innerWidth := w - 2*(borderSize)
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := h - 2*(borderSize)
	if innerHeight < 1 {
		innerHeight = 1
	}

	return innerWidth, innerHeight
}

func evenWidthHeight(w, h int) (newW, newH int) {
	if w%2 != 0 {
		w--
	}
	if h%2 != 0 {
		h--
	}
	return w, h
}

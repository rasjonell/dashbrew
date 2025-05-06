package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/components"
	"github.com/rasjonell/dashbrew/internal/config"
)

func (m *model) buildComponentMap(node *config.LayoutNode) {
	if node == nil {
		return
	}

	switch node.Type {
	case "component":
		if node.Component != nil {
			comp := components.NewComponent(node.Component, m.cfg.Style)
			m.components[comp.ID()] = comp
		}
	case "container":
		for _, child := range node.Children {
			m.buildComponentMap(child)
		}
	}
}

func (m *model) renderNode(
	node *config.LayoutNode,
	width, height int,
	focusedComponentId string,
) string {
	if node == nil {
		return lipgloss.NewStyle().Width(width).Height(height).Render("")
	}

	switch node.Type {
	case "component":
		if node.Component == nil {
			return lipgloss.NewStyle().Width(width).Height(height).Render("[Error: nil componnet config]")
		}
		id := components.ComponentId(node.Component)
		comp, exists := m.components[id]
		if !exists {
			return lipgloss.NewStyle().Width(width).Height(height).Render("[Error: component %s not found]", id)
		}

		isFocused := id == focusedComponentId

		if isFocused && m.isAdding && comp.SupportsAdd() {
			_, focusedStyle, _ := components.GetBorderStyle(m.cfg.Style)
			borderStyle := focusedStyle

			addInput := comp.GetAddInput()
			prompt := fmt.Sprintf("New ToDo: %s_", addInput)

			w, h := components.CalcWidthHeight(width, height)
			addOverlayContent := lipgloss.Place(w, h,
				lipgloss.Center, lipgloss.Center,
				prompt,
			)

			return borderStyle.Width(w).Height(h).Render(addOverlayContent)

		}

		return comp.View(width, height, isFocused)

	case "container":
		if len(node.Children) == 0 {
			return lipgloss.NewStyle().Width(width).Height(height).Render("")
		}

		totalFlex := 0
		for _, child := range node.Children {
			totalFlex += components.GetFlex(child.Flex)
		}

		offset := 0
		numChildren := len(node.Children)

		var rendered []string

		for i, child := range node.Children {
			flex := components.GetFlex(child.Flex)
			isLast := i == numChildren-1

			var childWidth, childHeight int
			if node.Direction == "row" {
				childHeight = height
				if isLast {
					childWidth = width - offset
				} else {
					childWidth = width * flex / totalFlex
				}
				offset += childWidth
			} else {
				childWidth = width
				if isLast {
					childHeight = height - offset
				} else {
					childHeight = height * flex / totalFlex
				}
				offset += childHeight
			}

			rendered = append(rendered, m.renderNode(child, childWidth, childHeight, focusedComponentId))
		}

		if node.Direction == "row" {
			return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
		}

		return lipgloss.JoinVertical(lipgloss.Left, rendered...)

	default:
		w, h := components.CalcWidthHeight(width, height)
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#ff0000")).
			Render(fmt.Sprintf("[Unknown Layout Type: %s]", node.Type))
	}
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

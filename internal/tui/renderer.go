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
			_, focusedStyle, _ := components.GetBorderStyle(m.cfg.Style.Border)
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

		view := comp.View(width, height, isFocused)
		if !isFocused && m.cfg.Style != nil && m.cfg.Style.FocusMode == "highlight" {
			view = lipgloss.NewStyle().Faint(true).Render(view)
		}
		return view

	case "container":
		if len(node.Children) == 0 {
			return lipgloss.NewStyle().Width(width).Height(height).Render("")
		}

		totalFlex := 0
		for _, child := range node.Children {
			totalFlex += components.GetFlex(child.Flex)
		}

		numChildren := len(node.Children)
		gap := layoutGap(m.cfg.Style)
		totalGap := gap * max(0, numChildren-1)

		availableW, availableH := width, height
		if node.Direction == "row" {
			availableW = max(0, width-totalGap)
		} else {
			availableH = max(0, height-totalGap)
		}

		offset := 0
		var rendered []string

		for i, child := range node.Children {
			flex := components.GetFlex(child.Flex)
			isLast := i == numChildren-1

			var childWidth, childHeight int
			if node.Direction == "row" {
				childHeight = height
				if isLast {
					childWidth = availableW - offset
				} else {
					childWidth = availableW * flex / totalFlex
				}
				offset += childWidth
			} else {
				childWidth = width
				if isLast {
					childHeight = availableH - offset
				} else {
					childHeight = availableH * flex / totalFlex
				}
				offset += childHeight
			}

			rendered = append(rendered, m.renderNode(child, childWidth, childHeight, focusedComponentId))
		}

		if node.Direction == "row" {
			parts := interleaveRow(rendered, gap, height)
			return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
		}
		parts := interleaveColumn(rendered, gap, width)
		return lipgloss.JoinVertical(lipgloss.Left, parts...)

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

func layoutGap(s *config.StyleConfig) int {
	if s == nil || s.Layout == nil {
		return 0
	}
	g := s.Layout.Gap
	if g < 0 {
		return 0
	}
	return g
}

// interleaveRow inserts a `gap`-wide whitespace column between each rendered
// child of a row container.
func interleaveRow(children []string, gap, height int) []string {
	if gap <= 0 || len(children) < 2 {
		return children
	}
	spacer := lipgloss.NewStyle().Width(gap).Height(height).Render("")
	out := make([]string, 0, len(children)*2-1)
	for i, c := range children {
		if i > 0 {
			out = append(out, spacer)
		}
		out = append(out, c)
	}
	return out
}

// interleaveColumn inserts a `gap`-tall whitespace row between each rendered
// child of a column container.
func interleaveColumn(children []string, gap, width int) []string {
	if gap <= 0 || len(children) < 2 {
		return children
	}
	spacer := lipgloss.NewStyle().Width(width).Height(gap).Render("")
	out := make([]string, 0, len(children)*2-1)
	for i, c := range children {
		if i > 0 {
			out = append(out, spacer)
		}
		out = append(out, c)
	}
	return out
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

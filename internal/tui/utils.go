package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/rasjonell/dashbrew/internal/config"
)

const nbsp = 0xA0

func (m *model) buildComponentMap(node *config.LayoutNode) {
	if node.Type == "component" && node.Component != nil {
		id := componentId(node.Component)
		m.componentMap[id] = node.Component
		if node.Component.Type == "text" {
			m.viewports[id] = viewport.New(0, 0)
		}
	}
	for _, child := range node.Children {
		m.buildComponentMap(child)
	}
}

func componentId(comp *config.Component) string {
	id := comp.ID
	if id == "" {
		id = fmt.Sprintf("%p", comp)
	}

	return id
}

func getFlex(node *config.LayoutNode) int {
	flex := node.Flex
	if flex == 0 {
		flex = 1
	}

	return flex
}

package tui

import (
	"fmt"
	"slices"

	"github.com/rasjonell/dashbrew/internal/config"
)

var additionSupportedComponents = []string{"todo"}

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

func supportsAddition(compType string) bool {
	return slices.Index(
		additionSupportedComponents,
		compType,
	) > -1
}

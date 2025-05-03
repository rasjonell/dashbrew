package tui

import (
	"slices"

	"github.com/rasjonell/dashbrew/internal/config"
)

var additionSupportedComponents = []string{"todo"}

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

package components

import (
	"fmt"

	"github.com/rasjonell/dashbrew/internal/config"
)

const (
	borderSize = 1
)

func ComponentId(comp *config.Component) string {
	id := comp.ID
	if id == "" {
		id = fmt.Sprintf("%p", comp)
	}

	return id
}

func CalcWidthHeight(w, h int) (int, int) {
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

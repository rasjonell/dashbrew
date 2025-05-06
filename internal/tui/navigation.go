package tui

import (
	"math"

	"github.com/rasjonell/dashbrew/internal/components"
	"github.com/rasjonell/dashbrew/internal/config"
)

type navigationMap struct {
	Up    string
	Down  string
	Left  string
	Right string
}

type boundingBox struct {
	X  int
	Y  int
	W  int
	H  int
	ID string
}

func (m *model) handleResize(msgWidth, msgHeight int) {
	w, h := evenWidthHeight(msgWidth, msgHeight)
	if w == m.width && h == m.height && m.ready {
		return
	}

	m.width = w
	m.height = h
	m.ready = false

	if m.cfg == nil || m.cfg.Layout == nil || len(m.components) == 0 {
		m.ready = true
		return
	}

	newBoxes := make(map[string]*boundingBox)
	calculateBoundingBoxes(m.cfg.Layout, 0, 0, w, h, newBoxes)
	m.componentBoxes = newBoxes

	m.navMap = calculateNavigationMap(m.componentBoxes)
	m.ready = true

	if _, exists := m.componentBoxes[m.focusedComponentId]; !exists {
		m.focusedComponentId = findFirstComponent(m.cfg.Layout)
	}
}

func (m *model) focusClicked(x, y int) {
	for id, box := range m.componentBoxes {
		if x >= box.X && x <= box.X+box.W &&
			y >= box.Y && y <= box.Y+box.H {
			m.focusedComponentId = id
		}
	}
}

func (m *model) tryFocus(targetID string) {
	if targetComp, ok := m.components[targetID]; ok && targetComp.IsFocusable() {
		m.focusedComponentId = targetID
	}
}

func calculateBoundingBoxes(
	node *config.LayoutNode,
	x, y, width, height int,
	boxes map[string]*boundingBox,
) {
	if node == nil || width <= 0 || height <= 0 {
		return
	}

	w, h := width, height

	switch node.Type {
	case "component":
		if node.Component != nil {
			id := components.ComponentId(node.Component)
			boxes[id] = &boundingBox{
				X: x, Y: y, W: w, H: h, ID: id,
			}
		}
	case "container":
		if len(node.Children) == 0 {
			return
		}

		totalFlex := 0
		for _, child := range node.Children {
			totalFlex += components.GetFlex(child.Flex)
		}

		offset := 0
		numChildren := len(node.Children)
		for i, child := range node.Children {
			flex := components.GetFlex(child.Flex)

			isLast := i == numChildren-1

			var childX, childY, childW, childH int
			if node.Direction == "row" {
				childX = x + offset
				childY = y
				childH = h
				if isLast {
					childW = w - offset
				} else {
					childW = w * flex / totalFlex
				}
				offset += childW
			} else {
				childX = x
				childY = y + offset
				childW = w

				if isLast {
					childH = h - offset
				} else {
					childH = h * flex / totalFlex
				}
				offset += childH
			}

			calculateBoundingBoxes(child, childX, childY, childW, childH, boxes)
		}
	}
}

func calculateNavigationMap(boxes map[string]*boundingBox) map[string]*navigationMap {
	navMap := make(map[string]*navigationMap)

	for sourceID, sourceBox := range boxes {
		currentNav := &navigationMap{}

		minDistUp := math.MaxFloat64
		minDistDown := math.MaxFloat64
		minDistLeft := math.MaxFloat64
		minDistRight := math.MaxFloat64

		sourceTopX := float64(sourceBox.X)
		sourceTopY := float64(sourceBox.Y)

		for targetID, targetBox := range boxes {
			if sourceID == targetID {
				continue
			}

			targetTopX := float64(targetBox.X)
			targetTopY := float64(targetBox.Y)

			vOverlap := max(sourceBox.Y, targetBox.Y) < min(sourceBox.Y+sourceBox.H, targetBox.Y+targetBox.H)
			hOverlap := max(sourceBox.X, targetBox.X) < min(sourceBox.X+sourceBox.W, targetBox.X+targetBox.W)

			if vOverlap && targetBox.X >= sourceBox.X+sourceBox.W {
				dist := math.Abs(targetTopX - sourceTopX)
				if dist < minDistRight {
					minDistRight = dist
					currentNav.Right = targetID
				}
			}

			if vOverlap && targetBox.X+targetBox.W <= sourceBox.X {
				dist := math.Abs(targetTopX - sourceTopX)
				if dist < minDistLeft {
					minDistLeft = dist
					currentNav.Left = targetID
				}
			}

			if hOverlap && targetBox.Y >= sourceBox.Y+sourceBox.H {
				dist := math.Abs(targetTopY - sourceTopY)
				if dist < minDistDown {
					minDistDown = dist
					currentNav.Down = targetID
				}
			}

			if hOverlap && targetBox.Y+targetBox.H <= sourceBox.Y {
				dist := math.Abs(targetTopY - sourceTopY)
				if dist < minDistUp {
					minDistUp = dist
					currentNav.Up = targetID
				}
			}
		}
		navMap[sourceID] = currentNav
	}

	return navMap
}

func findFirstComponent(node *config.LayoutNode) string {
	if node == nil {
		return ""
	}
	if node.Type == "component" && node.Component != nil {
		return components.ComponentId(node.Component)
	}

	for _, child := range node.Children {
		if id := findFirstComponent(child); id != "" {
			return id
		}
	}

	return ""
}

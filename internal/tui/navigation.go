package tui

import (
	"math"

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
	m.width = w
	m.height = h

	if m.cfg == nil || m.cfg.Layout == nil {
		return
	}

	m.componentBoxes = make(map[string]*boundingBox)
	calculateBoundingBoxes(m.cfg.Layout, 0, 0, w, h, m.componentBoxes)
	m.navMap = calculateNavigationMap(m.componentBoxes)
	m.updateViewports()

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

func calculateBoundingBoxes(
	node *config.LayoutNode,
	x, y, width, height int,
	boxes map[string]*boundingBox,
) {
	if node == nil {
		return
	}

	w, h := width, height

	switch node.Type {
	case "component":
		if node.Component != nil {
			id := componentId(node.Component)
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
			totalFlex += getFlex(child)
		}

		offset := 0
		for _, child := range node.Children {
			flex := getFlex(child)

			var childX, childY, childW, childH int
			if node.Direction == "row" {
				childX = x + offset
				childY = y
				childH = h
				childW = w * flex / totalFlex
				offset += childW
			} else {
				childX = x
				childY = y + offset
				childW = w
				childH = h * flex / totalFlex
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

		sourceCenterX := float64(sourceBox.X) + float64(sourceBox.W)/2.0
		sourceCenterY := float64(sourceBox.Y) + float64(sourceBox.H)/2.0

		for targetID, targetBox := range boxes {
			if sourceID == targetID {
				continue
			}

			targetCenterX := float64(targetBox.X) + float64(targetBox.W)/2.0
			targetCenterY := float64(targetBox.Y) + float64(targetBox.H)/2.0

			vOverlap := max(sourceBox.Y, targetBox.Y) < min(sourceBox.Y+sourceBox.H, targetBox.Y+targetBox.H)
			hOverlap := max(sourceBox.X, targetBox.X) < min(sourceBox.X+sourceBox.W, targetBox.X+targetBox.W)

			if vOverlap && targetBox.X >= sourceBox.X+sourceBox.W {
				dist := math.Abs(targetCenterX - sourceCenterX)
				if dist < minDistRight {
					minDistRight = dist
					currentNav.Right = targetID
				}
			}

			if vOverlap && targetBox.X+targetBox.W <= sourceBox.X {
				dist := math.Abs(targetCenterX - sourceCenterX)
				if dist < minDistLeft {
					minDistLeft = dist
					currentNav.Left = targetID
				}
			}

			if hOverlap && targetBox.Y >= sourceBox.Y+sourceBox.H {
				dist := math.Abs(targetCenterY - sourceCenterY)
				if dist < minDistDown {
					minDistDown = dist
					currentNav.Down = targetID
				}
			}

			if hOverlap && targetBox.Y+targetBox.H >= sourceBox.Y {
				dist := math.Abs(targetCenterY - sourceCenterY)
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
		return componentId(node.Component)
	}

	for _, child := range node.Children {
		if id := findFirstComponent(child); id != "" {
			return id
		}
	}

	return ""
}

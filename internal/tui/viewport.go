package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
)

func (m *model) updateViewports() {
	for id, comp := range m.componentMap {
		if comp.Type != "text" {
			continue
		}

		box, ok := m.componentBoxes[id]
		if !ok {
			return
		}

		compW, compH := calcWidthHeight(box.W, box.H)

		vpWidth := max(0, compW)
		vpHeight := max(0, compH)

		vp, exists := m.viewports[id]
		if !exists {
			vp = viewport.New(vpWidth, vpHeight)
			vp.SetContent("[loading...]")
		} else {
			vp.Width = vpWidth
			vp.Height = vpHeight
		}

		vp.YPosition = 1
		vp.MouseWheelEnabled = true

		switch comp.Type {
		case "text":
			if output, dataOk := m.textOutputs[id]; dataOk {
				if output.Error() != nil {
					vp.SetContent(wrapContent(fmt.Sprintf("[error]\n%s", output.Error()), vp.Width))
				} else {
					vp.SetContent(wrapContent(output.Output(), vp.Width))
				}
			} else if vp.TotalLineCount() == 0 {
				vp.SetContent("[loading...]")
			}
		}

		m.viewports[id] = vp
	}
	m.vpReady = true
}

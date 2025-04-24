package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

func (m *model) renderViewportComponent(comp *config.Component, w, h int) string {
	if !m.ready {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Initializing Viewport...]")
	}

	id := componentId(comp)
	vp, ok := m.textComponents[id]
	if !ok {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Error: Viewport not found]")
	}

	_, _, border := m.getBorderStyle()
	header := renderViewportHeader(comp.Title, border)
	combinedHeight := lipgloss.Height(header) + 1

	vp.Width = w
	vp.Height = max(0, h-combinedHeight)
	m.textComponents[id] = vp

	var footer string
	if !(vp.AtTop() && vp.AtBottom()) {
		footer = renderViewportFooter(vp, vp.Width)
	} else {
		footer = ""
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		vp.View(),
		footer,
	)
}

func (m *model) createViewportComponent(id string) {
	vp := viewport.New(0, 0)
	vp.SetContent("[loading...]")
	m.textComponents[id] = vp
}

func (m *model) updateViewportComponent(id string, compW, compH int) {
	vpWidth := max(0, compW)
	vpHeight := max(0, compH)

	vp, exists := m.textComponents[id]
	if !exists {
		vp = viewport.New(vpWidth, vpHeight)
		vp.SetContent("[loading...]")
	} else {
		vp.Width = vpWidth
		vp.Height = vpHeight
	}

	vp.YPosition = 1
	vp.MouseWheelEnabled = true

	if output, dataOk := m.viewportOutputs[id]; dataOk {
		if output.Error() != nil {
			vp.SetContent(wrapContent(fmt.Sprintf("[error]\n%s", output.Error()), vp.Width))
		} else {
			vp.SetContent(wrapContent(output.Output(), vp.Width))
		}
	} else if vp.TotalLineCount() == 0 {
		vp.SetContent("[loading...]")
	}

	m.textComponents[id] = vp
}

func (m *model) setViewportContent(id string, output data.FetchOutput) {
	m.viewportOutputs[id] = output

	if vp, ok := m.textComponents[id]; ok && m.ready {
		if output.Error() != nil {
			vp.SetContent(wrapContent(fmt.Sprintf("[error]\n%s", output.Error()), vp.Width))
		} else {
			vp.SetContent(wrapContent(output.Output(), vp.Width))
		}
		vp.GotoTop()
		m.textComponents[id] = vp
	}
}

func renderViewportHeader(title string, border lipgloss.Border) string {
	style := lipgloss.NewStyle().Border(border, false, true, true, false).Padding(0, 1)
	if title == "" {
		title = "Untitled"
	}
	return style.Render(title)
}

func renderViewportFooter(vp viewport.Model, vpWidth int) string {
	percentStr := fmt.Sprintf("%3.f%% ⥮ ", vp.ScrollPercent()*100)
	paddingWidth := max(0, vpWidth-lipgloss.Width(percentStr))

	padding := strings.Repeat(" ", paddingWidth)

	return padding + percentStr
}

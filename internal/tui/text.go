package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

var (
	vpTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).BorderForeground(lipgloss.Color("69")) // Nice
	}()

	vapInfoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return vpTitleStyle.BorderStyle(b)
	}()

	headerHeight  = lipgloss.Height(renderViewportHeader("", 0))
	footerHeight  = lipgloss.Height(renderViewportFooter(viewport.Model{}, 0))
	vMarginHeight = headerHeight + footerHeight
)

func (m *model) renderTextComponent(comp *config.Component, w, h int) string {
	if !m.vpReady {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Initializing Viewport...]")
	}

	id := componentId(comp)
	vp, ok := m.viewports[id]
	if !ok {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Error: Viewport not found]")
	}

	header := renderViewportHeader(comp.Title, vp.Width)
	footer := renderViewportFooter(vp, vp.Width)
	vp.Width = w
	vp.Height = max(0, h-lipgloss.Height(header)-lipgloss.Height(footer))

	m.viewports[id] = vp

	return fmt.Sprintf("%s\n%s\n%s", header, vp.View(), footer)
}

func (m *model) setTextContent(id string, output data.FetchOutput) {
	m.textOutputs[id] = output

	if vp, ok := m.viewports[id]; ok && m.vpReady {
		if output.Error() != nil {
			vp.SetContent(fmt.Sprintf("[error]\n%s", output.Error()))
		} else {
			vp.SetContent(lipgloss.NewStyle().Width(vp.Width).Render(output.Output()))
		}
		vp.GotoTop()
		m.viewports[id] = vp
	}
}

func renderViewportHeader(title string, vpWidth int) string {
	if title == "" {
		title = "Untitled"
	}
	styledTitle := vpTitleStyle.Render(title)
	lineWidth := max(0, vpWidth-lipgloss.Width(styledTitle))
	line := strings.Repeat(" ", lineWidth)
	return lipgloss.JoinHorizontal(lipgloss.Left, styledTitle, line)
}

func renderViewportFooter(vp viewport.Model, vpWidth int) string {
	info := vapInfoStyle.Render(fmt.Sprintf("%3.f%%", vp.ScrollPercent()*100))
	lineWidth := max(0, vpWidth-lipgloss.Width(info))
	line := strings.Repeat(" ", lineWidth)
	return lipgloss.JoinHorizontal(lipgloss.Right, line, info)
}

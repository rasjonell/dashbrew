package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/data"
)

type TextComponent struct {
	baseComponent
	viewport viewport.Model
	content  string
}

func newTextComponent(base baseComponent) *TextComponent {
	vp := viewport.New(0, 0)
	vp.SetContent("[loading...]")
	vp.YPosition = 1
	vp.MouseWheelEnabled = true
	return &TextComponent{
		baseComponent: base,
		viewport:      vp,
		content:       "[loading...]",
	}
}

func (c *TextComponent) View(w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	c.viewport.Width = innerWidth
	footerHeight := 0
	if !(c.viewport.AtTop() && c.viewport.AtBottom()) {
		footerHeight = 2
	}
	c.viewport.Height = max(0, innerHeight-headerHeight)
	c.viewport.SetContent(WrapContent(c.content, c.viewport.Width))

	var footer string
	if footerHeight > 0 {
		footer = c.renderFooter(c.viewport.Width, c.viewport.ScrollPercent(), c.config.Data.Caption)
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		c.viewport.View(),
		footer,
	)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *TextComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c

	if result.Error() != nil {
		newInstance.err = result.Error()
		newInstance.content = fmt.Sprintf("[error]\n%s", result.Error())
	} else {
		newInstance.err = nil
		newInstance.content = result.Output()
	}

	newInstance.viewport.SetContent(WrapContent(newInstance.content, newInstance.viewport.Width))
	newInstance.viewport.GotoTop()
	return &newInstance, nil
}

func (c *TextComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		c.viewport, cmd = c.viewport.Update(msg)
		return c, cmd
	}
	return c, nil
}

func (c *TextComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

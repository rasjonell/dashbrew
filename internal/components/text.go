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
	if c.inActionOverlayState() {
		return renderActionOverlay(c.baseComponent, w, h, focused)
	}

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
		footerHeight = 1
	}
	c.viewport.Height = max(0, innerHeight-headerHeight-footerHeight)
	c.viewport.SetContent(WrapContent(c.content, c.viewport.Width))

	parts := []string{header, c.viewport.View()}
	if footerHeight > 0 {
		parts = append(parts, c.renderFooter(c.viewport.Width, c.viewport.ScrollPercent(), c.config.Data.Caption))
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, parts...)

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
	newInstance := *c
	if cmd, handled := tickActionSpinner(&newInstance.baseComponent, msg); handled {
		return &newInstance, cmd
	}

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

func (c *TextComponent) BeginAction(label, outputAs, pendingID string) (Component, tea.Cmd) {
	newInstance := *c
	tick := newInstance.beginAction(label, outputAs, pendingID)
	return &newInstance, tick
}

func (c *TextComponent) ApplyActionResult(pendingID, out, outputAs string) (Component, tea.Cmd) {
	newInstance := *c
	if !newInstance.applyActionResult(pendingID, out) {
		return c, nil
	}
	return &newInstance, nil
}

func (c *TextComponent) ResetAction() (Component, tea.Cmd) {
	newInstance := *c
	newInstance.resetAction()
	return &newInstance, nil
}

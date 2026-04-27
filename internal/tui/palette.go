package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
)

type paletteItem struct {
	componentID    string
	componentTitle string
	actionLabel    string
	key            string
	action         config.ActionConfig
}

func (i paletteItem) Title() string {
	return fmt.Sprintf("[%s] %s", i.componentTitle, i.actionLabel)
}
func (i paletteItem) Description() string { return fmt.Sprintf("key: %s · type: %s", i.key, i.action.Type) }
func (i paletteItem) FilterValue() string {
	return i.componentTitle + " " + i.actionLabel + " " + i.key
}

type paletteState struct {
	open bool
	list list.Model
}

func (m *model) openPalette() {
	m.prevFocusBeforePalette = m.focusedComponentId

	items := []list.Item{}
	for _, comp := range m.components {
		title := ""
		if cfg := comp.Config(); cfg != nil {
			title = cfg.Title
			if title == "" {
				title = cfg.Type
			}
		}
		for _, b := range comp.Bindings() {
			items = append(items, paletteItem{
				componentID:    comp.ID(),
				componentTitle: title,
				actionLabel:    b.Label,
				key:            b.Key,
				action:         b.Action,
			})
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	l := list.New(items, delegate, 0, 0)
	l.Title = "Actions"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	m.palette = paletteState{open: true, list: l}
	m.resizePalette()
}

func (m *model) closePalette() {
	m.palette = paletteState{}
}

func (m *model) resizePalette() {
	if !m.palette.open {
		return
	}
	w := m.width * 60 / 100
	h := m.height * 60 / 100
	if w < 30 {
		w = min(30, m.width-4)
	}
	if h < 8 {
		h = min(8, m.height-4)
	}
	// reserve a couple rows for the bordered modal frame
	m.palette.list.SetWidth(w - 4)
	m.palette.list.SetHeight(h - 4)
}

func (m *model) handlePaletteKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// If filter is engaged, let the list consume the Esc to clear it.
		if m.palette.list.FilterState() != list.Unfiltered {
			var cmd tea.Cmd
			m.palette.list, cmd = m.palette.list.Update(msg)
			return cmd
		}
		m.focusedComponentId = m.prevFocusBeforePalette
		m.closePalette()
		return nil
	case "ctrl+k":
		m.focusedComponentId = m.prevFocusBeforePalette
		m.closePalette()
		return nil
	case "enter":
		// In filter mode, Enter applies the filter — let the list handle it,
		// then on the next Enter (in FilterApplied state) we run the selection.
		if m.palette.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.palette.list, cmd = m.palette.list.Update(msg)
			return cmd
		}
		if sel, ok := m.palette.list.SelectedItem().(paletteItem); ok {
			componentID := sel.componentID
			label := sel.actionLabel
			action := sel.action
			m.focusedComponentId = componentID
			m.closePalette()
			return m.dispatchAction(componentID, label, action)
		}
		return nil
	}
	var cmd tea.Cmd
	m.palette.list, cmd = m.palette.list.Update(msg)
	return cmd
}

func (m *model) renderPalette() string {
	w := m.width * 60 / 100
	h := m.height * 60 / 100
	if w < 30 {
		w = min(30, m.width-4)
	}
	if h < 8 {
		h = min(8, m.height-4)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(1, 1).
		Width(w - 2).
		Height(h - 2).
		Render(m.palette.list.View())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

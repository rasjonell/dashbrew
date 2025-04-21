package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
)

type refreshMsg struct {
	ID string
}

func (m *model) scheduleRefreshes() []tea.Cmd {
	var cmds []tea.Cmd

	for id, comp := range m.componentMap {
		if comp.Data.RefreshInterval > 0 {
			cmds = append(cmds, m.scheduleSingleRefresh(id, comp))
		}
	}

	return cmds
}

func (m *model) scheduleSingleRefresh(id string, comp *config.Component) tea.Cmd {
	refreshInterval := 5 // default
	if comp.Data.RefreshInterval > 0 {
		refreshInterval = comp.Data.RefreshInterval
	}

	return func() tea.Msg {
		time.Sleep(time.Duration(refreshInterval) * time.Second)
		return refreshMsg{
			ID: id,
		}
	}
}

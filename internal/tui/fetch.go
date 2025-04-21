package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type fetchResultMsg struct {
	ID     string
	Result data.FetchOutput
}

func (m *model) fetchAllData() []tea.Cmd {
	var cmds []tea.Cmd
	for id, comp := range m.componentMap {
		cmds = append(cmds, fetchComponentAsyncCmd(id, comp))
	}
	return cmds
}

func fetchComponentAsyncCmd(id string, comp *config.Component) tea.Cmd {
	return func() tea.Msg {
		var result data.FetchOutput

		switch comp.Data.Source {
		case "script":
			result = data.RunScript(comp.Data.Command)
		case "api":
			result = data.RunAPI(comp.Data.URL)
		}

		return fetchResultMsg{
			ID:     id,
			Result: result,
		}
	}
}

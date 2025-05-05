package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/components"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type fetchResultMsg struct {
	ID     string
	Result data.FetchOutput
}

func (m *model) fetchAllData() []tea.Cmd {
	var cmds []tea.Cmd
	for id, comp := range m.components {
		cmds = append(cmds, fetchComponentAsyncCmd(id, comp.Config()))
	}
	return cmds
}

func fetchComponentAsyncCmd(id string, comp *config.Component) tea.Cmd {
	if comp.Data == nil {
		return func() tea.Msg {
			return fetchResultMsg{
				ID:     id,
				Result: data.NewFetchOutput("", fmt.Errorf("component data source is nil")),
			}
		}
	}

	return func() tea.Msg {
		var result data.FetchOutput

		if comp.Type == "todo" {
			items, err := data.ReadTodoFile(comp.Data.Source)
			return fetchResultMsg{
				ID: id,
				Result: &components.TodoFetchOutput{
					Err:       err,
					TodoItems: items,
				},
			}
		}

		switch comp.Data.Source {
		case "script":
			result = data.RunScript(comp.Data.Command)
		case "api":
			result = data.RunAPI(comp.Data.URL, comp.Data.JSONPath)
		default:
			result = data.NewFetchOutput("", fmt.Errorf("unknown data source %s", comp.Data.Source))
		}

		return fetchResultMsg{
			ID:     id,
			Result: result,
		}
	}
}

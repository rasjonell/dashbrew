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

type todoFetchOutput struct {
	err   error
	items []*data.TodoOutput
}

func (t *todoFetchOutput) Output() string            { return "" }
func (t *todoFetchOutput) Error() error              { return t.err }
func (t *todoFetchOutput) Items() []*data.TodoOutput { return t.items }

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

		if comp.Type == "todo" {
			items, err := data.ReadTodoFile(comp.Data.Source)
			return fetchResultMsg{
				ID: id,
				Result: &todoFetchOutput{
					err:   err,
					items: items,
				},
			}
		}

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

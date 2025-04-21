package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type fetchDataMsg struct{}

func fetchDataCmd() tea.Msg {
	return fetchDataMsg{}
}

func (m *model) fetchAllData() {
	for id, comp := range m.componentMap {
		m.fetchSingleComponentData(id, comp)
	}
}

func (m *model) fetchSingleComponentData(id string, comp *config.Component) {
	var output data.FetchOutput

	switch comp.Data.Source {
	case "script":
		output = data.RunScript(comp.Data.Command)
	case "api":
		output = data.RunAPI(comp.Data.URL)
	}

	switch comp.Type {
	case "text":
		m.setTextContent(id, output)
	}
}

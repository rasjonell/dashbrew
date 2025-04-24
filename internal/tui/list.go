package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type ListItem struct {
	Val string
}

func (i ListItem) Title() string       { return i.Val }
func (i ListItem) Description() string { return "" }
func (i ListItem) FilterValue() string { return i.Val }

func (m *model) renderListComponent(comp *config.Component, w, h int) string {
	if !m.ready {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Initializing List...]")
	}
	id := componentId(comp)
	listModel, ok := m.listComponents[id]

	if !ok {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Error: List not found]")
	}

	return listModel.View()
}

func (m *model) createListComponent(id string, comp *config.Component) {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	// TODO: styles
	newList := list.New(nil, delegate, 0, 0)
	newList.Title = comp.Title
	newList.SetShowHelp(false)
	newList.SetShowStatusBar(true)
	newList.SetFilteringEnabled(true)
	m.listComponents[id] = newList
}

func (m *model) updateListComponent(id string, compW, compH int) {
	if listModel, listExists := m.listComponents[id]; listExists {
		listModel.SetSize(compW, compH)
		m.listComponents[id] = listModel
	}
}

func (m *model) setListContent(id string, result data.FetchOutput, comp *config.Component) tea.Cmd {
	listModel, listExists := m.listComponents[id]
	if !listExists {
		return nil
	}

	var items []list.Item
	if result.Error() != nil {
		items = []list.Item{ListItem{Val: fmt.Sprintf("[Error: %v]", result.Error())}}
	} else {
		items = parseDataToListItems(comp, result.Output())
	}
	cmd := listModel.SetItems(items)
	m.listComponents[id] = listModel
	return cmd
}

func parseDataToListItems(comp *config.Component, rawData string) []list.Item {
	var items []list.Item

	switch comp.Data.Source {
	case "script":
		lines := strings.Split(strings.TrimSpace(rawData), "\n")
		for _, line := range lines {
			if line != "" {
				items = append(items, ListItem{Val: line})
			}
		}
	case "api":
		var stringSlice []string
		err := json.Unmarshal([]byte(rawData), &stringSlice)
		if err == nil {
			for _, s := range stringSlice {
				items = append(items, ListItem{Val: s})
			}
		} else {
			items = append(items, ListItem{Val: strings.TrimSpace(rawData)})
		}
	}

	return items
}

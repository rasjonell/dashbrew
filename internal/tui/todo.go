package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type TodoListItem struct {
	*data.TodoOutput
	Index int
}

func (i TodoListItem) Title() string {
	box := "[ ]"
	if i.TodoOutput.Done {
		box = "[x]"
	}

	return fmt.Sprintf("%s %s", box, i.TodoOutput.Title)
}
func (i TodoListItem) Description() string { return "" }
func (i TodoListItem) FilterValue() string { return i.Title() }

func (m *model) renderTodoComponent(comp *config.Component, w, h int) string {
	if !m.ready {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Initializing ToDo List...]")
	}
	id := componentId(comp)
	listModel, ok := m.listComponents[id]

	if !ok {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, "[Error: ToDo List not found]")
	}

	if m.isAdding && id == m.focusedComponentId {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			"New todo: "+m.additionInput+"_")
	}

  _, _, border := m.getBorderStyle()
  header := renderComponentHeader(comp.Title, border)
  listModel.SetHeight(max(0, h - lipgloss.Height(header)))
  m.listComponents[id] = listModel

	return lipgloss.JoinVertical(lipgloss.Left,header, listModel.View())
}

func (m *model) createTodoComponent(id string) {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	// TODO: styles
	newList := list.New(nil, delegate, 0, 0)
  newList.SetShowTitle(false)
	newList.SetShowHelp(false)
	newList.SetShowStatusBar(false)
	newList.SetFilteringEnabled(true)
	m.listComponents[id] = newList
}

func (m *model) setTodoContent(id string, items []*data.TodoOutput, selectIdx ...int) tea.Cmd {
	listModel, exists := m.listComponents[id]
	if !exists {
		return nil
	}

	newSelectIndex := listModel.GlobalIndex()
	if len(selectIdx) > 0 {
		newSelectIndex = selectIdx[0]
	}

	var listItems []list.Item
	for idx, item := range items {
		listItems = append(listItems, TodoListItem{TodoOutput: item, Index: idx})
	}
	cmd := listModel.SetItems(listItems)
	listModel.Select(newSelectIndex)
	m.listComponents[id] = listModel
	return cmd
}

func (m *model) toggleTodoState(listModel list.Model) {
	selected := listModel.SelectedItem()
	if todoItem, ok := selected.(TodoListItem); ok {
		todoItem.Done = !todoItem.Done
		comp := m.componentMap[m.focusedComponentId]
		path := comp.Data.Source
		items, _ := data.ReadTodoFile(path)
		if todoItem.Index < len(items) {
			items[todoItem.Index].Done = todoItem.Done
			data.WriteTodoFile(path, items)
			m.setTodoContent(m.focusedComponentId, items)
		}
	}
}

func (m *model) addNewTodo() {
	comp := m.componentMap[m.focusedComponentId]
	path := comp.Data.Source

	items, _ := data.ReadTodoFile(path)
	items = append(items, &data.TodoOutput{Done: false, Title: m.additionInput})
	data.WriteTodoFile(path, items)
	m.setTodoContent(m.focusedComponentId, items)
	m.isAdding = false
	m.additionInput = ""
}

func (m *model) removeTodo(listModel list.Model) {
	selected := listModel.SelectedItem()
	if todoItem, ok := selected.(TodoListItem); ok {
		comp := m.componentMap[m.focusedComponentId]
		path := comp.Data.Source
		items, _ := data.ReadTodoFile(path)
		if todoItem.Index < len(items) {
			items = append(items[:todoItem.Index], items[todoItem.Index+1:]...)
			data.WriteTodoFile(path, items)
			m.setTodoContent(m.focusedComponentId, items, todoItem.Index-1)
		}
	}
}

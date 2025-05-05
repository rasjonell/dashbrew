package components

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/data"
)

type ListItem struct {
	Val string
}

func (i ListItem) Title() string       { return i.Val }
func (i ListItem) Description() string { return "" }
func (i ListItem) FilterValue() string { return i.Val }

type ListComponent struct {
	baseComponent
	list list.Model
}

func newListComponent(base baseComponent) *ListComponent {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	// TODO: styles
	l := list.New(nil, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &ListComponent{
		baseComponent: base,
		list:          l,
	}
}

func (c *ListComponent) View(w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(c.styles)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	c.list.SetWidth(innerWidth)
	c.list.SetHeight(max(0, innerHeight-headerHeight))

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		c.list.View(),
	)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *ListComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		c.list, cmd = c.list.Update(msg)
		return c, cmd
	}
	return c, nil
}

func (c *ListComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c

	var cmd tea.Cmd
	var items []list.Item

	if result.Error() != nil {
		newInstance.err = result.Error()
		items = []list.Item{ListItem{Val: fmt.Sprintf("[Error: %v]", result.Error())}}
	} else {
		newInstance.err = nil
		items = c.parseDataToListItems(result.Output())
	}

	cmd = newInstance.list.SetItems(items)
	return &newInstance, cmd
}

func (c *ListComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

func (c *ListComponent) parseDataToListItems(rawData string) []list.Item {
	var items []list.Item

	switch c.config.Data.Source {
	case "script":
		lines := strings.Split(strings.TrimSpace(rawData), "\n")
		for _, line := range lines {
			if line != "" {
				items = append(items, ListItem{Val: strings.TrimSpace(line)})
			}
		}
	case "api":
		var jsonArray []any
		err := json.Unmarshal([]byte(rawData), &jsonArray)
		if err == nil {
			for _, s := range jsonArray {
				items = append(items, ListItem{Val: fmt.Sprintf("%v", s)})
			}
		} else {
			items = append(items, ListItem{Val: strings.TrimSpace(rawData)})
		}
	}

	return items
}

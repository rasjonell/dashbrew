package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type TodoListItem struct {
	*data.TodoOutput
	Index  int
	styles *config.StyleConfig
}

func (i TodoListItem) Title() string {
	if i.TodoOutput.Done {
		successColor := PaletteColor(i.styles, "success", lipgloss.Color("#5faf5f"))
		box := lipgloss.NewStyle().Foreground(successColor).Render("[✓]")
		text := lipgloss.NewStyle().Strikethrough(true).Faint(true).Render(i.TodoOutput.Title)
		return fmt.Sprintf("%s %s", box, text)
	}
	return fmt.Sprintf("%s %s", "[ ]", i.TodoOutput.Title)
}

func (i TodoListItem) Description() string { return "" }
func (i TodoListItem) FilterValue() string {
	// Filter on plain text so users can search by raw title regardless of styles.
	box := "[ ]"
	if i.TodoOutput.Done {
		box = "[x]"
	}
	return fmt.Sprintf("%s %s", box, i.TodoOutput.Title)
}

type TodoFetchOutput struct {
	Err       error
	TodoItems []*data.TodoOutput
}

func (t *TodoFetchOutput) Output() string            { return "" }
func (t *TodoFetchOutput) Error() error              { return t.Err }
func (t *TodoFetchOutput) Items() []*data.TodoOutput { return t.TodoItems }

type TodoComponent struct {
	baseComponent
	addInput string
	list     list.Model
	items    []*data.TodoOutput
}

func newTodoComponent(base baseComponent) *TodoComponent {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	// TODO: styles
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &TodoComponent{
		baseComponent: base,
		list:          l,
		items:         []*data.TodoOutput{},
	}
}

func (c *TodoComponent) SupportsAdd() bool   { return true }
func (c *TodoComponent) GetAddInput() string { return c.addInput }

func (c *TodoComponent) View(w, h int, focused bool) string {
	if c.inActionOverlayState() {
		return renderActionOverlay(&c.baseComponent, w, h, focused)
	}

	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border, focused, innerWidth)
	headerHeight := lipgloss.Height(header)

	c.list.SetWidth(innerWidth)
	c.list.SetHeight(max(0, innerHeight-headerHeight))

	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, c.list.View())

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *TodoComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	newInstance := *c

	if cmd, handled := tickActionSpinner(&newInstance.baseComponent, msg); handled {
		return &newInstance, cmd
	}
	if cmd, handled := updateActionViewport(&newInstance.baseComponent, msg); handled {
		return &newInstance, cmd
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if newInstance.list.FilterState() != list.Filtering {
			switch {
			case key.Matches(msg, keys.Space):
				cmd = newInstance.toggleTodoState()
				cmds = append(cmds, cmd)
			case key.Matches(msg, keys.Delete):
				cmd = newInstance.removeTodo()
				cmds = append(cmds, cmd)
			}
		}
	}

	newInstance.list, cmd = newInstance.list.Update(msg)
	cmds = append(cmds, cmd)
	return &newInstance, tea.Batch(cmds...)
}

func (c *TodoComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c
	var cmd tea.Cmd

	if todoRes, ok := result.(*TodoFetchOutput); ok {
		if todoRes.Error() != nil {
			newInstance.err = todoRes.Error()
			errorItem := TodoListItem{
				Index: -1,
				TodoOutput: &data.TodoOutput{
					Title: fmt.Sprintf("[Error: %v]", todoRes.Error()),
				},
				styles: c.styles,
			}
			cmd = newInstance.list.SetItems([]list.Item{errorItem})
			newInstance.items = []*data.TodoOutput{}
		} else {
			newInstance.err = nil
			newInstance.items = todoRes.Items()
			cmd = newInstance.updateListItems()
		}
	}

	return &newInstance, cmd
}

func (c *TodoComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	newInstance := *c

	switch {
	case key.Matches(msg, keys.Esc):
		newInstance.addInput = ""
		return &newInstance, true, nil

	case key.Matches(msg, keys.Backspace):
		if len(newInstance.addInput) > 0 {
			newInstance.addInput = newInstance.addInput[:len(newInstance.addInput)-1]
		}
		return &newInstance, false, nil

	case key.Matches(msg, keys.Enter):
		if newInstance.addInput != "" {
			cmd := newInstance.addNewTodo()
			newInstance.addInput = ""
			return &newInstance, true, cmd
		}

		newInstance.addInput = ""
		return &newInstance, true, nil

	default:
		if msg.Type == tea.KeyRunes {
			newInstance.addInput += msg.String()
		} else if msg.Type == tea.KeySpace {
			newInstance.addInput += " "
		}
		return &newInstance, false, nil
	}
}

func (c *TodoComponent) updateListItems(selectIdx ...int) tea.Cmd {
	listItems := make([]list.Item, len(c.items))
	for i, item := range c.items {
		listItems[i] = TodoListItem{TodoOutput: item, Index: i, styles: c.styles}
	}

	newIdx := -1
	currentIdx := c.list.Index()

	if len(selectIdx) > 0 {
		newIdx = selectIdx[0]
	} else if currentIdx >= 0 && currentIdx < len(listItems) {
		newIdx = currentIdx
	} else if len(listItems) > 0 {
		newIdx = 0
	}

	cmd := c.list.SetItems(listItems)
	if newIdx >= 0 && newIdx < len(listItems) {
		c.list.Select(newIdx)
	}

	return cmd
}

func (c *TodoComponent) toggleTodoState() tea.Cmd {
	selected := c.list.SelectedItem()
	if todoItem, ok := selected.(TodoListItem); ok && todoItem.Index >= 0 && todoItem.Index < len(c.items) {
		c.items[todoItem.Index].Done = !c.items[todoItem.Index].Done
		c.writeTodos()
		return c.updateListItems()
	}
	return nil
}

func (c *TodoComponent) addNewTodo() tea.Cmd {
	if c.addInput == "" {
		return nil
	}

	newItem := &data.TodoOutput{Done: false, Title: c.addInput}
	c.items = append(c.items, newItem)
	c.addInput = ""

	c.writeTodos()

	return c.updateListItems(len(c.items) - 1)
}

func (c *TodoComponent) removeTodo() tea.Cmd {
	selected := c.list.SelectedItem()
	if todoItem, ok := selected.(TodoListItem); ok && todoItem.Index >= 0 && todoItem.Index < len(c.items) {
		ogIdx := todoItem.Index
		c.items = append(c.items[:ogIdx], c.items[ogIdx+1:]...)

		c.writeTodos()

		newIdx := ogIdx
		if newIdx >= len(c.items) {
			newIdx = len(c.items) - 1
		}
		if newIdx < 0 {
			newIdx = 0
		}

		return c.updateListItems(newIdx)
	}

	return nil
}

func (c *TodoComponent) writeTodos() {
	err := data.WriteTodoFile(c.config.Data.Source, c.items)
	if err != nil {
		c.err = fmt.Errorf("failed to save todo: %w", err)
	} else {
		c.err = nil
	}
}

func (c *TodoComponent) BeginAction(label, outputAs, pendingID string) (Component, tea.Cmd) {
	newInstance := *c
	tick := newInstance.beginAction(label, outputAs, pendingID)
	return &newInstance, tick
}

func (c *TodoComponent) ApplyActionResult(pendingID, out, outputAs string) (Component, tea.Cmd) {
	newInstance := *c
	if !newInstance.applyActionResult(pendingID, out) {
		return c, nil
	}
	return &newInstance, nil
}

func (c *TodoComponent) ResetAction() (Component, tea.Cmd) {
	newInstance := *c
	newInstance.resetAction()
	return &newInstance, nil
}

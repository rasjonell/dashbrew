package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type model struct {
	cfg *config.DashboardConfig

	width              int
	height             int
	ready              bool
	isAdding           bool
	initialized        bool
	focusedComponentId string

	additionInput   string
	viewportOutputs map[string]data.FetchOutput

	componentBoxes map[string]*boundingBox
	navMap         map[string]*navigationMap
	componentMap   map[string]*config.Component

	listComponents map[string]list.Model
	textComponents map[string]viewport.Model
}

type componentOutput struct {
	Output string
	Error  error
}

type keyMap struct {
	Quit      key.Binding
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Space     key.Binding
	Enter     key.Binding
	Esc       key.Binding
	Add       key.Binding
	Refresh   key.Binding
	Backspace key.Binding
	Delete    key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
	),
	Up: key.NewBinding(
		key.WithKeys("shift+up", "K"),
	),
	Down: key.NewBinding(
		key.WithKeys("shift+down", "J"),
	),
	Left: key.NewBinding(
		key.WithKeys("shift+left", "H"),
	),
	Right: key.NewBinding(
		key.WithKeys("shift+right", "L"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
	),
	Add: key.NewBinding(
		key.WithKeys("a", "A"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r", "R"),
	),
	Backspace: key.NewBinding(
		key.WithKeys(tea.KeyBackspace.String()),
	),
	Delete: key.NewBinding(
		key.WithKeys(tea.KeyBackspace.String(), tea.KeyDelete.String(), "d", "D"),
	),
}

func New(cfg *config.DashboardConfig) tea.Model {
	return &model{
		cfg:         cfg,
		ready:       false,
		isAdding:    false,
		initialized: false,

		viewportOutputs: make(map[string]data.FetchOutput),

		componentBoxes: make(map[string]*boundingBox),
		navMap:         make(map[string]*navigationMap),
		componentMap:   make(map[string]*config.Component),

		listComponents: make(map[string]list.Model),
		textComponents: make(map[string]viewport.Model),
	}
}

func (m *model) Init() tea.Cmd {
	m.buildComponentMap(m.cfg.Layout)
	m.focusedComponentId = findFirstComponent(m.cfg.Layout)

	fetchCmds := m.fetchAllData()
	refreshCmds := m.scheduleRefreshes()

	cmds := append(fetchCmds, refreshCmds...)
	cmds = append(cmds, tea.ClearScreen)

	m.initialized = true

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	focusedComponentType := ""
	if comp, ok := m.componentMap[m.focusedComponentId]; ok {
		focusedComponentType = comp.Type
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg.Width, msg.Height)

	case fetchResultMsg:
		if comp, ok := m.componentMap[msg.ID]; ok {
			switch comp.Type {
			case "text":
				m.setViewportContent(msg.ID, msg.Result)
			case "list":
				cmd = m.setListContent(msg.ID, msg.Result, comp)
				cmds = append(cmds, cmd)
			case "todo":
				if todoRes, ok := msg.Result.(*todoFetchOutput); ok {
					m.setTodoContent(msg.ID, todoRes.Items())
				}
			}
		}

	case refreshMsg:
		if comp, ok := m.componentMap[msg.ID]; ok {
			fetchCmd := fetchComponentAsyncCmd(msg.ID, comp)
			rescheduleCmd := m.scheduleSingleRefresh(msg.ID, comp)
			cmd = tea.Batch(fetchCmd, rescheduleCmd)
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			m.focusClicked(msg.X, msg.Y)
		}

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		if m.isAdding {
			switch {
			case key.Matches(msg, keys.Esc):
				m.isAdding = false
				m.additionInput = ""
				return m, nil
			case key.Matches(msg, keys.Backspace):
				if len(m.additionInput) > 0 {
					m.additionInput = m.additionInput[:len(m.additionInput)-1]
				}
				return m, nil
			case key.Matches(msg, keys.Enter):
				switch focusedComponentType {
				case "todo":
					m.addNewTodo()
					return m, nil
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.additionInput += msg.String()
				} else if msg.Type == tea.KeySpace {
					m.additionInput += " "
				}
				return m, nil
			}
		} else {
			switch {
			case key.Matches(msg, keys.Up):
				if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Up != "" {
					m.focusedComponentId = nav.Up
					return m, nil
				}
			case key.Matches(msg, keys.Down):
				if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Down != "" {
					m.focusedComponentId = nav.Down
					return m, nil
				}
			case key.Matches(msg, keys.Left):
				if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Left != "" {
					m.focusedComponentId = nav.Left
					return m, nil
				}
			case key.Matches(msg, keys.Right):
				if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Right != "" {
					m.focusedComponentId = nav.Right
					return m, nil
				}
			case key.Matches(msg, keys.Refresh):
				if comp, ok := m.componentMap[m.focusedComponentId]; ok && comp.Data.RefreshInterval > 0 {
					return m, fetchComponentAsyncCmd(m.focusedComponentId, comp)
				}
			case key.Matches(msg, keys.Add):
				if comp, ok := m.componentMap[m.focusedComponentId]; ok && supportsAddition(comp.Type) {
					m.isAdding = true
					m.additionInput = ""
					return m, nil
				}
			}
		}

		if m.ready {
			switch focusedComponentType {
			case "text":
				if vp, vpOk := m.textComponents[m.focusedComponentId]; vpOk {
					m.textComponents[m.focusedComponentId], cmd = vp.Update(msg)
					cmds = append(cmds, cmd)
				}
			case "list":
				if listModel, listOk := m.listComponents[m.focusedComponentId]; listOk {
					m.listComponents[m.focusedComponentId], cmd = listModel.Update(msg)
					cmds = append(cmds, cmd)
				}
			case "todo":
				if listModel, listOk := m.listComponents[m.focusedComponentId]; listOk {
					m.listComponents[m.focusedComponentId], cmd = listModel.Update(msg)
					cmds = append(cmds, cmd)
					switch {
					case key.Matches(msg, keys.Space):
						m.toggleTodoState(listModel)
					case key.Matches(msg, keys.Delete):
						m.removeTodo(listModel)
					}
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.cfg == nil {
		return "No config loaded."
	}

	if m.width == 0 || m.height == 0 || !m.ready {
		return "Loading..."
	}

	if len(m.componentBoxes) != 0 && m.width > 0 && m.height > 0 {
		m.handleResize(m.width, m.height)
	}

	return m.renderNode(
		m.cfg.Layout,
		m.width, m.height,
		m.getComponentContent,
		m.focusedComponentId,
	)
}

package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/components"
	"github.com/rasjonell/dashbrew/internal/config"
)

type model struct {
	cfg *config.DashboardConfig

	width              int
	height             int
	ready              bool
	isAdding           bool
	initialized        bool
	focusedComponentId string

	components map[string]components.Component

	componentBoxes map[string]*boundingBox
	navMap         map[string]*navigationMap
}

type componentOutput struct {
	Output string
	Error  error
}

type keyMap struct {
	Up      key.Binding
	Add     key.Binding
	Down    key.Binding
	Left    key.Binding
	Quit    key.Binding
	Right   key.Binding
	Refresh key.Binding
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
	Add: key.NewBinding(
		key.WithKeys("a", "A"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r", "R"),
	),
}

func New(cfg *config.DashboardConfig) tea.Model {
	return &model{
		cfg:         cfg,
		ready:       false,
		isAdding:    false,
		initialized: false,

		components: make(map[string]components.Component),

		componentBoxes: make(map[string]*boundingBox),
		navMap:         make(map[string]*navigationMap),
	}
}

func (m *model) Init() tea.Cmd {
	if m.cfg == nil || m.cfg.Layout == nil {
		// TODO: error cmd
		m.initialized = true
		return nil
	}

	m.buildComponentMap(m.cfg.Layout)

	if len(m.components) == 0 {
		// TODO: error cmd
		m.initialized = true
		return nil
	}

	m.focusedComponentId = findFirstComponent(m.cfg.Layout)
	if m.focusedComponentId == "" && len(m.components) > 0 {
		for id := range m.components {
			m.focusedComponentId = id
			break
		}
	}

	var initCmds []tea.Cmd
	for _, comp := range m.components {
		initCmds = append(initCmds, comp.Init())
	}

	fetchCmds := m.fetchAllData()
	refreshCmds := m.scheduleRefreshes()

	cmds := append(initCmds, fetchCmds...)
	cmds = append(cmds, refreshCmds...)
	cmds = append(cmds, tea.ClearScreen)

	m.initialized = true

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	if !m.initialized || m.focusedComponentId == "" {
		if msg, ok := msg.(tea.WindowSizeMsg); ok {
			m.handleResize(msg.Width, msg.Width)
		} else {
			return m, nil
		}
	}

	focusedComp, focusedExists := m.components[m.focusedComponentId]

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg.Width, msg.Height)

	case fetchResultMsg:
		if comp, ok := m.components[msg.ID]; ok {
			updatedComp, cmd := comp.SetContent(msg.Result)
			m.components[msg.ID] = updatedComp
			cmds = append(cmds, cmd)
		}

	case refreshMsg:
		if comp, ok := m.components[msg.ID]; ok {
			fetchCmd := fetchComponentAsyncCmd(comp.ID(), comp.Config())
			rescheduleCmd := m.scheduleSingleRefresh(comp.ID(), comp.Config())
			cmds = append(cmds, fetchCmd, rescheduleCmd)
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			m.focusClicked(msg.X, msg.Y)
			focusedComp, focusedExists = m.components[m.focusedComponentId]
		}

		if focusedExists && focusedComp.IsFocusable() {
			updatedComp, cmd := focusedComp.Update(msg)
			m.components[m.focusedComponentId] = updatedComp
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		if m.isAdding {
			if focusedExists && focusedComp.SupportsAdd() {
				var existAddMode bool
				var updatedComp components.Component

				updatedComp, existAddMode, cmd = focusedComp.HandleAddMode(msg)
				m.components[m.focusedComponentId] = updatedComp
				cmds = append(cmds, cmd)
				if existAddMode {
					m.isAdding = false
				}
			} else {
				m.isAdding = false
			}

			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, keys.Up):
			if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Up != "" {
				m.tryFocus(nav.Up)
			}
		case key.Matches(msg, keys.Down):
			if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Down != "" {
				m.tryFocus(nav.Down)
			}
		case key.Matches(msg, keys.Left):
			if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Left != "" {
				m.tryFocus(nav.Left)
			}
		case key.Matches(msg, keys.Right):
			if nav, ok := m.navMap[m.focusedComponentId]; ok && nav.Right != "" {
				m.tryFocus(nav.Right)
			}

		case key.Matches(msg, keys.Add):
			if focusedExists && focusedComp.SupportsAdd() {
				m.isAdding = true
			}

		case key.Matches(msg, keys.Refresh):
			if focusedExists && focusedComp.SupportsRefresh() {
				cmd = fetchComponentAsyncCmd(focusedComp.ID(), focusedComp.Config())
				cmds = append(cmds, cmd)
			}

		default:
			if focusedExists && focusedComp.IsFocusable() {
				updatedComp, cmd := focusedComp.Update(msg)
				m.components[m.focusedComponentId] = updatedComp
				cmds = append(cmds, cmd)
			}
		}
	}

	if !m.ready && m.width > 0 && m.height > 0 {
		m.ready = true
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.cfg == nil {
		return "Error: no config loaded."
	}
	if !m.initialized {
		return "Initializing..."
	}
	if m.width == 0 || m.height == 0 {
		return "Resizing..."
	}

	return m.renderNode(
		m.cfg.Layout,
		m.width, m.height,
		m.focusedComponentId,
	)
}

package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type model struct {
	cfg *config.DashboardConfig

	width       int
	height      int
	initialized bool

	vpReady   bool
	viewports map[string]viewport.Model

	textOutputs  map[string]data.FetchOutput
	componentMap map[string]*config.Component

	componentBoxes     map[string]*boundingBox
	navMap             map[string]*navigationMap
	focusedComponentId string
}

type componentOutput struct {
	Output string
	Error  error
}

type keyMap struct {
	Quit  key.Binding
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q", "esc"),
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
}

func New(cfg *config.DashboardConfig) tea.Model {
	return &model{
		cfg:         cfg,
		initialized: false,

		vpReady:   false,
		viewports: make(map[string]viewport.Model),

		textOutputs:  make(map[string]data.FetchOutput),
		componentMap: make(map[string]*config.Component),

		componentBoxes: make(map[string]*boundingBox),
		navMap:         make(map[string]*navigationMap),
	}
}

func (m *model) Init() tea.Cmd {
	m.buildComponentMap(m.cfg.Layout)
	m.focusedComponentId = findFirstComponent(m.cfg.Layout)

	cmds := m.scheduleRefreshes()
	cmds = append(cmds, fetchDataCmd, tea.ClearScreen)
	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg.Width, msg.Height)

	case fetchDataMsg:
		m.fetchAllData()
		m.initialized = true

	case refreshMsg:
		if comp, ok := m.componentMap[msg.ID]; ok {
			m.fetchSingleComponentData(msg.ID, comp)
			cmd = m.scheduleSingleRefresh(msg.ID, comp)
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			m.focusClicked(msg.X, msg.Y)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
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
		}
	}

	if comp, ok := m.componentMap[m.focusedComponentId]; ok && comp.Type == "text" {
		if vp, vpOk := m.viewports[m.focusedComponentId]; vpOk {
			m.viewports[m.focusedComponentId], cmd = vp.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.cfg == nil {
		return "No config loaded."
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if len(m.componentBoxes) == 0 && m.width > 0 && m.height > 0 {
		m.handleResize(m.width, m.height)
	}

	return renderNode(
		m.cfg.Layout,
		m.width, m.height,
		m.getComponentContent,
		m.focusedComponentId,
	)
}

package components

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type keyMap struct {
	Esc       key.Binding
	Enter     key.Binding
	Space     key.Binding
	Delete    key.Binding
	Backspace key.Binding
}

var keys = keyMap{
	Esc: key.NewBinding(
		key.WithKeys("esc"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete", "d", "D"),
	),
	Backspace: key.NewBinding(
		key.WithKeys(tea.KeyBackspace.String()),
	),
}

type Component interface {
	ID() string
	Type() string
	IsFocusable() bool
	SupportsAdd() bool
	GetAddInput() string
	SupportsRefresh() bool
	Config() *config.Component

	Init() tea.Cmd
	View(w, h int, focused bool) string
	Update(msg tea.Msg) (Component, tea.Cmd)
	SetContent(result data.FetchOutput) (Component, tea.Cmd)
	HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd)
}

func NewComponent(cfg *config.Component, styles *config.StyleConfig) Component {
	id := ComponentId(cfg)
	base := baseComponent{
		id:     id,
		config: cfg,
		styles: styles,
	}

	switch cfg.Type {
	case "text":
		return newTextComponent(base)
	case "list":
		return newListComponent(base)
	case "todo":
		return newTodoComponent(base)
	case "chart":
		return newChartComponent(base)
	case "table":
		return newTableComponent(base)
	case "histogram":
		return newHistogramComponent(base)
	default:
		return newErrorComponent(base, "unknown component type: "+cfg.Type)
	}
}

type baseComponent struct {
	err    error
	id     string
	config *config.Component
	styles *config.StyleConfig
}

func (b baseComponent) Init() tea.Cmd             { return nil }
func (b baseComponent) GetAddInput() string       { return "" }
func (b baseComponent) ID() string                { return b.id }
func (b baseComponent) IsFocusable() bool         { return true }
func (b baseComponent) SupportsAdd() bool         { return false }
func (b baseComponent) Config() *config.Component { return b.config }
func (b baseComponent) Type() string              { return b.config.Type }

func (b baseComponent) SupportsRefresh() bool {
	return b.config.Data != nil && b.config.Data.RefreshInterval > 0
}

func (b baseComponent) renderHeader(border lipgloss.Border) string {
	style := lipgloss.NewStyle().Border(border, false, true, true, false).Padding(0, 1)
	title := b.config.Title
	if title == "" {
		title = "Untitled"
	}
	return style.Render(title)
}

func (b baseComponent) renderFooter(w int, percent float64, caption string) string {
	percentStr := fmt.Sprintf("%3.f%% ⥮ ", percent*100)
	percentWidth := lipgloss.Width(percentStr)
	captionStyle := lipgloss.NewStyle().Width(w - percentWidth).AlignHorizontal(lipgloss.Center).Bold(true)
	percentStyle := lipgloss.NewStyle().Width(w - captionStyle.GetWidth()).AlignHorizontal(lipgloss.Right)

	if caption != "" {
		return lipgloss.JoinHorizontal(lipgloss.Left,
			captionStyle.Render(caption),
			percentStyle.Render(percentStr),
		)
	}

	return percentStyle.Width(w).Render(percentStr)
}

type errorComponent struct {
	baseComponent
	errorMessage string
}

func newErrorComponent(base baseComponent, errMsg string) *errorComponent {
	base.err = fmt.Errorf(errMsg)
	return &errorComponent{baseComponent: base, errorMessage: errMsg}
}

func (c *errorComponent) Init() tea.Cmd                                           { return nil }
func (c *errorComponent) IsFocusable() bool                                       { return false }
func (c *errorComponent) Update(msg tea.Msg) (Component, tea.Cmd)                 { return c, nil }
func (c *errorComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) { return c, nil }
func (c *errorComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

func (c *errorComponent) View(width, height int, focused bool) string {
	style, _, border := GetBorderStyle(c.styles.Border)

	innerWidth, innerHeight := CalcWidthHeight(width, height)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	errorContent := lipgloss.Place(
		innerWidth,
		max(0, innerHeight-headerHeight),
		lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("[Error: %s]\nType: %s\nID: %s", c.errorMessage, c.config.Type, c.id),
	)

	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, errorContent)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		BorderForeground(lipgloss.Color("#ff0000")).
		Render(fullContent)
}

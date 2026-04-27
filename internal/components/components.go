package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

	Bindings() []config.Binding
	ActionActive() bool
	ActionLoading() bool

	BeginAction(label, outputAs, pendingID string) (Component, tea.Cmd)
	ApplyActionResult(pendingID, out, outputAs string) (Component, tea.Cmd)
	ResetAction() (Component, tea.Cmd)

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

// actionState captures whether a component is currently displaying the result
// of a replace_pane action instead of its normal data, or is in the middle of
// running one. The zero value means inactive/not-loading. Sticky behaviour:
// once `active` is true, the component keeps showing the action output until
// ResetAction is called (R / Esc when focused).
type actionState struct {
	active    bool
	loading   bool
	output    string
	label     string
	outputAs  string
	pendingID string // mints per-action so late-arriving results can be discarded after a reset
	spinner   spinner.Model
}

type baseComponent struct {
	err    error
	id     string
	config *config.Component
	styles *config.StyleConfig
	action actionState
}

func (b baseComponent) Init() tea.Cmd             { return nil }
func (b baseComponent) GetAddInput() string       { return "" }
func (b baseComponent) ID() string                { return b.id }
func (b baseComponent) IsFocusable() bool         { return true }
func (b baseComponent) SupportsAdd() bool         { return false }
func (b baseComponent) Config() *config.Component { return b.config }
func (b baseComponent) Type() string              { return b.config.Type }
func (b baseComponent) ActionActive() bool        { return b.action.active }
func (b baseComponent) ActionLoading() bool       { return b.action.loading }

func (b baseComponent) Bindings() []config.Binding {
	if b.config == nil {
		return nil
	}
	return b.config.Bindings
}

// beginAction transitions the component into loading state. Concrete components
// call this after cloning themselves (newInstance := *c) so the mutation is
// scoped to the new value returned to bubbletea. The returned tea.Cmd kicks
// off the first spinner tick — concrete components must include it in the
// Batch they return from BeginAction.
func (b *baseComponent) beginAction(label, outputAs, pendingID string) tea.Cmd {
	b.action.loading = true
	b.action.active = false
	b.action.label = label
	b.action.outputAs = outputAs
	b.action.pendingID = pendingID
	b.action.output = ""
	b.action.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	return b.action.spinner.Tick
}

// applyActionResult reports whether the result should be installed (true) or
// dropped because of a pendingID mismatch (false — the user reset between
// dispatch and result, or fired a different action since).
func (b *baseComponent) applyActionResult(pendingID, out string) bool {
	if pendingID != b.action.pendingID || b.action.pendingID == "" {
		return false
	}
	b.action.loading = false
	b.action.active = true
	b.action.output = out
	return true
}

func (b *baseComponent) resetAction() {
	b.action = actionState{}
}

// tickActionSpinner forwards a spinner.TickMsg to the loading spinner of a
// component. Concrete Update methods call this on a freshly cloned base; if
// handled is true they return the new instance with the returned cmd.
func tickActionSpinner(b *baseComponent, msg tea.Msg) (tea.Cmd, bool) {
	if !b.action.loading {
		return nil, false
	}
	if _, ok := msg.(spinner.TickMsg); !ok {
		return nil, false
	}
	var cmd tea.Cmd
	b.action.spinner, cmd = b.action.spinner.Update(msg)
	return cmd, true
}

// renderActionOverlay renders a component's pane while it is loading or
// showing sticky replace_pane output. Concrete View()s branch into this
// helper before doing their normal render. For output_as="inherit" plus
// active state the caller must skip the overlay (the inherit path pushes
// the action output through SetContent so the normal View renders it).
func renderActionOverlay(b baseComponent, w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(b.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}
	innerWidth, innerHeight := CalcWidthHeight(w, h)
	header := b.renderHeader(border)
	headerHeight := lipgloss.Height(header)
	contentHeight := max(0, innerHeight-headerHeight)

	var content string
	switch {
	case b.action.loading:
		label := b.action.label
		if label == "" {
			label = "action"
		}
		spinnerView := b.action.spinner.View()
		msg := spinnerView + " Running: " + label
		if innerWidth < 10 || contentHeight < 3 {
			msg = "[…]"
		}
		content = lipgloss.Place(innerWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg)
	case b.action.active:
		wrapped := WrapContent(b.action.output, innerWidth)
		lines := strings.Split(wrapped, "\n")
		if len(lines) > contentHeight {
			lines = lines[:contentHeight]
		}
		content = lipgloss.NewStyle().Width(innerWidth).Height(contentHeight).Render(strings.Join(lines, "\n"))
	}

	full := lipgloss.JoinVertical(lipgloss.Left, header, content)
	return borderStyle.Width(innerWidth).Height(innerHeight).Render(full)
}

// inActionOverlayState reports whether the component should render via the
// action overlay (loading or sticky non-inherit) instead of its normal View.
func (b baseComponent) inActionOverlayState() bool {
	if b.action.loading {
		return true
	}
	if b.action.active && b.action.outputAs != "inherit" {
		return true
	}
	return false
}

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
func (c *errorComponent) BeginAction(label, outputAs, pendingID string) (Component, tea.Cmd) {
	return c, nil
}
func (c *errorComponent) ApplyActionResult(pendingID, out, outputAs string) (Component, tea.Cmd) {
	return c, nil
}
func (c *errorComponent) ResetAction() (Component, tea.Cmd) { return c, nil }

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

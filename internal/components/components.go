package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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
		styles: config.EffectiveStyles(styles, cfg),
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
	viewport  viewport.Model
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

	// Initialise the action overlay viewport. Width/height are set during
	// render (when we know the pane size); content gets re-wrapped each
	// render in case the pane was resized.
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	vp.SetContent(out)
	b.action.viewport = vp
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

// updateActionViewport forwards key / mouse messages to the action overlay's
// scrollable viewport when the component is showing sticky text output.
// Inherit-mode action panes render through the component's normal View, so
// scrolling there belongs to the component's own viewport, not the action
// overlay.
func updateActionViewport(b *baseComponent, msg tea.Msg) (tea.Cmd, bool) {
	if !b.action.active || b.action.outputAs == "inherit" {
		return nil, false
	}
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		var cmd tea.Cmd
		b.action.viewport, cmd = b.action.viewport.Update(msg)
		return cmd, true
	}
	return nil, false
}

// renderActionOverlay renders a component's pane while it is loading or
// showing sticky replace_pane output. Concrete View()s branch into this
// helper before doing their normal render. For output_as="inherit" plus
// active state the caller must skip the overlay (the inherit path pushes
// the action output through SetContent so the normal View renders it).
//
// The pointer receiver is intentional: when active, we resize the action
// viewport to the current pane geometry and re-wrap content, mirroring how
// TextComponent.View mutates its own viewport on each render.
func renderActionOverlay(b *baseComponent, w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(b.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}
	innerWidth, innerHeight := CalcWidthHeight(w, h)
	header := b.renderHeader(border, focused, innerWidth)
	headerHeight := lipgloss.Height(header)

	var content, footer string
	footerHeight := 0

	switch {
	case b.action.loading:
		contentHeight := max(0, innerHeight-headerHeight)
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
		// reserve a footer row for scroll % + reset hint when the pane is tall
		// enough to spare it.
		if innerHeight-headerHeight > 3 {
			footerHeight = 1
		}
		viewportHeight := max(0, innerHeight-headerHeight-footerHeight)
		b.action.viewport.Width = innerWidth
		b.action.viewport.Height = viewportHeight
		b.action.viewport.SetContent(WrapContent(b.action.output, innerWidth))
		content = b.action.viewport.View()
		if footerHeight > 0 {
			footer = renderActionFooter(innerWidth, b.action.viewport.ScrollPercent(), b.action.label)
		}
	}

	parts := []string{header, content}
	if footer != "" {
		parts = append(parts, footer)
	}
	full := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return borderStyle.Width(innerWidth).Height(innerHeight).Render(full)
}

// renderActionFooter renders the bottom-of-overlay row showing scroll % on
// the right and a reset hint on the left.
func renderActionFooter(w int, percent float64, label string) string {
	hint := "↻ R · ✕ Esc"
	if label != "" {
		hint = "[" + label + "]  " + hint
	}
	percentStr := fmt.Sprintf("%3.f%% ⥮", percent*100)
	pw := lipgloss.Width(percentStr)
	hintStyle := lipgloss.NewStyle().Width(w - pw).AlignHorizontal(lipgloss.Left).Faint(true)
	pctStyle := lipgloss.NewStyle().Width(pw).AlignHorizontal(lipgloss.Right).Faint(true)
	return lipgloss.JoinHorizontal(lipgloss.Left, hintStyle.Render(hint), pctStyle.Render(percentStr))
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

func (b baseComponent) renderHeader(border lipgloss.Border, focused bool, innerWidth int) string {
	title := b.config.Title
	if title == "" {
		title = "Untitled"
	}
	if b.config.Icon != "" {
		title = b.config.Icon + " " + title
	}

	// Border sides on the title cell depend on the title's alignment so the
	// cell visually anchors to the side of the pane the title sits on:
	//   left   → right + bottom         (anchored to the left edge)
	//   center → left + right + bottom  (free-floating)
	//   right  → left + bottom          (anchored to the right edge)
	align := b.config.TitleAlign
	right, bottom, left := true, true, false
	switch align {
	case "center":
		right, bottom, left = true, true, true
	case "right":
		right, bottom, left = false, true, true
	}

	titleStyle := lipgloss.NewStyle()
	style := lipgloss.NewStyle().Border(border, false, right, bottom, left).Padding(0, 1)

	if focused {
		titleStyle = titleStyle.Bold(true)
		if b.styles != nil && b.styles.Border != nil {
			color := GetColor(b.styles.Border.FocusedColor, lipgloss.Color("default"))
			titleStyle = titleStyle.Foreground(color)
			if right {
				style = style.BorderRightForeground(color)
			}
			if bottom {
				style = style.BorderBottomForeground(color)
			}
			if left {
				style = style.BorderLeftForeground(color)
			}
		}
	}

	cell := style.Render(titleStyle.Render(title))

	if align == "" || innerWidth <= 0 {
		return cell
	}
	cellWidth := lipgloss.Width(cell)
	if cellWidth >= innerWidth {
		return cell
	}
	switch align {
	case "center":
		return lipgloss.PlaceHorizontal(innerWidth, lipgloss.Center, cell)
	case "right":
		return lipgloss.PlaceHorizontal(innerWidth, lipgloss.Right, cell)
	default: // "left"
		return cell
	}
}

func (b baseComponent) renderFooter(w int, percent float64, caption string) string {
	dim := PaletteColor(b.styles, "dim", lipgloss.Color("#808080"))
	percentStr := fmt.Sprintf("%3.f%% ⥮ ", percent*100)
	percentWidth := lipgloss.Width(percentStr)
	captionStyle := lipgloss.NewStyle().Width(w - percentWidth).AlignHorizontal(lipgloss.Center).Foreground(dim)
	percentStyle := lipgloss.NewStyle().Width(w - captionStyle.GetWidth()).AlignHorizontal(lipgloss.Right).Foreground(dim).Faint(true)

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

	header := c.renderHeader(border, focused, innerWidth)
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

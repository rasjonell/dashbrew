package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// helpLine styles. Keep them as package-level vars so the rendering of every
// frame doesn't reallocate them.
var (
	helpKeyStyle   = lipgloss.NewStyle()
	helpLabelStyle = lipgloss.NewStyle().Faint(true)
	helpSepStyle   = lipgloss.NewStyle().Faint(true)
	helpRuleColor  = lipgloss.Color("8") // bright black / grey
)

// renderHelpLine produces the two rows reserved at the bottom of the terminal:
// a thin separator rule above and the contextual help line below. Contents
// adapt to the current mode (error modal, palette, add, action loading or
// active, default dashboard) and to the focused component's declared bindings.
func (m *model) renderHelpLine() string {
	text := m.helpLineText()

	contentWidth := max(0, m.width-2)
	style := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(helpRuleColor)

	return style.Render(truncateInline(text, contentWidth))
}

func (m *model) helpLineText() string {
	if m.errorModal != nil {
		return helpItem("Esc", "dismiss")
	}
	if m.palette.open {
		return joinDotPlain(
			helpItem("↑↓", "nav"),
			helpItem("/", "filter"),
			helpItem("Enter", "run"),
			helpItem("Ctrl+K", "close"),
		)
	}
	if m.isAdding {
		return joinDotPlain(
			helpItem("Esc", "cancel"),
			helpItem("Backspace", "delete"),
			helpItem("Enter", "submit"),
		)
	}

	globals := []string{
		helpItem("^C", "quit"),
		helpItem("Shift+arrows", "nav"),
		helpItem("Ctrl+K", "palette"),
	}

	comp, ok := m.components[m.focusedComponentId]
	if !ok {
		return joinDotPlain(globals...)
	}

	if comp.ActionLoading() {
		return joinDotPlain(
			helpItem("^C", "quit"),
			helpItem("Esc", "cancel"),
			helpItem("R", "cancel + refetch"),
		)
	}
	if comp.ActionActive() {
		return joinDotPlain(append(globals,
			helpItem("↑↓", "scroll"),
			helpItem("Esc", "clear"),
			helpItem("R", "refetch"),
		)...)
	}

	if comp.SupportsRefresh() {
		globals = append(globals, helpItem("R", "refresh"))
	}
	if comp.SupportsAdd() {
		globals = append(globals, helpItem("A", "add"))
	}
	base := joinDotPlain(globals...)

	bindings := comp.Bindings()
	if len(bindings) == 0 {
		return base
	}

	compItems := make([]string, 0, len(bindings))
	for _, b := range bindings {
		label := b.Label
		if label == "" {
			label = b.Action.Type
		}
		compItems = append(compItems, helpItem(b.Key, label))
	}

	title := ""
	if cfg := comp.Config(); cfg != nil {
		title = strings.TrimSpace(cfg.Title)
	}
	if title == "" {
		title = m.focusedComponentId
	}

	return base +
		helpSepStyle.Render("  │  ") +
		helpKeyStyle.Bold(true).Render(title) +
		helpLabelStyle.Render(": ") +
		joinDotPlain(compItems...)
}

// helpItem styles a single key/label pair: <key in normal weight> <label faint>.
func helpItem(key, label string) string {
	return helpKeyStyle.Render(key) + " " + helpLabelStyle.Render(label)
}

// joinDotPlain joins pre-styled segments with a faint dot separator.
func joinDotPlain(segments ...string) string {
	return strings.Join(segments, helpSepStyle.Render(" · "))
}

// truncateInline trims s with an ellipsis so its printable width fits in w.
// ansi.Truncate handles styled segments (escape codes are preserved across
// the cut) and grapheme widths correctly.
func truncateInline(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	return ansi.Truncate(s, w, "…")
}

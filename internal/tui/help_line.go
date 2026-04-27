package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpLine produces the single-row help line that always sits at the
// bottom of the terminal. The contents adapt to the current mode (error
// modal, palette, add, action loading/active, default dashboard) and to the
// focused component's declared bindings.
func (m *model) renderHelpLine() string {
	text := m.helpLineText()
	style := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Faint(true)
	avail := max(0, m.width-2)
	return style.Render(truncateInline(text, avail))
}

func (m *model) helpLineText() string {
	if m.errorModal != nil {
		return "Esc dismiss"
	}
	if m.palette.open {
		return joinDot("↑↓ nav", "/ filter", "Enter run", "Ctrl+K close")
	}
	if m.isAdding {
		return joinDot("Esc cancel", "Backspace delete", "Enter submit")
	}

	globals := []string{"^C quit", "Shift+arrows nav", "Ctrl+K palette"}

	comp, ok := m.components[m.focusedComponentId]
	if !ok {
		return joinDot(globals...)
	}

	if comp.ActionLoading() {
		return joinDot("^C quit", "Esc cancel", "R cancel + refetch")
	}
	if comp.ActionActive() {
		return joinDot("^C quit", "↑↓ scroll", "Esc clear", "R refetch")
	}

	if comp.SupportsRefresh() {
		globals = append(globals, "R refresh")
	}
	if comp.SupportsAdd() {
		globals = append(globals, "A add")
	}
	base := joinDot(globals...)

	bindings := comp.Bindings()
	if len(bindings) == 0 {
		return base
	}

	bs := make([]string, 0, len(bindings))
	for _, b := range bindings {
		label := b.Label
		if label == "" {
			label = b.Action.Type
		}
		bs = append(bs, fmt.Sprintf("%s %s", b.Key, label))
	}

	title := ""
	if cfg := comp.Config(); cfg != nil {
		title = strings.TrimSpace(cfg.Title)
	}
	if title == "" {
		title = m.focusedComponentId
	}

	return base + "  │  " + title + ": " + joinDot(bs...)
}

func joinDot(items ...string) string {
	return strings.Join(items, " · ")
}

// truncateInline trims s with an ellipsis so its printable width fits in w.
// Uses lipgloss.Width to handle ANSI escapes and wide runes correctly.
func truncateInline(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	out := make([]rune, 0, len(s))
	width := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if width+rw > w-1 {
			break
		}
		out = append(out, r)
		width += rw
	}
	return string(out) + "…"
}

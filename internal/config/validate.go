package config

import (
	"fmt"
	"slices"
	"strings"
)

var reservedKeys = map[string]struct{}{
	"ctrl+c":     {},
	"ctrl+k":     {},
	"shift+up":   {},
	"shift+down": {},
	"shift+left": {},
	"shift+right": {},
	"H":          {},
	"J":          {},
	"K":          {},
	"L":          {},
	"a":          {},
	"A":          {},
	"r":          {},
	"R":          {},
	"esc":        {},
}

var validActionTypes = map[string]struct{}{
	"fire_and_forget": {},
	"replace_pane":    {},
}

var validOutputAs = map[string]struct{}{
	"":        {}, // empty = default ("text")
	"text":    {},
	"inherit": {},
}

// internalKeysByType lists the keys each component type consumes from its own
// Update loop. Bindings on a component cannot use any of these (they would be
// shadowed by component-internal handling).
//
// Keep in sync with each component's Update method:
//   - todo:      internal/components/todo.go
//   - text:      internal/components/text.go (viewport scroll)
//   - histogram: internal/components/histogram.go (viewport scroll)
//   - list:      internal/components/list.go (bubbles/list nav + filter)
//   - table:     internal/components/table.go (bubbles/table nav)
//   - chart:     no internal keys
var internalKeysByType = map[string][]string{
	"todo":      {" ", "delete", "d", "D", "enter", "backspace"},
	"text":      {"up", "down", "pgup", "pgdown", "home", "end"},
	"histogram": {"up", "down", "pgup", "pgdown", "home", "end"},
	"list":      {"up", "down", "/", "enter"},
	"table":     {"up", "down", "pgup", "pgdown", "home", "end"},
	"chart":     {},
}

// keyAliases lets users write friendly names that we expand to the literal
// strings bubbletea emits via tea.KeyMsg.String().
var keyAliases = map[string]string{
	"space": " ",
}

// NormalizeKey canonicalizes a key string for matching against tea.KeyMsg.String().
// Modifiers (ctrl/alt/shift) are lowercased and reordered to ctrl+alt+shift+key.
// The trailing key segment is preserved with its original case so single-letter
// shifted keys like "H" remain distinct from "h". Aliases like "space" expand
// to " " (the literal space character bubbletea emits).
func NormalizeKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if alias, ok := keyAliases[strings.ToLower(s)]; ok {
		return alias
	}
	parts := strings.Split(s, "+")
	if len(parts) == 1 {
		return s
	}
	var ctrl, alt, shift bool
	for _, p := range parts[:len(parts)-1] {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case "ctrl":
			ctrl = true
		case "alt":
			alt = true
		case "shift":
			shift = true
		default:
			return s // unknown modifier — leave opaque, validator will reject below
		}
	}
	key := strings.TrimSpace(parts[len(parts)-1])
	if alias, ok := keyAliases[strings.ToLower(key)]; ok {
		key = alias
	}
	var b strings.Builder
	if ctrl {
		b.WriteString("ctrl+")
	}
	if alt {
		b.WriteString("alt+")
	}
	if shift {
		b.WriteString("shift+")
	}
	b.WriteString(key)
	return b.String()
}

// Validate runs all v2 config invariants. It is called from LoadConfig after
// defaults are filled, so any error returned here surfaces at startup.
func (cfg *DashboardConfig) Validate() error {
	if cfg.Layout == nil {
		return fmt.Errorf("layout: missing")
	}
	return validateNode(cfg.Layout, "layout")
}

func validateNode(node *LayoutNode, path string) error {
	if node == nil {
		return nil
	}
	switch node.Type {
	case "component":
		if node.Component != nil {
			if err := validateComponent(node.Component, path+".component"); err != nil {
				return err
			}
		}
	case "container":
		for i, child := range node.Children {
			if err := validateNode(child, fmt.Sprintf("%s.children[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateComponent(c *Component, path string) error {
	seenKeys := make(map[string]int, len(c.Bindings))
	for i, b := range c.Bindings {
		bPath := fmt.Sprintf("%s.bindings[%d]", path, i)
		if err := validateBinding(b, c.Type, bPath); err != nil {
			return err
		}
		nk := NormalizeKey(b.Key)
		if prev, ok := seenKeys[nk]; ok {
			return fmt.Errorf("%s.key %q: duplicate of bindings[%d] on the same component", bPath, b.Key, prev)
		}
		seenKeys[nk] = i
	}
	return nil
}

func validateBinding(b Binding, componentType, path string) error {
	if strings.TrimSpace(b.Key) == "" {
		return fmt.Errorf("%s.key: must not be empty", path)
	}
	nk := NormalizeKey(b.Key)
	if _, reserved := reservedKeys[nk]; reserved {
		return fmt.Errorf("%s.key %q: reserved by dashbrew", path, b.Key)
	}
	if slices.Contains(internalKeysByType[componentType], nk) {
		return fmt.Errorf("%s.key %q: conflicts with internal key for component type %q", path, b.Key, componentType)
	}
	return validateAction(b.Action, componentType, path+".action")
}

func validateAction(a ActionConfig, componentType, path string) error {
	if _, ok := validActionTypes[a.Type]; !ok {
		return fmt.Errorf("%s.type %q: must be %q or %q", path, a.Type, "fire_and_forget", "replace_pane")
	}
	if strings.TrimSpace(a.Command) == "" {
		return fmt.Errorf("%s.command: must not be empty", path)
	}
	if _, ok := validOutputAs[a.OutputAs]; !ok {
		return fmt.Errorf("%s.output_as %q: must be %q or %q", path, a.OutputAs, "text", "inherit")
	}
	if a.OutputAs == "inherit" && a.Type == "fire_and_forget" {
		return fmt.Errorf("%s.output_as %q: only meaningful when type is %q", path, a.OutputAs, "replace_pane")
	}
	if a.OutputAs == "inherit" && componentType == "todo" {
		return fmt.Errorf("%s.output_as %q: not supported for component type %q (todo's data path is file-based)", path, a.OutputAs, componentType)
	}
	if a.TimeoutSeconds < 0 {
		return fmt.Errorf("%s.timeout_seconds %d: must be >= 0", path, a.TimeoutSeconds)
	}
	return nil
}

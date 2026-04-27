package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed themes/*.json
var embeddedThemes embed.FS

// Theme is a named palette preset, loaded from JSON either at build time
// (embedded under internal/config/themes) or from ~/.config/dashbrew/themes
// at startup (handled in LoadCustomThemes).
type Theme struct {
	Name    string         `json:"name"`
	Palette *PaletteConfig `json:"palette"`
}

var themeRegistry = map[string]Theme{}

func init() {
	entries, err := embeddedThemes.ReadDir("themes")
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := embeddedThemes.ReadFile("themes/" + e.Name())
		if err != nil {
			continue
		}
		var t Theme
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		if t.Name == "" {
			t.Name = strings.TrimSuffix(e.Name(), ".json")
		}
		themeRegistry[t.Name] = t
	}
}

// RegisterTheme adds or overrides a theme in the registry. Used to inject
// user-provided themes loaded from disk.
func RegisterTheme(t Theme) {
	if t.Name == "" {
		return
	}
	themeRegistry[t.Name] = t
}

// LookupTheme returns a copy of the named theme along with whether it exists.
func LookupTheme(name string) (Theme, bool) {
	t, ok := themeRegistry[name]
	return t, ok
}

// AvailableThemes returns the sorted list of currently-registered theme names.
func AvailableThemes() []string {
	names := make([]string, 0, len(themeRegistry))
	for n := range themeRegistry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// resolvePalette merges palettes by precedence:
//
//  1. built-in default theme (always populated)
//  2. named theme from style.theme (if set and known)
//  3. explicit style.palette (user overrides)
//  4. legacy style.border.color / focusedColor (overlay onto palette.border /
//     borderFocused so existing v1 configs still drive the resolved palette)
//
// The result is written back into cfg.Style.Palette so components can read it
// from a single place.
func (cfg *DashboardConfig) resolvePalette() error {
	if cfg.Style == nil {
		cfg.Style = &StyleConfig{}
	}

	resolved := &PaletteConfig{}

	// 1. defaults
	if def, ok := LookupTheme("default"); ok && def.Palette != nil {
		mergePaletteInto(resolved, def.Palette)
	}

	// 2. named theme
	if name := strings.TrimSpace(cfg.Style.Theme); name != "" {
		t, ok := LookupTheme(name)
		if !ok {
			return fmt.Errorf("style.theme %q: unknown theme (available: %s)",
				cfg.Style.Theme, strings.Join(AvailableThemes(), ", "))
		}
		if t.Palette != nil {
			mergePaletteInto(resolved, t.Palette)
		}
	}

	// 3. explicit palette overrides
	if cfg.Style.Palette != nil {
		mergePaletteInto(resolved, cfg.Style.Palette)
	}

	// 4. legacy border field overrides
	if cfg.Style.Border != nil {
		if cfg.Style.Border.Color != "" {
			resolved.Border = cfg.Style.Border.Color
		}
		if cfg.Style.Border.FocusedColor != "" {
			resolved.BorderFocused = cfg.Style.Border.FocusedColor
		}
	}

	cfg.Style.Palette = resolved
	// Mirror the resolved values back into the legacy border block so callers
	// that still read from there continue to work without changes.
	if cfg.Style.Border == nil {
		cfg.Style.Border = &BorderStyleConfig{}
	}
	if cfg.Style.Border.Color == "" {
		cfg.Style.Border.Color = resolved.Border
	}
	if cfg.Style.Border.FocusedColor == "" {
		cfg.Style.Border.FocusedColor = resolved.BorderFocused
	}
	return nil
}

// mergePaletteInto copies non-empty fields from src onto dst.
func mergePaletteInto(dst, src *PaletteConfig) {
	if src == nil {
		return
	}
	if src.Background != "" {
		dst.Background = src.Background
	}
	if src.Foreground != "" {
		dst.Foreground = src.Foreground
	}
	if src.Dim != "" {
		dst.Dim = src.Dim
	}
	if src.Accent != "" {
		dst.Accent = src.Accent
	}
	if src.Success != "" {
		dst.Success = src.Success
	}
	if src.Warning != "" {
		dst.Warning = src.Warning
	}
	if src.Error != "" {
		dst.Error = src.Error
	}
	if src.Border != "" {
		dst.Border = src.Border
	}
	if src.BorderFocused != "" {
		dst.BorderFocused = src.BorderFocused
	}
}

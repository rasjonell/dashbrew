package config

import (
	"encoding/json"
	"os"
)

type DashboardConfig struct {
	Layout *LayoutNode  `json:"layout"`
	Style  *StyleConfig `json:"style,omitempty"`
}

type LayoutNode struct {
	Type      string `json:"type"`
	Flex      int    `json:"flex,omitempty"`
	Direction string `json:"direction,omitempty"`

	Children  []*LayoutNode `json:"children,omitempty"`
	Component *Component    `json:"component,omitempty"`
}

type Component struct {
	Type       string                `json:"type"`
	Title      string                `json:"title"`
	TitleAlign string                `json:"titleAlign,omitempty"` // "left" (default) | "center" | "right"
	Icon       string                `json:"icon,omitempty"`       // literal glyph prepended to title
	Data       *DataConfig           `json:"data"`
	ID         string                `json:"id,omitempty"`
	Bindings   []Binding             `json:"bindings,omitempty"`
	Style      *ComponentStyleConfig `json:"style,omitempty"`
}

// ComponentStyleConfig carries per-component visual overrides on top of the
// dashboard-wide theme. Generic fields apply to all components; named fields
// are component-type specific.
type ComponentStyleConfig struct {
	// Generic overrides (any component)
	Palette    *PaletteConfig `json:"palette,omitempty"`    // override role colors for this pane only
	BorderType string         `json:"borderType,omitempty"` // override border style for this pane only

	// Histogram
	BarStyle  string            `json:"barStyle,omitempty"`  // "single" | "alternate" | "gradient" | "map"
	BarColors map[string]string `json:"barColors,omitempty"` // per-bin overrides for "map" style
}

// EffectiveStyles returns a per-component style config: a clone of the
// dashboard-wide style with the component's own style block layered on top.
// Used by the components package when constructing each pane.
func EffectiveStyles(global *StyleConfig, comp *Component) *StyleConfig {
	if global == nil {
		global = &StyleConfig{}
	}
	eff := &StyleConfig{
		Theme:   global.Theme,
		Global:  global.Global,
		Palette: clonePalette(global.Palette),
		Border:  cloneBorder(global.Border),
	}
	if comp == nil || comp.Style == nil {
		return eff
	}
	if comp.Style.Palette != nil {
		if eff.Palette == nil {
			eff.Palette = &PaletteConfig{}
		}
		mergePaletteInto(eff.Palette, comp.Style.Palette)
	}
	if comp.Style.BorderType != "" {
		if eff.Border == nil {
			eff.Border = &BorderStyleConfig{}
		}
		eff.Border.Type = comp.Style.BorderType
	}
	return eff
}

func clonePalette(p *PaletteConfig) *PaletteConfig {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}

func cloneBorder(b *BorderStyleConfig) *BorderStyleConfig {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}

type Binding struct {
	Key    string       `json:"key"`
	Label  string       `json:"label"`
	Action ActionConfig `json:"action"`
}

type ActionConfig struct {
	Type           string `json:"type"`
	Command        string `json:"command"`
	OutputAs       string `json:"output_as,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

type DataConfig struct {
	Source          string          `json:"source"`
	JSONPath        string          `json:"json_path"`
	X               string          `json:"x,omitempty"`
	Y               string          `json:"y,omitempty"`
	URL             string          `json:"url,omitempty"`
	Command         string          `json:"command,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	Columns         []*ColumnConfig `json:"columns,omitempty"`
	RefreshMode     string          `json:"refresh_mode,omitempty"`
	RefreshInterval int             `json:"refresh_interval,omitempty"`
}

type ColumnConfig struct {
	Label string `json:"label"`
	Field string `json:"field,omitempty"`
	Flex  int    `json:"flex,omitempty"`
}

type StyleConfig struct {
	Theme   string             `json:"theme,omitempty"`
	Palette *PaletteConfig     `json:"palette,omitempty"`
	Global  *GlobalStyleConfig `json:"global,omitempty"`
	Border  *BorderStyleConfig `json:"border,omitempty"`
}

// PaletteConfig is the role-based color palette. Each field is a hex string
// (e.g. "#ff5555") or empty to fall back to the active theme's value, then
// to the built-in default. Wired by Resolve() at config-load time.
type PaletteConfig struct {
	Background    string `json:"background,omitempty"`
	Foreground    string `json:"foreground,omitempty"`
	Dim           string `json:"dim,omitempty"`
	Accent        string `json:"accent,omitempty"`
	Success       string `json:"success,omitempty"`
	Warning       string `json:"warning,omitempty"`
	Error         string `json:"error,omitempty"`
	Border        string `json:"border,omitempty"`
	BorderFocused string `json:"borderFocused,omitempty"`
}

type GlobalStyleConfig struct {
	TextColor        string `json:"textColor,omitempty"`
	HighlightedColor string `json:"highlightedColor,omitempty"`
}

type BorderStyleConfig struct {
	Type         string `json:"type,omitempty"`
	Color        string `json:"color,omitempty"`
	FocusedColor string `json:"focusedColor,omitempty"`
}

func LoadConfig(path string) (*DashboardConfig, error) {
	loadCustomThemesOnce()

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg *DashboardConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	if cfg.Style == nil {
		cfg.Style = &StyleConfig{}
	}

	if cfg.Style.Global == nil {
		cfg.Style.Global = &GlobalStyleConfig{}
	}

	if cfg.Style.Border == nil {
		cfg.Style.Border = &BorderStyleConfig{}
	}

	if err := cfg.resolvePalette(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

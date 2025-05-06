package config

import (
	"encoding/json"
	"os"
)

type DashboardConfig struct {
	Layout *LayoutNode  `json:"layout"`
	Style  *StyleConfig `json:"style,omitempty"`
}

type StyleConfig struct {
	BorderType         string `json:"borderType,omitempty"`
	BorderColor        string `json:"BorderColor,omitempty"`
	FocusedBorderColor string `json:"focusedBorderColor,omitempty"`
}

type LayoutNode struct {
	Type      string `json:"type"`
	Flex      int    `json:"flex,omitempty"`
	Direction string `json:"direction,omitempty"`

	Children  []*LayoutNode `json:"children,omitempty"`
	Component *Component    `json:"component,omitempty"`
}

type Component struct {
	Type  string      `json:"type"`
	Title string      `json:"title"`
	Data  *DataConfig `json:"data"`
	ID    string      `json:"id,omitempty"`
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

func LoadConfig(path string) (*DashboardConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg *DashboardConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

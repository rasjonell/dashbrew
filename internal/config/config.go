package config

import (
	"encoding/json"
	"os"
)

type DashboardConfig struct {
	Style  *StyleConfig `json:"style,omitempty"`
	Layout *LayoutNode  `json:"layout"`
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
	Source          string `json:"source"`
	X               string `json:"x,omitempty"`
	Y               string `json:"y,omitempty"`
	URL             string `json:"url,omitempty"`
	Command         string `json:"command,omitempty"`
	RefreshInterval int    `json:"refresh_interval,omitempty"`
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

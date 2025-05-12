package components

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/data"
)

type HistogramComponent struct {
	baseComponent
	maxValue int
	content  string
	labels   []string
	bins     map[string]int
	viewport viewport.Model
}

func newHistogramComponent(base baseComponent) *HistogramComponent {
	vp := viewport.New(0, 0)
	vp.SetContent("[loading...]")
	vp.YPosition = 1
	vp.MouseWheelEnabled = true

	return &HistogramComponent{
		baseComponent: base,
		maxValue:      0,
		viewport:      vp,
		content:       "",
		labels:        make([]string, 0),
		bins:          make(map[string]int),
	}
}

func (c *HistogramComponent) View(w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	footerHeight := 0
	if !(c.viewport.AtTop() && c.viewport.AtBottom()) {
		footerHeight = 1
	}
	var footer string
	if footerHeight > 0 {
		footer = c.renderFooter(c.viewport.Width, c.viewport.ScrollPercent(), c.config.Data.Caption)
	}

	c.viewport.Width = innerWidth
	c.viewport.Height = max(0, innerHeight-headerHeight-footerHeight)

	if c.err != nil {
		c.content = fmt.Sprintf("[Error fetching/parsing data]\n%s", c.err.Error())
	} else if len(c.bins) == 0 {
		c.content = "[Loading or No Data]"
	} else if innerWidth < 5 {
		c.content = "[Area Too Small]"
	} else {
		c.content = c.renderHistogram(innerWidth, footerHeight > 0)
	}

	c.viewport.SetContent(c.content)

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		c.viewport.View(),
		footer,
	)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *HistogramComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		c.viewport, cmd = c.viewport.Update(msg)
		return c, cmd
	}

	return c, nil
}

func (c *HistogramComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

func (c *HistogramComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c

	if result.Error() != nil {
		newInstance.err = result.Error()
		newInstance.bins = nil
		newInstance.maxValue = 0
		newInstance.labels = nil
	} else {
		parsedBins, parseErr := c.parseDataToHistogram(result.Output())
		if parseErr != nil {
			newInstance.err = parseErr
			newInstance.bins = nil
			newInstance.maxValue = 0
			newInstance.labels = nil
		} else {
			newInstance.err = nil
			newInstance.bins = parsedBins

			newInstance.maxValue = 0
			newInstance.labels = make([]string, 0, len(parsedBins))
			for k, v := range parsedBins {
				newInstance.labels = append(newInstance.labels, k)
				newInstance.maxValue = max(newInstance.maxValue, v)
			}
			sort.Strings(newInstance.labels)
		}
	}

	return &newInstance, nil
}

func (c *HistogramComponent) renderHistogram(w int, withFooter bool) string {
	baseColor := GetColor(c.styles.Border.FocusedColor, "#484848")
	barColor := BrightenColor(baseColor, 30)
	barStyle := lipgloss.NewStyle().Foreground(barColor)
	lineStyle := lipgloss.NewStyle().MarginBottom(1)
	captionStyle := lipgloss.NewStyle().Width(w).AlignHorizontal(lipgloss.Center).Bold(true).MarginTop(1)

	maxLabelLength := 10
	for _, key := range c.labels {
		if len(key) > maxLabelLength {
			maxLabelLength = min(len(key), 20)
		}
	}

	lines := make([]string, 0, len(c.labels))
	for _, key := range c.labels {
		count := c.bins[key]
		label := key
		if len(label) > maxLabelLength {
			label = label[:maxLabelLength-3] + "..."
		}

		countText := fmt.Sprintf("%d", count)

		maxBarLength := w - maxLabelLength - len(countText) - 7
		barLength := max(1, int(float64(count)/float64(c.maxValue)*float64(maxBarLength)))

		bar := strings.Repeat("█", barLength)

		paddedLabel := fmt.Sprintf("%-*s", maxLabelLength, label)
		line := lineStyle.Render(fmt.Sprintf("%s [%s] | %s", paddedLabel, countText, barStyle.Render(bar)))
		lines = append(lines, line)
	}

	if !withFooter && c.config.Data.Caption != "" {
		lines = append(lines, captionStyle.Render(c.config.Data.Caption))
	}

	return strings.Join(lines, "\n")
}

func (c *HistogramComponent) parseDataToHistogram(rawData string) (map[string]int, error) {
	var jsonData any
	err := json.Unmarshal([]byte(rawData), &jsonData)

	if err == nil {
		if jsonMap, ok := jsonData.(map[string]any); ok {
			result := make(map[string]int)
			for k, v := range jsonMap {
				if numVal, ok := v.(float64); ok {
					result[k] = int(numVal)
				} else {
					result[k] = 1
				}
			}
			return result, nil
		}

		if jsonArray, ok := jsonData.([]any); ok {
			result := make(map[string]int)
			for _, v := range jsonArray {
				result[fmt.Sprintf("%v", v)]++
			}
			return result, nil
		}

		return nil, fmt.Errorf("Unsupported JSON format. Expected top-level {[key]: number} or Array of primitive values")
	}

	lines := strings.Split(strings.TrimSpace(rawData), "\n")
	result := make(map[string]int)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			vStr := strings.TrimSpace(parts[1])
			v, err := strconv.Atoi(vStr)
			if err != nil {
				result[k]++
			} else {
				result[k] += v
			}
		} else {
			result[line]++
		}
	}

	return result, nil
}

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
	if c.inActionOverlayState() {
		return renderActionOverlay(&c.baseComponent, w, h, focused)
	}

	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border, focused, innerWidth)
	headerHeight := lipgloss.Height(header)

	// Decide whether the footer (caption + scroll %) is needed based on this
	// frame's content rather than the viewport's stale dimensions from the
	// previous render — otherwise the first frame after data arrives uses the
	// previous frame's no-data state and gets the height calculation wrong.
	c.viewport.Width = innerWidth

	var content string
	switch {
	case c.err != nil:
		content = fmt.Sprintf("[Error fetching/parsing data]\n%s", c.err.Error())
	case len(c.bins) == 0:
		content = "[Loading or No Data]"
	case innerWidth < 5:
		content = "[Area Too Small]"
	default:
		// Render with caption inline; we'll re-render without the inline caption
		// below if a footer ends up being shown.
		content = c.renderHistogram(innerWidth, false)
	}

	contentHeight := lipgloss.Height(content)
	available := max(0, innerHeight-headerHeight)
	footerHeight := 0
	if contentHeight > available {
		footerHeight = 1
		// re-render without the inline caption since the footer renders it
		if c.err == nil && len(c.bins) > 0 && innerWidth >= 5 {
			content = c.renderHistogram(innerWidth, true)
		}
	}

	c.content = content
	c.viewport.Height = max(0, innerHeight-headerHeight-footerHeight)
	c.viewport.SetContent(c.content)

	parts := []string{header, c.viewport.View()}
	if footerHeight > 0 {
		parts = append(parts, c.renderFooter(c.viewport.Width, c.viewport.ScrollPercent(), c.config.Data.Caption))
	}
	fullContent := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *HistogramComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	newInstance := *c
	if cmd, handled := tickActionSpinner(&newInstance.baseComponent, msg); handled {
		return &newInstance, cmd
	}
	if cmd, handled := updateActionViewport(&newInstance.baseComponent, msg); handled {
		return &newInstance, cmd
	}

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

func (c *HistogramComponent) BeginAction(label, outputAs, pendingID string) (Component, tea.Cmd) {
	newInstance := *c
	tick := newInstance.beginAction(label, outputAs, pendingID)
	return &newInstance, tick
}

func (c *HistogramComponent) ApplyActionResult(pendingID, out, outputAs string) (Component, tea.Cmd) {
	newInstance := *c
	if !newInstance.applyActionResult(pendingID, out) {
		return c, nil
	}
	return &newInstance, nil
}

func (c *HistogramComponent) ResetAction() (Component, tea.Cmd) {
	newInstance := *c
	newInstance.resetAction()
	return &newInstance, nil
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
	lineStyle := lipgloss.NewStyle().MarginBottom(1)
	captionStyle := lipgloss.NewStyle().Width(w).AlignHorizontal(lipgloss.Center).Bold(true).MarginTop(1)

	maxLabelLength := 10
	for _, key := range c.labels {
		if len(key) > maxLabelLength {
			maxLabelLength = min(len(key), 20)
		}
	}

	lines := make([]string, 0, len(c.labels))
	for i, key := range c.labels {
		count := c.bins[key]
		label := key
		if len(label) > maxLabelLength {
			label = label[:maxLabelLength-3] + "..."
		}

		countText := fmt.Sprintf("%d", count)

		maxBarLength := w - maxLabelLength - len(countText) - 7
		barLength := max(1, int(float64(count)/float64(c.maxValue)*float64(maxBarLength)))

		bar := strings.Repeat("█", barLength)
		barColor := c.barColor(key, count, i)
		barStyle := lipgloss.NewStyle().Foreground(barColor)

		paddedLabel := fmt.Sprintf("%-*s", maxLabelLength, label)
		line := lineStyle.Render(fmt.Sprintf("%s [%s] | %s", paddedLabel, countText, barStyle.Render(bar)))
		lines = append(lines, line)
	}

	if !withFooter && c.config.Data.Caption != "" {
		lines = append(lines, captionStyle.Render(c.config.Data.Caption))
	}

	return strings.Join(lines, "\n")
}

// barColor picks a bar's foreground color based on the configured barStyle.
// Defaults to the theme accent. Modes:
//   - "single" (default) — every bar uses theme.accent
//   - "alternate"        — alternates accent / dim
//   - "gradient"         — interpolates dim -> accent based on value/maxValue
//   - "map"              — uses style.barColors[key], falls back to accent
func (c *HistogramComponent) barColor(key string, count, idx int) lipgloss.Color {
	accent := PaletteColor(c.styles, "accent", lipgloss.Color("#bd93f9"))
	dim := PaletteColor(c.styles, "dim", lipgloss.Color("#808080"))

	mode := ""
	var perBin map[string]string
	if cs := c.config.Style; cs != nil {
		mode = cs.BarStyle
		perBin = cs.BarColors
	}

	switch mode {
	case "alternate":
		if idx%2 == 1 {
			return dim
		}
		return accent
	case "gradient":
		if c.maxValue <= 0 {
			return accent
		}
		t := float64(count) / float64(c.maxValue)
		return blendHexColors(string(dim), string(accent), t, accent)
	case "map":
		if hex, ok := perBin[key]; ok {
			return GetColor(hex, accent)
		}
		return accent
	default: // "" or "single"
		return accent
	}
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

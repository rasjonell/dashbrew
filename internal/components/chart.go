package components

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
	"github.com/rasjonell/dashbrew/internal/data"
)

type ChartComponent struct {
	baseComponent
	plotData []float64
}

func newChartComponent(base baseComponent) *ChartComponent {
	return &ChartComponent{
		baseComponent: base,
		plotData:      nil,
	}
}

func (c *ChartComponent) View(w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	chartHeight := max(0, innerHeight-headerHeight-2)
	chartWidth := max(0, innerWidth-8)

	var chartContent string
	if c.err != nil {
		errorMsg := fmt.Sprintf("[Error fetching/parsing data]\n%s", c.err.Error())
		chartContent = lipgloss.Place(chartWidth, chartHeight,
			lipgloss.Center, lipgloss.Center,
			WrapContent(errorMsg, chartWidth),
		)
	} else if len(c.plotData) == 0 {
		chartContent = lipgloss.Place(chartWidth, chartHeight,
			lipgloss.Center, lipgloss.Center,
			"[Loading or No Data]",
		)
	} else {
		if chartHeight < 1 || chartWidth < 5 {
			chartContent = lipgloss.Place(chartWidth, chartHeight,
				lipgloss.Center, lipgloss.Center,
				"[Area Too Small]",
			)
		} else {
			opts := []asciigraph.Option{
				asciigraph.Width(chartWidth),
				asciigraph.Height(chartHeight),
			}
			if c.config.Data != nil && c.config.Data.Caption != "" {
				opts = append(opts, asciigraph.Caption(c.config.Data.Caption))
			}

			chartContent = asciigraph.Plot(c.plotData, opts...)
		}
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		chartContent,
	)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *ChartComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

func (c *ChartComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c

	if result.Error() != nil {
		newInstance.err = result.Error()
		newInstance.plotData = nil
	} else {
		parsedData, parseErr := c.parseDataToChartPoints(result.Output())
		if parseErr != nil {
			newInstance.err = fmt.Errorf("failed to parse chart data %w", parseErr)
			newInstance.plotData = nil
		} else {
			newInstance.err = nil
			if c.config.Data.RefreshMode == "append" {
				newInstance.plotData = append(newInstance.plotData, parsedData...)
			} else {
				newInstance.plotData = parsedData
			}
		}
	}

	return &newInstance, nil
}

func (c *ChartComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

func (c *ChartComponent) parseDataToChartPoints(rawData string) ([]float64, error) {
	var jsonArray []float64
	errJson := json.Unmarshal([]byte(rawData), &jsonArray)
	if errJson == nil {
		if len(jsonArray) == 0 {
			return nil, fmt.Errorf("Date source returned empty JSON array")
		}
		return jsonArray, nil
	}

	lines := strings.Split(strings.TrimSpace(rawData), "\n")
	points := make([]float64, 0, len(lines))

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		val, err := strconv.ParseFloat(trimmedLine, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%s' as float64", line)
		}
		points = append(points, val)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no data points found after parsing")
	}

	return points, nil
}

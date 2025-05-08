package components

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
	"github.com/rasjonell/dashbrew/internal/data"
)

type TableComponent struct {
	baseComponent
	table table.Model
	cols  []table.Column
}

func newTableComponent(base baseComponent) *TableComponent {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(5),
		table.WithStyles(getTableStyles(base.styles)),
	)

	return &TableComponent{
		baseComponent: base,
		table:         t,
		cols:          getTableColumns(base.config.Data.Columns, 0),
	}
}

func (c *TableComponent) View(w, h int, focused bool) string {
	style, focusedStyle, border := GetBorderStyle(c.styles.Border)
	borderStyle := style
	if focused {
		borderStyle = focusedStyle
		c.table.Focus()
	} else {
		c.table.Blur()
	}

	innerWidth, innerHeight := CalcWidthHeight(w, h)

	header := c.renderHeader(border)
	headerHeight := lipgloss.Height(header)

	// TODO: add a separate Resize() trigger for components
	c.cols = getTableColumns(c.config.Data.Columns, innerWidth-6)
	c.table.SetColumns(c.cols)

	tableHeight := max(1, innerHeight-headerHeight-1)
	c.table.SetHeight(tableHeight)

	var tableContent string
	if c.err != nil {
		errorMsg := fmt.Sprintf("[Error fetching/parsing data]\n%s", c.err.Error())
		tableContent = lipgloss.Place(innerWidth, tableHeight,
			lipgloss.Center, lipgloss.Center,
			WrapContent(errorMsg, innerWidth),
		)
	} else if len(c.table.Rows()) == 0 && c.err == nil {
		tableContent = lipgloss.Place(innerWidth, tableHeight,
			lipgloss.Center, lipgloss.Center,
			"[Loading or No Data]",
		)
	} else {
		tableContent = c.table.View()
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tableContent,
	)

	return borderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(fullContent)
}

func (c *TableComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	c.table, cmd = c.table.Update(msg)
	return c, cmd
}

func (c *TableComponent) SetContent(result data.FetchOutput) (Component, tea.Cmd) {
	newInstance := *c

	if result.Error() != nil {
		newInstance.err = result.Error()
		newInstance.table.SetRows([]table.Row{})
	} else {
		parsedRows, parseErr := c.parseDataToTableRows(result.Output(), c.config.Data.Columns)
		if parseErr != nil {
			newInstance.err = fmt.Errorf("Failed to parse table data: %w", parseErr)
			newInstance.table.SetRows([]table.Row{})
		} else {
			newInstance.err = nil
			newInstance.table.SetRows(parsedRows)
			newInstance.table.GotoTop()
		}
	}

	return &newInstance, nil
}

func (c *TableComponent) HandleAddMode(msg tea.KeyMsg) (Component, bool, tea.Cmd) {
	return c, true, nil
}

func (c *TableComponent) parseDataToTableRows(rawData string, columns []*config.ColumnConfig) ([]table.Row, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("Cannot parse table data without column definitions")
	}

	var parsedData any
	err := json.Unmarshal([]byte(rawData), &parsedData)
	if err != nil {
		// TODO: parse CSV as well
		return nil, fmt.Errorf("Failed to parse intput as JSON: %w", err)
	}

	dataArray, ok := parsedData.([]any)
	if !ok {
		return nil, fmt.Errorf("Unsupported JSON structure: expected top-level array")
	}

	if len(dataArray) == 0 {
		return []table.Row{}, nil
	}

	rows := make([]table.Row, 0, len(dataArray))
	firstElem := dataArray[0]

	switch firstElem.(type) {
	case []any: // Array of String Arrays [][]string
		for i, rowData := range dataArray {
			rowInterface, ok := rowData.([]any)
			if !ok {
				return nil, fmt.Errorf("Expected array of string arrays, but element at index %d is not an array", i)
			}

			row := make(table.Row, len(columns))
			for j := range columns {
				if j < len(columns) {
					row[j] = fmt.Sprintf("%v", rowInterface[j])
				} else {
					row[j] = ""
				}
			}
			rows = append(rows, row)
		}
		return rows, nil

	case map[string]any: // Array of JSON Objects
		for i, rowData := range dataArray {
			rowMap, ok := rowData.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("Expected array of objects, but element at index %d is not an object", i)
			}

			row := make(table.Row, len(columns))
			for j, colCfg := range columns {
				if colCfg.Field == "" {
					row[j] = ""
					continue
				}

				if cellData, keyExists := rowMap[colCfg.Field]; keyExists {
					row[j] = fmt.Sprintf("%v", cellData)
				} else {
					row[j] = ""
				}
			}
			rows = append(rows, row)
		}
		return rows, nil

	default: // Unknown structures
		return nil, fmt.Errorf("Unsupported JSON structure: expected array of string arrays or array of objects, got array of %T", firstElem)
	}
}

func getTableStyles(styleCfg *config.StyleConfig) table.Styles {
	s := table.DefaultStyles()
	normalFg, _, border := GetBorderStyle(styleCfg.Border)

	s.Header = s.Header.
		BorderStyle(border).
		BorderForeground(normalFg.GetBorderTopBackground()).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Bold(true)

	return s
}

func getTableColumns(columns []*config.ColumnConfig, w int) []table.Column {
	totalFlex := 0
	for _, cfgCol := range columns {
		totalFlex += GetFlex(cfgCol.Flex)
	}
	totalFlex = max(1, totalFlex)

	tableCols := make([]table.Column, len(columns))
	for i, cfgCol := range columns {
		width := w * GetFlex(cfgCol.Flex) / totalFlex
		tableCols[i] = table.Column{Title: cfgCol.Label, Width: width}
	}
	return tableCols
}

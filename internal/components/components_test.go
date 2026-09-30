package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/rasjonell/dashbrew/internal/config"
)

func TestRenderHeaderFitsAvailableWidth(t *testing.T) {
	for _, align := range []string{"left", "center", "right"} {
		for _, width := range []int{4, 5, 8, 12} {
			t.Run(align+"/"+string(rune('0'+width)), func(t *testing.T) {
				component := baseComponent{
					config: &config.Component{
						Title:      "Очень длинный заголовок панели",
						TitleAlign: align,
					},
				}
				_, _, border := GetBorderStyle(&config.BorderStyleConfig{Type: "normal"})

				header := component.renderHeader(border, false, width)
				if got := lipgloss.Width(header); got > width {
					t.Fatalf("header width = %d, want <= %d", got, width)
				}
			})
		}
	}
}

func TestTableColumnsFitAvailableWidth(t *testing.T) {
	const innerWidth = 40
	columns := []*config.ColumnConfig{
		{Flex: 1},
		{Flex: 1},
		{Flex: 1},
		{Flex: 1},
		{Flex: 1},
	}

	tableColumns := getTableColumns(columns, max(0, innerWidth-2*len(columns)))
	usedWidth := 0
	for _, column := range tableColumns {
		usedWidth += column.Width + 2
	}
	if usedWidth > innerWidth {
		t.Fatalf("table width = %d, want <= %d", usedWidth, innerWidth)
	}
}

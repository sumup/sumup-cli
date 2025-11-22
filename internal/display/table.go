package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const fallbackWidth = 120

// RenderTable prints rows in a table using the terminal width to wrap columns.
func RenderTable(title string, headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Printf("%s: No items to display\n", title)
		return
	}

	width, ok := terminalWidth()
	if !ok {
		width = fallbackWidth
	}

	basePadding := lipgloss.NewStyle().PaddingRight(2)
	headerStyle := basePadding.Foreground(lipgloss.Color("#8E8E8E")).Bold(true)
	idStyle := basePadding.Foreground(SumUpPink).Bold(true)
	defaultStyle := basePadding

	idColumns := make([]bool, len(headers))
	for i, header := range headers {
		if isIDHeader(header) {
			idColumns[i] = true
		}
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderRow(false).
		BorderColumn(false).
		BorderHeader(false).
		Headers(headers...).
		Rows(rows...).
		Width(width).
		Wrap(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col >= 0 && col < len(idColumns) && idColumns[col] {
				return idStyle
			}
			return defaultStyle
		})

	fmt.Println(title)
	fmt.Println(t.Render())
}

func isIDHeader(header string) bool {
	return strings.EqualFold(strings.TrimSpace(header), "id")
}

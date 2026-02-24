package display

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

const fallbackWidth = 120

// RenderTable prints rows in a table using the terminal width to wrap columns.
func RenderTable(title string, headers []string, rows [][]attribute.Value) {
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

	stringRows := make([][]string, len(rows))
	for i, row := range rows {
		stringRows[i] = make([]string, len(headers))
		for j := range headers {
			if j < len(row) {
				stringRows[i][j] = row[j].Text
				continue
			}
			stringRows[i][j] = "-"
		}
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderRow(false).
		BorderColumn(false).
		BorderHeader(false).
		Headers(headers...).
		Rows(stringRows...).
		Width(width).
		Wrap(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			style := defaultStyle
			if rowIndex := row - 1; rowIndex >= 0 && rowIndex < len(rows) && col >= 0 && col < len(rows[rowIndex]) {
				style = style.Inherit(rows[rowIndex][col].Style)
			}
			if col >= 0 && col < len(idColumns) && idColumns[col] {
				style = style.Inherit(idStyle)
			}
			return style
		})

	fmt.Println(title)
	fmt.Println(t.Render())
}

func isIDHeader(header string) bool {
	return strings.EqualFold(strings.TrimSpace(header), "id")
}

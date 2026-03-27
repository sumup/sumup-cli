package display

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

const fallbackWidth = 120

// TableOptions customizes human-readable table rendering.
type TableOptions struct {
	Title             string
	EmptyText         string
	IdentifierColumns []int
}

// RenderTable prints rows in a table using the terminal width to wrap columns.
func RenderTable(w io.Writer, title string, headers []string, rows [][]attribute.Value) error {
	return RenderTableWithOptions(w, headers, rows, TableOptions{
		Title:     title,
		EmptyText: "No items to display",
	})
}

// RenderTableWithOptions prints rows in a table using configurable presentation options.
func RenderTableWithOptions(w io.Writer, headers []string, rows [][]attribute.Value, opts TableOptions) error {
	if len(rows) == 0 {
		title := strings.TrimSpace(opts.Title)
		if title == "" {
			return writef(w, "%s\n", opts.EmptyText)
		}
		return writef(w, "%s: %s\n", title, opts.EmptyText)
	}

	width, ok := terminalWidth(w)
	if !ok {
		width = fallbackWidth
	}

	basePadding := lipgloss.NewStyle().PaddingRight(2)
	headerStyle := basePadding.Foreground(lipgloss.Color("#8E8E8E")).Bold(true)
	idStyle := basePadding.Foreground(SumUpPink).Bold(true)
	defaultStyle := basePadding

	identifierColumns := make(map[int]struct{}, len(opts.IdentifierColumns))
	for _, idx := range opts.IdentifierColumns {
		if idx >= 0 && idx < len(headers) {
			identifierColumns[idx] = struct{}{}
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
			if _, ok := identifierColumns[col]; ok {
				style = style.Inherit(idStyle)
			}
			return style
		})

	out := writerOrStdout(w)
	if strings.TrimSpace(opts.Title) != "" {
		if _, err := fmt.Fprintln(out, opts.Title); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, t.Render()); err != nil {
		return err
	}
	return nil
}

package display_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func TestPrintJSON(t *testing.T) {
	t.Run("renders pretty json", func(t *testing.T) {
		var out bytes.Buffer

		err := display.PrintJSON(&out, map[string]string{"status": "ok"})

		require.NoError(t, err)
		assert.Equal(t, normalizeOutput("{\n  \"status\": \"ok\"\n}\n"), normalizeOutput(out.String()))
	})
}

func TestDataList(t *testing.T) {
	t.Run("renders key value rows", func(t *testing.T) {
		var out bytes.Buffer

		err := display.DataList(&out, []attribute.KeyValue{
			attribute.Attribute("Status", attribute.Styled("ok")),
		})

		require.NoError(t, err)
		assert.Equal(t, normalizeOutput("Status: ok\n"), normalizeOutput(out.String()))
	})
}

func TestRenderTable(t *testing.T) {
	t.Run("renders title and row content", func(t *testing.T) {
		var out bytes.Buffer

		err := display.RenderTable(&out, "Items", []string{"ID"}, [][]attribute.Value{
			{attribute.ValueOf("123")},
		})

		require.NoError(t, err)
		assert.Equal(t, normalizeOutput("Items\nID\n123"), normalizeOutput(out.String()))
	})
}

func TestRenderTableWithOptionsSupportsEmptyTextWithoutTitle(t *testing.T) {
	t.Run("renders custom empty text without a title", func(t *testing.T) {
		var out bytes.Buffer

		err := display.RenderTableWithOptions(&out, []string{"ID"}, nil, display.TableOptions{
			EmptyText: "Nothing here",
		})

		require.NoError(t, err)
		assert.Equal(t, "Nothing here", strings.TrimSpace(out.String()))
	})
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func normalizeOutput(value string) string {
	value = ansiPattern.ReplaceAllString(value, "")
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.Join(normalized, "\n")
}

func TestRenderSections(t *testing.T) {
	t.Run("renders section titles and lines", func(t *testing.T) {
		var out bytes.Buffer

		err := display.RenderSections(&out, []display.Section{
			{
				Title: "Transaction",
				Pairs: []attribute.KeyValue{
					attribute.Attribute("Status", attribute.Styled("ok")),
				},
			},
			{
				Title: "Events",
				Lines: []string{"- created"},
			},
		})

		require.NoError(t, err)
		rendered := out.String()
		assert.Contains(t, rendered, "Transaction")
		assert.Contains(t, rendered, "Events")
		assert.Contains(t, rendered, "- created")
	})
}

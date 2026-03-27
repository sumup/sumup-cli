package display_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func TestPrintJSON(t *testing.T) {
	var out bytes.Buffer

	if err := display.PrintJSON(&out, map[string]string{"status": "ok"}); err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	const want = "{\n  \"status\": \"ok\"\n}\n"
	if normalizeOutput(out.String()) != normalizeOutput(want) {
		t.Fatalf("PrintJSON() output = %q, want %q", out.String(), want)
	}
}

func TestDataList(t *testing.T) {
	var out bytes.Buffer

	display.DataList(&out, []attribute.KeyValue{
		attribute.Attribute("Status", attribute.Styled("ok")),
	})

	const want = "Status: ok\n"
	if normalizeOutput(out.String()) != normalizeOutput(want) {
		t.Fatalf("DataList() output = %q, want %q", out.String(), want)
	}
}

func TestRenderTable(t *testing.T) {
	var out bytes.Buffer

	display.RenderTable(&out, "Items", []string{"ID"}, [][]attribute.Value{
		{attribute.ValueOf("123")},
	})

	const want = "Items\nID\n123"
	if normalizeOutput(out.String()) != normalizeOutput(want) {
		t.Fatalf("RenderTable() output = %q, want normalized %q", out.String(), want)
	}
}

func TestRenderTableWithOptionsSupportsEmptyTextWithoutTitle(t *testing.T) {
	var out bytes.Buffer

	display.RenderTableWithOptions(&out, []string{"ID"}, nil, display.TableOptions{
		EmptyText: "Nothing here",
	})

	rendered := out.String()
	if strings.TrimSpace(rendered) != "Nothing here" {
		t.Fatalf("RenderTableWithOptions() output = %q, want custom empty text", rendered)
	}
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
	var out bytes.Buffer

	display.RenderSections(&out, []display.Section{
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

	rendered := out.String()
	if !strings.Contains(rendered, "Transaction") || !strings.Contains(rendered, "Events") || !strings.Contains(rendered, "- created") {
		t.Fatalf("RenderSections() output = %q, want section titles and lines", rendered)
	}
}

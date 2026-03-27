package display_test

import (
	"bytes"
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

	if !strings.Contains(out.String(), `"status": "ok"`) {
		t.Fatalf("PrintJSON() output = %q, want JSON body", out.String())
	}
}

func TestDataList(t *testing.T) {
	var out bytes.Buffer

	display.DataList(&out, []attribute.KeyValue{
		attribute.Attribute("Status", attribute.Styled("ok")),
	})

	rendered := out.String()
	if !strings.Contains(rendered, "ok") || !strings.Contains(rendered, ":") {
		t.Fatalf("DataList() output = %q, want rendered pair", rendered)
	}
}

func TestRenderTable(t *testing.T) {
	var out bytes.Buffer

	display.RenderTable(&out, "Items", []string{"ID"}, [][]attribute.Value{
		{attribute.ValueOf("123")},
	})

	rendered := out.String()
	if !strings.Contains(rendered, "Items") || !strings.Contains(rendered, "123") {
		t.Fatalf("RenderTable() output = %q, want title and row content", rendered)
	}
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

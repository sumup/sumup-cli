package version

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestVersionCommandWritesToAppOutput(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCommand()
	cmd.Metadata = map[string]any{
		app.ContextKey: &app.Context{Output: &out},
	}

	if err := cmd.Action(context.Background(), cmd); err != nil {
		t.Fatalf("Action() error = %v", err)
	}

	if !strings.Contains(out.String(), "Version:") {
		t.Fatalf("Action() output = %q, want version details", out.String())
	}
}

func TestVersionCommandPrintsJSONWhenRequested(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCommand()
	cmd.Metadata = map[string]any{
		app.ContextKey: &app.Context{Output: &out, JSONOutput: true},
	}

	if err := cmd.Action(context.Background(), cmd); err != nil {
		t.Fatalf("Action() error = %v", err)
	}

	if !strings.Contains(out.String(), `"version"`) {
		t.Fatalf("Action() output = %q, want JSON payload", out.String())
	}
}

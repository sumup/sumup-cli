package version

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/buildinfo"
)

func TestVersionCommandPrintsLongBuildInfo(t *testing.T) {
	originalVersion := buildinfo.Version
	originalCommit := buildinfo.Commit
	originalDate := buildinfo.Date
	buildinfo.Version = "1.2.3"
	buildinfo.Commit = "abcdef123456"
	buildinfo.Date = "2026-03-27T00:00:00Z"
	t.Cleanup(func() {
		buildinfo.Version = originalVersion
		buildinfo.Commit = originalCommit
		buildinfo.Date = originalDate
	})

	var out bytes.Buffer

	cmd := NewCommand()
	cmd.Metadata = map[string]any{
		app.ContextKey: &app.Context{Output: &out},
	}
	if err := cmd.Action(context.Background(), cmd); err != nil {
		t.Fatalf("Action() error = %v", err)
	}

	rendered := strings.ReplaceAll(out.String(), "\r\n", "\n")
	const want = "Version: 1.2.3\nCommit: abcdef123456\nDate: 2026-03-27T00:00:00Z\n"
	if rendered != want {
		t.Fatalf("Action() output = %q, want %q", rendered, want)
	}
}

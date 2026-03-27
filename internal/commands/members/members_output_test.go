package members

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestRenderDeleteMemberResultPrintsJSONWhenRequested(t *testing.T) {
	var out bytes.Buffer

	appCtx := &app.Context{
		JSONOutput: true,
		Output:     &out,
	}

	if err := renderDeleteMemberResult(appCtx); err != nil {
		t.Fatalf("renderDeleteMemberResult() error = %v", err)
	}

	if !strings.Contains(out.String(), `"status": "deleted"`) {
		t.Fatalf("renderDeleteMemberResult() output = %q, want deleted acknowledgement", out.String())
	}
}

func TestRenderDeleteMemberResultWritesStatusMessageInHumanMode(t *testing.T) {
	var statusOut bytes.Buffer

	appCtx := &app.Context{
		StatusOutput: &statusOut,
	}

	if err := renderDeleteMemberResult(appCtx); err != nil {
		t.Fatalf("renderDeleteMemberResult() error = %v", err)
	}

	if !strings.Contains(statusOut.String(), "Member deleted") {
		t.Fatalf("renderDeleteMemberResult() status output = %q, want delete message", statusOut.String())
	}
}

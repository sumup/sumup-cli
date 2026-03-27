package transactions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestRenderRefundResultWritesStatusMessageInHumanMode(t *testing.T) {
	var statusOut bytes.Buffer

	appCtx := &app.Context{
		StatusOutput: &statusOut,
	}

	if err := renderRefundResult(appCtx); err != nil {
		t.Fatalf("renderRefundResult() error = %v", err)
	}

	if !strings.Contains(statusOut.String(), "Transaction refunded") {
		t.Fatalf("renderRefundResult() status output = %q, want refund message", statusOut.String())
	}
}

func TestRenderRefundResultPrintsJSONWhenRequested(t *testing.T) {
	var out bytes.Buffer

	appCtx := &app.Context{
		JSONOutput: true,
		Output:     &out,
	}

	if err := renderRefundResult(appCtx); err != nil {
		t.Fatalf("renderRefundResult() error = %v", err)
	}

	if !strings.Contains(out.String(), `"status": "refunded"`) {
		t.Fatalf("renderRefundResult() output = %q, want refunded JSON acknowledgement", out.String())
	}
}

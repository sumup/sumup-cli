package message_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumup/sumup-cli/internal/display/message"
)

func TestSuccessWritesToProvidedWriterWithoutANSIForBuffers(t *testing.T) {
	var out bytes.Buffer

	message.Success(&out, "created %s", "reader")

	rendered := out.String()
	if !strings.Contains(rendered, "created reader") {
		t.Fatalf("Success() output = %q, want rendered message", rendered)
	}
	if strings.Contains(rendered, "\033[") {
		t.Fatalf("Success() output = %q, want no ANSI escapes for non-terminal writers", rendered)
	}
}

package message_test

import (
	"bytes"
	"testing"

	"github.com/sumup/sumup-cli/internal/display/message"
)

func TestSuccessWritesToProvidedWriterWithoutANSIForBuffers(t *testing.T) {
	var out bytes.Buffer

	message.Success(&out, "created %s", "reader")

	const want = "✓ created reader\n"
	if out.String() != want {
		t.Fatalf("Success() output = %q, want %q", out.String(), want)
	}
}

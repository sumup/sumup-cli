package message_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/sumup/sumup-cli/internal/display/message"
)

func TestSuccessWritesToProvidedWriterWithoutANSIForBuffers(t *testing.T) {
	var out bytes.Buffer

	if err := message.Success(&out, "created %s", "reader"); err != nil {
		t.Fatalf("Success() error = %v", err)
	}

	const want = "✓ created reader\n"
	if out.String() != want {
		t.Fatalf("Success() output = %q, want %q", out.String(), want)
	}
}

func TestSuccessReturnsWriterErrors(t *testing.T) {
	expectedErr := errors.New("write failed")

	err := message.Success(failingWriter{err: expectedErr}, "created")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Success() error = %v, want %v", err, expectedErr)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

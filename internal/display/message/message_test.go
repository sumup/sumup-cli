package message_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/display/message"
)

func TestSuccessWritesToProvidedWriterWithoutANSIForBuffers(t *testing.T) {
	t.Run("writes plain output for non-terminal writers", func(t *testing.T) {
		var out bytes.Buffer

		err := message.Success(&out, "created %s", "reader")

		require.NoError(t, err)
		assert.Equal(t, "✓ created reader\n", out.String())
	})
}

func TestSuccessReturnsWriterErrors(t *testing.T) {
	t.Run("returns writer errors", func(t *testing.T) {
		expectedErr := errors.New("write failed")

		err := message.Success(failingWriter{err: expectedErr}, "created")

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

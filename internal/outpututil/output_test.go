package outpututil

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriterOrDefault(t *testing.T) {
	t.Run("returns fallback for nil writer", func(t *testing.T) {
		var fallback bytes.Buffer

		writer := WriterOrDefault(nil, &fallback)

		assert.Same(t, &fallback, writer)
	})

	t.Run("returns provided writer", func(t *testing.T) {
		var provided bytes.Buffer
		var fallback bytes.Buffer

		writer := WriterOrDefault(&provided, &fallback)

		assert.Same(t, &provided, writer)
	})
}

func TestFprintf(t *testing.T) {
	t.Run("writes using the provided writer", func(t *testing.T) {
		var out bytes.Buffer

		err := Fprintf(&out, nil, "hello %s", "world")

		require.NoError(t, err)
		assert.Equal(t, "hello world", out.String())
	})

	t.Run("returns writer errors", func(t *testing.T) {
		err := Fprintf(failingWriter{err: errWriteFailed}, nil, "ignored")

		require.Error(t, err)
		assert.ErrorIs(t, err, errWriteFailed)
	})
}

func TestFprintln(t *testing.T) {
	t.Run("writes a line using the provided writer", func(t *testing.T) {
		var out bytes.Buffer

		err := Fprintln(&out, nil, "hello")

		require.NoError(t, err)
		assert.Equal(t, "hello\n", out.String())
	})
}

func TestShouldUseColor(t *testing.T) {
	t.Run("disables color for non-terminal writers", func(t *testing.T) {
		assert.False(t, shouldUseColor(false, "xterm-256color", ""))
	})

	t.Run("disables color when no color is requested", func(t *testing.T) {
		assert.False(t, shouldUseColor(true, "xterm-256color", "1"))
	})

	t.Run("disables color for dumb terminals", func(t *testing.T) {
		assert.False(t, shouldUseColor(true, "dumb", ""))
	})

	t.Run("enables color for capable terminals", func(t *testing.T) {
		assert.True(t, shouldUseColor(true, "xterm-256color", ""))
	})
}

var errWriteFailed = errors.New("write failed")

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

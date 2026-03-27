package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldUseColor(t *testing.T) {
	t.Run("returns false for non-terminal writers", func(t *testing.T) {
		assert.False(t, shouldUseColor(false, "xterm-256color", ""))
	})

	t.Run("returns false when no color is requested", func(t *testing.T) {
		assert.False(t, shouldUseColor(true, "xterm-256color", "1"))
	})

	t.Run("returns false for dumb terminals", func(t *testing.T) {
		assert.False(t, shouldUseColor(true, "dumb", ""))
	})

	t.Run("returns true for interactive color-capable terminals", func(t *testing.T) {
		assert.True(t, shouldUseColor(true, "xterm-256color", ""))
	})
}

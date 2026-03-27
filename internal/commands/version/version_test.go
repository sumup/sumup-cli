package version_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/app"
	versioncmd "github.com/sumup/sumup-cli/internal/commands/version"
)

func TestNewCommand(t *testing.T) {
	t.Run("writes build details to app output", func(t *testing.T) {
		var out bytes.Buffer
		cmd := versioncmd.NewCommand()
		cmd.Metadata = map[string]any{
			app.ContextKey: &app.Context{Output: &out},
		}

		err := cmd.Action(context.Background(), cmd)

		require.NoError(t, err)
		assert.Contains(t, out.String(), "Version:")
	})

	t.Run("prints json when requested", func(t *testing.T) {
		var out bytes.Buffer
		cmd := versioncmd.NewCommand()
		cmd.Metadata = map[string]any{
			app.ContextKey: &app.Context{Output: &out, JSONOutput: true},
		}

		err := cmd.Action(context.Background(), cmd)

		require.NoError(t, err)
		assert.Contains(t, out.String(), `"version"`)
	})
}

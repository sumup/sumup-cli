package version_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/buildinfo"
	versioncmd "github.com/sumup/sumup-cli/internal/commands/version"
)

func TestNewCommandWithBuildMetadata(t *testing.T) {
	t.Run("prints long build info", func(t *testing.T) {
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

		cmd := versioncmd.NewCommand()
		cmd.Metadata = map[string]any{
			app.ContextKey: &app.Context{Output: &out},
		}

		err := cmd.Action(context.Background(), cmd)

		require.NoError(t, err)
		rendered := strings.ReplaceAll(out.String(), "\r\n", "\n")
		assert.Equal(t, "Version: 1.2.3\nCommit: abcdef123456\nDate: 2026-03-27T00:00:00Z\n", rendered)
	})
}

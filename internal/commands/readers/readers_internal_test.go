package readers

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestNewCommand(t *testing.T) {
	t.Run("reader context-aware commands do not require merchant code", func(t *testing.T) {
		cmd := NewCommand()

		for _, name := range []string{"get", "update", "terminate"} {
			subcommand := findSubcommand(t, cmd, name)
			flag := findStringFlag(t, subcommand, "merchant-code")
			assert.False(t, flag.Required, "%s merchant-code flag should allow context fallback", name)
		}
	})
}

func TestFormatCreateReaderError(t *testing.T) {
	t.Run("wraps sparse problem responses without panicking", func(t *testing.T) {
		err := formatCreateReaderError(&sumup.Problem{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "create reader:")
	})
}

func TestRenderReader(t *testing.T) {
	t.Run("uses exact local timestamps when requested", func(t *testing.T) {
		var out bytes.Buffer
		updatedAt := time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)

		reader := &sumup.Reader{
			ID:        "reader-1",
			Name:      "Front Desk",
			Status:    sumup.ReaderStatusPaired,
			UpdatedAt: updatedAt,
			Device: sumup.ReaderDevice{
				Identifier: "device-1",
			},
		}

		err := renderReader(&app.Context{Output: &out, ExactTimestamps: true}, &out, reader)

		require.NoError(t, err)
		assert.Contains(t, out.String(), updatedAt.In(time.Local).Format(time.RFC3339))
	})
}

func findSubcommand(t *testing.T, cmd *cli.Command, name string) *cli.Command {
	t.Helper()

	for _, subcommand := range cmd.Commands {
		if subcommand.Name == name {
			return subcommand
		}
	}

	require.Failf(t, "subcommand not found", "subcommand %q not found", name)
	return nil
}

func findStringFlag(t *testing.T, cmd *cli.Command, name string) *cli.StringFlag {
	t.Helper()

	for _, flag := range cmd.Flags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if ok && stringFlag.Name == name {
			return stringFlag
		}
	}

	require.Failf(t, "string flag not found", "string flag %q not found on %s", name, cmd.Name)
	return nil
}

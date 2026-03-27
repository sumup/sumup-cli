package readers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"
	"github.com/urfave/cli/v3"
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

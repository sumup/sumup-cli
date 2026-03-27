package readers

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestContextAwareReaderCommandsDoNotRequireMerchantCodeFlag(t *testing.T) {
	cmd := NewCommand()

	for _, name := range []string{"get", "update", "terminate"} {
		subcommand := findSubcommand(t, cmd, name)
		flag := findStringFlag(t, subcommand, "merchant-code")
		if flag.Required {
			t.Fatalf("%s merchant-code flag is required, want context fallback", name)
		}
	}
}

func findSubcommand(t *testing.T, cmd *cli.Command, name string) *cli.Command {
	t.Helper()

	for _, subcommand := range cmd.Commands {
		if subcommand.Name == name {
			return subcommand
		}
	}

	t.Fatalf("subcommand %q not found", name)
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

	t.Fatalf("string flag %q not found on %s", name, cmd.Name)
	return nil
}

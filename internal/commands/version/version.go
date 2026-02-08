package version

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/buildinfo"
)

// NewCommand returns the version command.
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print CLI build information",
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Println(buildinfo.Long())
			return nil
		},
	}
}

package version

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/buildinfo"
	"github.com/sumup/sumup-cli/internal/display"
)

// NewCommand returns the version command.
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print CLI build information",
		Action: func(_ context.Context, cmd *cli.Command) error {
			appCtx, err := app.GetAppContext(cmd)
			if err != nil {
				return err
			}

			if appCtx.JSONOutput {
				return display.PrintJSON(appCtx.Output, buildinfo.Current())
			}

			_, err = fmt.Fprintln(appCtx.Output, buildinfo.Long())
			return nil
		},
	}
}

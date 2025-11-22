package main

import (
	"context"
	"os"

	sumupclient "github.com/sumup/sumup-go/client"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func main() {
	cliApp := &cli.Command{
		Name:                  "sumup",
		Usage:                 "Command line tool for the SumUp API",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "API key used for authorization. Falls back to SUMUP_API_KEY.",
				Sources: cli.EnvVars("SUMUP_API_KEY"),
			},
			&cli.StringFlag{
				Name:    "base-url",
				Usage:   "Base URL for SumUp API calls.",
				Value:   sumupclient.APIUrl,
				Sources: cli.EnvVars("SUMUP_BASE_URL"),
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output results in JSON format instead of human-readable tables.",
			},
			&cli.BoolFlag{
				Name:  "exact-timestamps",
				Usage: "Show timestamp fields using the exact local time instead of relative strings.",
			},
		},
		Metadata: map[string]any{},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			appCtx, err := app.NewContext(
				cmd.String("api-key"),
				cmd.String("base-url"),
				cmd.Bool("json"),
				cmd.Bool("exact-timestamps"),
			)
			if err != nil {
				return ctx, err
			}
			cmd.Root().Metadata[app.ContextKey] = appCtx
			return ctx, nil
		},
		Commands: commands.All(),
	}

	if err := cliApp.Run(context.Background(), os.Args); err != nil {
		message.Error("%v", err)
		os.Exit(1)
	}
}

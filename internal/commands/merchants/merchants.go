package merchants

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "merchants",
		Usage: "Placeholder for the merchants API resource.",
		Commands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Get merchant information.",
				Action: getMerchant,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code to retrieve information for. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			},
		},
	}
}

func getMerchant(_ context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	message.Notify(appCtx.Output, "Getting merchant information for: %s", merchantCode)
	message.Warn(appCtx.Output, "Merchants functionality not yet fully implemented.")
	return nil
}

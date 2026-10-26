package tokens

import (
	"context"
	"fmt"
	"slices"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "tokens",
		Usage: "Create short-lived merchant tokens for SumUp SDKs.",
		Commands: []*cli.Command{
			apicommands.Bind("CreateMerchantToken", &cli.Command{
				Name:   "create",
				Usage:  "Create a merchant token for the requested actions. Requires --json to output the access token.",
				Action: createToken,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code to create a token for. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringSliceFlag{
						Name:     "action",
						Usage:    "Action to allow (checkout). Repeat for multiple actions.",
						Required: true,
					},
				},
			}),
		},
	}
}

func createToken(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	if !appCtx.JSONOutput {
		return fmt.Errorf("creating a merchant token requires --json to output the access token")
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	body := sumup.TokensCreateParams{}
	for _, action := range cmd.StringSlice("action") {
		if action != string(sumup.TokensCreateParamsActionCheckout) {
			return fmt.Errorf("unsupported token action %q: supported values: checkout", action)
		}
		value := sumup.TokensCreateParamsAction(action)
		if !slices.Contains(body.Actions, value) {
			body.Actions = append(body.Actions, value)
		}
	}
	if len(body.Actions) == 0 {
		return fmt.Errorf("provide at least one --action")
	}
	token, err := appCtx.Client.Tokens.Create(ctx, merchantCode, body)
	if err != nil {
		return fmt.Errorf("create merchant token: %w", err)
	}
	return display.PrintJSON(appCtx.Output, token)
}

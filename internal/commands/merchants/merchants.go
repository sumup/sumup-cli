package merchants

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "merchants",
		Usage: "Commands related to merchant accounts.",
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

func getMerchant(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	merchant, err := appCtx.Client.Merchants.Get(ctx, merchantCode, sumup.MerchantsGetParams{})
	if err != nil {
		return fmt.Errorf("get merchant: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, merchant)
	}

	return renderMerchant(appCtx, merchant)
}

func renderMerchant(appCtx *app.Context, merchant *sumup.Merchant) error {
	if merchant == nil {
		return nil
	}

	details := []attribute.KeyValue{
		attribute.Attribute("Merchant Code", attribute.Styled(merchant.MerchantCode)),
		attribute.OptionalString("Alias", merchant.Alias),
		attribute.Attribute("Country", attribute.Styled(merchant.Country)),
		attribute.Attribute("Default Currency", attribute.Styled(merchant.DefaultCurrency)),
		attribute.Attribute("Default Locale", attribute.Styled(merchant.DefaultLocale)),
		attribute.Attribute("Sandbox", attribute.Styled(util.BoolLabel(merchant.Sandbox))),
		attribute.OptionalString("Organization ID", merchant.OrganizationID),
		attribute.Attribute("Created At", attribute.Styled(util.TimeOrDash(appCtx, &merchant.CreatedAt))),
		attribute.Attribute("Updated At", attribute.Styled(util.TimeOrDash(appCtx, &merchant.UpdatedAt))),
	}

	if merchant.BusinessProfile != nil {
		details = append(details,
			attribute.OptionalString("Business Name", merchant.BusinessProfile.Name),
			attribute.OptionalString("Business Email", merchant.BusinessProfile.Email),
			attribute.Optional("Business Phone", merchant.BusinessProfile.PhoneNumber),
			attribute.OptionalString("Business Website", merchant.BusinessProfile.Website),
		)
	}

	return display.DataList(appCtx.Output, details)
}

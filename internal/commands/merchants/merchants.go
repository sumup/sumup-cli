package merchants

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/nullable"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "merchants",
		Usage: "Commands for retrieving merchant information.",
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
					&cli.StringFlag{
						Name:  "version",
						Usage: "Optional resource version to read. Supported value: latest.",
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

	params := sumup.MerchantsGetParams{}
	if version := cmd.String("version"); version != "" {
		params.Version = &version
	}

	merchant, err := appCtx.Client.Merchants.Get(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("get merchant: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(merchant)
	}

	renderMerchant(merchant)
	return nil
}

func renderMerchant(merchant *sumup.Merchant) {
	if merchant == nil {
		return
	}

	details := []attribute.KeyValue{
		attribute.Attribute("Merchant Code", attribute.Styled(merchant.MerchantCode)),
		attribute.Attribute("Country", attribute.Styled(merchant.Country)),
		attribute.Attribute("Default Currency", attribute.Styled(merchant.DefaultCurrency)),
		attribute.Attribute("Default Locale", attribute.Styled(merchant.DefaultLocale)),
		attribute.Attribute("Sandbox", attribute.Styled(boolLabel(merchant.Sandbox))),
		attribute.Attribute("Created At", attribute.Styled(merchant.CreatedAt.UTC().Format(time.RFC3339))),
		attribute.Attribute("Updated At", attribute.Styled(merchant.UpdatedAt.UTC().Format(time.RFC3339))),
	}

	if merchant.Alias != nil && *merchant.Alias != "" {
		details = append(details, attribute.OptionalString("Alias", merchant.Alias))
	}
	if merchant.BusinessType != nil && *merchant.BusinessType != "" {
		details = append(details, attribute.OptionalString("Business Type", merchant.BusinessType))
	}
	if merchant.OrganizationID != nil && *merchant.OrganizationID != "" {
		details = append(details, attribute.OptionalString("Organization ID", merchant.OrganizationID))
	}

	if business := merchant.BusinessProfile; business != nil {
		details = append(details, attribute.OptionalString("Business Name", business.Name))
		details = append(details, attribute.OptionalString("Business Email", business.Email))
		details = append(details, attribute.Optional("Business Phone", business.PhoneNumber))
		details = append(details, attribute.OptionalString("Business Website", business.Website))
		if address := formatAddress(business.Address); address != "" {
			details = append(details, attribute.Attribute("Business Address", attribute.Styled(address)))
		}
	}

	if company := merchant.Company; company != nil {
		details = append(details, attribute.OptionalString("Legal Name", company.Name))
		details = append(details, attribute.Optional("Legal Phone", company.PhoneNumber))
		if company.LegalType != nil && *company.LegalType != "" {
			details = append(details, attribute.Attribute("Legal Type", attribute.Styled(*company.LegalType)))
		}
		if website := nullableString(company.Website); website != "" {
			details = append(details, attribute.Attribute("Legal Website", attribute.Styled(website)))
		}
		if address := formatAddress(company.Address); address != "" {
			details = append(details, attribute.Attribute("Company Address", attribute.Styled(address)))
		}
		if address := formatAddress(company.TradingAddress); address != "" {
			details = append(details, attribute.Attribute("Trading Address", attribute.Styled(address)))
		}
	}

	display.DataList(details)
}

func formatAddress(address *sumup.Address) string {
	if address == nil {
		return ""
	}

	parts := make([]string, 0, 8)
	for _, line := range address.StreetAddress {
		if line != "" {
			parts = append(parts, line)
		}
	}

	appendIfPresent := func(value *string) {
		if value != nil && *value != "" {
			parts = append(parts, *value)
		}
	}

	appendIfPresent(address.City)
	appendIfPresent(address.PostTown)
	appendIfPresent(address.PostCode)
	appendIfPresent(address.ZipCode)
	appendIfPresent(address.State)
	appendIfPresent(address.Region)
	if address.Country != "" {
		parts = append(parts, string(address.Country))
	}

	return strings.Join(parts, ", ")
}

func boolLabel(value *bool) string {
	if value == nil {
		return "-"
	}
	if *value {
		return "Yes"
	}
	return "No"
}

func nullableString(value *nullable.Field[string]) string {
	if value == nil {
		return ""
	}
	current := value.Value()
	if current == nil {
		return ""
	}
	return *current
}

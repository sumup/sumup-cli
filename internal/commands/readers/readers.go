package readers

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/readers"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/currency"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "readers",
		Usage: "Commands for managing in-person readers.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List paired readers for a merchant.",
				Action: listReaders,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "merchant-code",
						Usage:    "Merchant code whose readers should be listed.",
						Sources:  cli.EnvVars("SUMUP_MERCHANT_CODE"),
						Required: true,
					},
				},
			},
			{
				Name:   "add",
				Usage:  "Pair a new reader with the merchant account.",
				Action: addReader,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "merchant-code",
						Usage:    "Merchant code that will own the new reader.",
						Sources:  cli.EnvVars("SUMUP_MERCHANT_CODE"),
						Required: true,
					},
					&cli.StringFlag{
						Name:     "pairing-code",
						Usage:    "Pairing code shown on the physical device.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Friendly name to help identify the reader.",
						Required: true,
					},
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a paired reader from the merchant account.",
				Action:    deleteReader,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "merchant-code",
						Usage:    "Merchant code that owns the reader.",
						Sources:  cli.EnvVars("SUMUP_MERCHANT_CODE"),
						Required: true,
					},
				},
			},
			{
				Name:      "checkout",
				Usage:     "Trigger a checkout on a specific reader device.",
				Action:    readerCheckout,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "merchant-code",
						Usage:    "Merchant code that owns the reader.",
						Sources:  cli.EnvVars("SUMUP_MERCHANT_CODE"),
						Required: true,
					},
					&cli.StringFlag{
						Name:     "amount",
						Usage:    "Amount to charge, expressed in major units (for example 14.99).",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "minor-unit",
						Usage: "Number of decimal places for the currency (for example 2 for EUR).",
						Value: 2,
					},
					&cli.StringFlag{
						Name:     "currency",
						Usage:    fmt.Sprintf("Currency used for the transaction amount. Supported: %s", strings.Join(currency.Supported(), ", ")),
						Required: true,
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Optional description shown in dashboards.",
					},
					&cli.StringFlag{
						Name:  "return-url",
						Usage: "URL that receives the payment result.",
					},
					&cli.StringFlag{
						Name:  "card-type",
						Usage: "Optional card type hint (required for some countries).",
					},
					&cli.IntFlag{
						Name:  "installments",
						Usage: "Number of installments (supported in select regions).",
					},
					&cli.IntFlag{
						Name:  "tip-timeout",
						Usage: "Seconds allowed for the cardholder to pick a tip rate.",
					},
					&cli.Float64SliceFlag{
						Name:  "tip-rate",
						Usage: "Provide multiple --tip-rate values to configure suggested tips (percentages 0.01-0.99).",
					},
					&cli.StringFlag{
						Name:  "affiliate-app-id",
						Usage: "Affiliate app ID to attribute the transaction.",
					},
					&cli.StringFlag{
						Name:  "affiliate-key",
						Usage: "Affiliate key to attribute the transaction.",
					},
					&cli.StringFlag{
						Name:  "affiliate-foreign-transaction-id",
						Usage: "Affiliate foreign transaction ID to attribute the transaction.",
					},
				},
			},
		},
	}
}

func listReaders(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	response, err := appCtx.Client.Readers.List(ctx, cmd.String("merchant-code"))
	if err != nil {
		return fmt.Errorf("list readers: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(response.Items)
	}

	rows := make([][]string, 0, len(response.Items))
	for _, reader := range response.Items {
		name := string(reader.Name)
		model := string(reader.Device.Model)
		rows = append(rows, []string{
			string(reader.ID),
			name,
			string(reader.Status),
			model,
			reader.Device.Identifier,
		})
	}

	display.RenderTable("Readers", []string{"ID", "Name", "Status", "Model", "Identifier"}, rows)
	return nil
}

func addReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	body := readers.CreateReaderBody{
		PairingCode: readers.ReaderPairingCode(cmd.String("pairing-code")),
		Name:        readers.ReaderName(cmd.String("name")),
	}

	reader, err := appCtx.Client.Readers.Create(ctx, cmd.String("merchant-code"), body)
	if err != nil {
		return fmt.Errorf("create reader: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(reader)
	}

	message.Success("Reader created")
	display.DataList([]attribute.KeyValue{
		attribute.ID(string(reader.ID)),
		attribute.Attribute("Name", attribute.Styled(string(reader.Name))),
		attribute.Attribute("Status", attribute.Styled(string(reader.Status))),
		attribute.Attribute("Model", attribute.Styled(string(reader.Device.Model))),
		attribute.Attribute("Identifier", attribute.Styled(reader.Device.Identifier)),
	})
	return nil
}

func deleteReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	err = appCtx.Client.Readers.Delete(ctx, cmd.String("merchant-code"), readers.ReaderId(readerID))
	if err != nil {
		return fmt.Errorf("delete reader: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(map[string]string{"status": "deleted"})
	}

	message.Success("Reader deleted")
	return nil
}

func readerCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}
	parsedCurrency, err := currency.Parse(cmd.String("currency"))
	if err != nil {
		return err
	}
	value, err := currency.ToMinorUnits(cmd.String("amount"), int32(cmd.Int("minor-unit")))
	if err != nil {
		return err
	}
	if value > int64(math.MaxInt32) || value < int64(math.MinInt32) {
		return fmt.Errorf("amount is too large to convert into minor units")
	}

	body := readers.CreateReaderCheckoutBody{
		TotalAmount: readers.CreateReaderCheckoutBodyTotalAmount{
			Currency:  currency.Code(parsedCurrency),
			MinorUnit: cmd.Int("minor-unit"),
			Value:     int(value),
		},
	}

	if desc := cmd.String("description"); desc != "" {
		body.Description = &desc
	}
	if returnURL := cmd.String("return-url"); returnURL != "" {
		body.ReturnUrl = &returnURL
	}
	if cardType := cmd.String("card-type"); cardType != "" {
		ct := readers.CreateReaderCheckoutBodyCardType(cardType)
		body.CardType = &ct
	}
	if cmd.IsSet("installments") {
		value := cmd.Int("installments")
		body.Installments = &value
	}
	if cmd.IsSet("tip-timeout") {
		value := cmd.Int("tip-timeout")
		body.TipTimeout = &value
	}
	if rates := cmd.Float64Slice("tip-rate"); len(rates) > 0 {
		body.TipRates = make([]float32, 0, len(rates))
		for _, rate := range rates {
			body.TipRates = append(body.TipRates, float32(rate))
		}
	}

	affiliate, err := buildAffiliatePayload(cmd)
	if err != nil {
		return err
	}
	if affiliate != nil {
		body.Affiliate = affiliate
	}

	response, err := appCtx.Client.Readers.CreateCheckout(ctx, cmd.String("merchant-code"), readerID, body)
	if err != nil {
		return fmt.Errorf("trigger reader checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(response)
	}

	message.Success("Checkout initiated")
	majorAmount := float64(value) / math.Pow10(cmd.Int("minor-unit"))
	details := make([]attribute.KeyValue, 0, 2)
	details = append(details, attribute.Attribute("Amount", attribute.Styled(currency.Format(majorAmount, parsedCurrency))))
	if desc := cmd.String("description"); desc != "" {
		details = append(details, attribute.Attribute("Description", attribute.Styled(desc)))
	}
	display.DataList(details)

	return nil
}

func buildAffiliatePayload(cmd *cli.Command) (*readers.CreateReaderCheckoutBodyAffiliate, error) {
	appID := cmd.String("affiliate-app-id")
	key := cmd.String("affiliate-key")
	foreignID := cmd.String("affiliate-foreign-transaction-id")
	if appID == "" && key == "" && foreignID == "" {
		return nil, nil
	}
	if appID == "" || key == "" || foreignID == "" {
		return nil, fmt.Errorf("affiliate requires --affiliate-app-id, --affiliate-key, and --affiliate-foreign-transaction-id")
	}
	return &readers.CreateReaderCheckoutBodyAffiliate{
		AppId:                appID,
		Key:                  key,
		ForeignTransactionId: foreignID,
	}, nil
}

package readers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/nullable"

	"github.com/sumup/sumup-cli/internal/apicommands"
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
			apicommands.Bind("ListReaders", &cli.Command{
				Name:   "list",
				Usage:  "List paired readers for a merchant.",
				Action: listReaders,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose readers should be listed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			}),
			apicommands.Bind("CreateReader", &cli.Command{
				Name:   "add",
				Usage:  "Pair a new reader with the merchant account.",
				Action: addReader,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that will own the new reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
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
			}),
			apicommands.Bind("DeleteReader", &cli.Command{
				Name:      "delete",
				Usage:     "Delete a paired reader from the merchant account.",
				Action:    deleteReader,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			}),
			apicommands.Bind("GetReaderStatus", &cli.Command{
				Name:      "status",
				Usage:     "Show the last known status of a reader.",
				Action:    readerStatus,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			}),
			apicommands.Bind("GetReader", &cli.Command{
				Name:      "get",
				Usage:     "Get a paired reader.",
				Action:    getReader,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:  "if-modified-since",
						Usage: "Optional If-Modified-Since query value.",
					},
				},
			}),
			apicommands.Bind("UpdateReader", &cli.Command{
				Name:      "update",
				Usage:     "Update a paired reader.",
				Action:    updateReader,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Updated reader name.",
						Required: true,
					},
				},
			}),
			apicommands.Bind("CreateReaderTerminate", &cli.Command{
				Name:      "terminate",
				Usage:     "Terminate the current reader checkout.",
				Action:    terminateCheckout,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			}),
			apicommands.Bind("CreateReaderCheckout", &cli.Command{
				Name:      "checkout",
				Usage:     "Trigger a checkout on a specific reader device.",
				Action:    readerCheckout,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
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
			}),
			apicommands.Bind("CreateGoReaderCheckout", &cli.Command{
				Name:      "go-checkout",
				Usage:     "Trigger a checkout on a SumUp Go reader.",
				Action:    goReaderCheckout,
				ArgsUsage: "<reader-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
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
						Name:     "client-transaction-id",
						Usage:    "Caller-supplied correlation identifier, used as the idempotency key.",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "tip-amount",
						Usage: "Optional tip amount in major units, added on top of the total amount.",
					},
					&cli.StringFlag{
						Name:  "affiliate-app-id",
						Usage: "Affiliate app ID to attribute the transaction.",
					},
					&cli.StringFlag{
						Name:  "affiliate-key",
						Usage: "Affiliate key to attribute the transaction.",
					},
				},
			}),
			apicommands.Bind("GetReaderCheckout", &cli.Command{
				Name:      "get-checkout",
				Usage:     "Get a checkout for a reader.",
				Action:    getReaderCheckout,
				ArgsUsage: "<reader-id> <checkout-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the reader. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			}),
		},
	}
}

func listReaders(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	response, err := appCtx.Client.Readers.List(ctx, merchantCode)
	if err != nil {
		return fmt.Errorf("list readers: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response.Items)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for _, reader := range response.Items {
		name := string(reader.Name)
		model := string(reader.Device.Model)
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(string(reader.ID)),
			attribute.ValueOf(name),
			attribute.ValueOf(string(reader.Status)),
			attribute.ValueOf(model),
			attribute.ValueOf(reader.Device.Identifier),
		})
	}

	return display.RenderTableWithOptions(appCtx.Output, []string{"ID", "Name", "Status", "Model", "Identifier"}, rows, display.TableOptions{
		Title:             "Readers",
		EmptyText:         "No items to display",
		IdentifierColumns: []int{0},
	})
}

func addReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	body := sumup.ReadersCreateParams{
		PairingCode: sumup.ReaderPairingCode(cmd.String("pairing-code")),
		Name:        sumup.ReaderName(cmd.String("name")),
	}

	reader, err := appCtx.Client.Readers.Create(ctx, merchantCode, body)
	if err != nil {
		return formatCreateReaderError(err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      reader,
		SuccessMessage: "Reader created",
		Details: []attribute.KeyValue{
			attribute.ID(string(reader.ID)),
			attribute.Attribute("Name", attribute.Styled(string(reader.Name))),
			attribute.Attribute("Status", attribute.Styled(string(reader.Status))),
			attribute.Attribute("Model", attribute.Styled(string(reader.Device.Model))),
			attribute.Attribute("Identifier", attribute.Styled(reader.Device.Identifier)),
		},
	})
}

func formatCreateReaderError(err error) error {
	if pErr := new(sumup.Problem); errors.As(err, &pErr) {
		return fmt.Errorf("create reader: %w", pErr)
	}
	return fmt.Errorf("create reader: %w", err)
}

func deleteReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	err = appCtx.Client.Readers.Delete(ctx, merchantCode, sumup.ReaderID(readerID))
	if err != nil {
		return fmt.Errorf("delete reader: %w", err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      map[string]string{"status": "deleted"},
		SuccessMessage: "Reader deleted",
	})
}

func readerCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}
	parsedCurrency, value, err := parseAmountMinorUnits(cmd, cmd.String("amount"))
	if err != nil {
		return err
	}

	body := sumup.ReadersCreateCheckoutParams{
		TotalAmount: sumup.CreateCheckoutRequestTotalAmount{
			Currency:  currency.Code(parsedCurrency),
			MinorUnit: cmd.Int("minor-unit"),
			Value:     value,
		},
	}

	if desc := cmd.String("description"); desc != "" {
		body.Description = &desc
	}
	if returnURL := cmd.String("return-url"); returnURL != "" {
		body.ReturnURL = &returnURL
	}
	if cardType := cmd.String("card-type"); cardType != "" {
		ct := sumup.CreateCheckoutRequestCardType(cardType)
		body.CardType = &ct
	}
	if cmd.IsSet("installments") {
		value := cmd.Int("installments")
		body.Installments = nullable.Int(value)
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
		body.Affiliate = nullable.Value(*affiliate)
	}

	response, err := appCtx.Client.Readers.CreateCheckout(ctx, merchantCode, readerID, body)
	if err != nil {
		return fmt.Errorf("trigger reader checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	if err := message.Success(appCtx.StatusOutput, "Checkout initiated"); err != nil {
		return err
	}
	majorAmount := float64(value) / math.Pow10(cmd.Int("minor-unit"))
	details := make([]attribute.KeyValue, 0, 2)
	details = append(details, attribute.Attribute("Amount", attribute.Styled(currency.Format(majorAmount, parsedCurrency))))
	if desc := cmd.String("description"); desc != "" {
		details = append(details, attribute.Attribute("Description", attribute.Styled(desc)))
	}
	return display.DataList(appCtx.Output, details)
}

func goReaderCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}
	parsedCurrency, value, err := parseAmountMinorUnits(cmd, cmd.String("amount"))
	if err != nil {
		return err
	}

	body := sumup.ReadersCreateGoCheckoutParams{
		ClientTransactionID: cmd.String("client-transaction-id"),
		TotalAmount: sumup.Amount{
			Currency: currency.Code(parsedCurrency),
			Value:    value,
		},
	}
	if cmd.IsSet("tip-amount") {
		_, tipValue, err := parseAmountMinorUnits(cmd, cmd.String("tip-amount"))
		if err != nil {
			return fmt.Errorf("tip amount: %w", err)
		}
		body.TipAmount = &tipValue
	}

	affiliate, err := buildGoAffiliatePayload(cmd)
	if err != nil {
		return err
	}
	if affiliate != nil {
		body.Affiliate = affiliate
	}

	response, err := appCtx.Client.Readers.CreateGoCheckout(ctx, merchantCode, sumup.ReaderID(readerID), body)
	if err != nil {
		return fmt.Errorf("trigger go reader checkout: %w", err)
	}

	details := []attribute.KeyValue{
		attribute.Attribute("Amount", attribute.Styled(currency.Format(float64(value)/math.Pow10(cmd.Int("minor-unit")), parsedCurrency))),
		attribute.Attribute("Client Transaction ID", attribute.Styled(cmd.String("client-transaction-id"))),
	}
	if response != nil && response.Data != nil {
		if response.Data.TransactionCode != nil && *response.Data.TransactionCode != "" {
			details = append(details, attribute.Attribute("Transaction Code", attribute.Styled(*response.Data.TransactionCode)))
		}
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      response,
		SuccessMessage: "Go reader checkout initiated",
		Details:        details,
	})
}

func getReaderCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	if cmd.Args().Len() != 2 {
		return fmt.Errorf("expected exactly 2 arguments: reader ID and checkout ID")
	}
	readerID := cmd.Args().Get(0)
	checkoutID := cmd.Args().Get(1)

	response, err := appCtx.Client.Readers.GetCheckout(ctx, merchantCode, readerID, checkoutID)
	if err != nil {
		return fmt.Errorf("get reader checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}
	if response == nil {
		return nil
	}

	return renderReaderCheckout(appCtx, appCtx.Output, &response.Data)
}

func readerStatus(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	response, err := appCtx.Client.Readers.GetStatus(ctx, merchantCode, readerID)
	if err != nil {
		return fmt.Errorf("get reader status: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	data := response.Data
	batteryTemp := attribute.OptionalValue(data.BatteryTemperature)
	if data.BatteryTemperature != nil {
		batteryTemp = attribute.ValueOf(fmt.Sprintf("%d°C", *data.BatteryTemperature))
	}

	details := []attribute.KeyValue{
		attribute.ID(readerID),
		attribute.Attribute("Status", attribute.Styled(string(data.Status))),
		attribute.Optional("State", data.State),
		attribute.Optional("Connection", data.ConnectionType),
		attribute.Attribute("Battery Level", attribute.Styled(readerStatusBatteryLevel(data.BatteryLevel))),
		attribute.Attribute("Battery Temp", batteryTemp),
		attribute.OptionalString("Firmware", data.FirmwareVersion),
		attribute.Attribute("Last Activity", attribute.Styled(util.TimeOrDash(appCtx, data.LastActivity))),
	}

	return display.DataList(appCtx.Output, details)
}

func getReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	params := sumup.ReadersGetParams{}
	if value := cmd.String("if-modified-since"); value != "" {
		params.IfModifiedSince = &value
	}

	reader, err := appCtx.Client.Readers.Get(ctx, merchantCode, sumup.ReaderID(readerID), params)
	if err != nil {
		return fmt.Errorf("get reader: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, reader)
	}

	return renderReader(appCtx, appCtx.Output, reader)
}

func updateReader(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	name := sumup.ReaderName(cmd.String("name"))
	body := sumup.ReadersUpdateParams{Name: &name}
	reader, err := appCtx.Client.Readers.Update(ctx, merchantCode, sumup.ReaderID(readerID), body)
	if err != nil {
		return fmt.Errorf("update reader: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, reader)
	}

	if err := message.Success(appCtx.StatusOutput, "Reader updated"); err != nil {
		return err
	}
	return renderReader(appCtx, appCtx.Output, reader)
}

func terminateCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}

	if err := appCtx.Client.Readers.TerminateCheckout(ctx, merchantCode, readerID); err != nil {
		return fmt.Errorf("terminate reader checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, map[string]string{"status": "termination_requested"})
	}

	return message.Success(appCtx.StatusOutput, "Reader checkout termination requested")
}

func renderReader(appCtx *app.Context, w io.Writer, reader *sumup.Reader) error {
	if reader == nil {
		return nil
	}

	return display.NewDetailsBuilder().
		AddID(string(reader.ID)).
		Add("Name", attribute.Styled(string(reader.Name))).
		Add("Status", attribute.Styled(string(reader.Status))).
		Add("Model", attribute.Styled(string(reader.Device.Model))).
		Add("Identifier", attribute.Styled(reader.Device.Identifier)).
		Add("Updated At", attribute.Styled(util.TimeOrDash(appCtx, &reader.UpdatedAt))).
		AddOptionalString("Service Account ID", reader.ServiceAccountID).
		Render(w)
}

func readerStatusBatteryLevel(v *float32) string {
	if v == nil {
		return "-"
	}

	return fmt.Sprintf("%.0f%%", *v)
}

func buildAffiliatePayload(cmd *cli.Command) (*sumup.CreateCheckoutRequestAffiliate, error) {
	appID := cmd.String("affiliate-app-id")
	key := cmd.String("affiliate-key")
	foreignID := cmd.String("affiliate-foreign-transaction-id")
	if appID == "" && key == "" && foreignID == "" {
		return nil, nil
	}
	if appID == "" || key == "" || foreignID == "" {
		return nil, fmt.Errorf("affiliate requires --affiliate-app-id, --affiliate-key, and --affiliate-foreign-transaction-id")
	}
	return &sumup.CreateCheckoutRequestAffiliate{
		AppID:                appID,
		Key:                  key,
		ForeignTransactionID: foreignID,
	}, nil
}

func buildGoAffiliatePayload(cmd *cli.Command) (*sumup.Affiliate, error) {
	appID := cmd.String("affiliate-app-id")
	key := cmd.String("affiliate-key")
	if appID == "" && key == "" {
		return nil, nil
	}
	if appID == "" || key == "" {
		return nil, fmt.Errorf("affiliate requires --affiliate-app-id and --affiliate-key")
	}
	return &sumup.Affiliate{
		AppID: appID,
		Key:   key,
	}, nil
}

func parseAmountMinorUnits(cmd *cli.Command, amount string) (sumup.Currency, int, error) {
	parsedCurrency, err := currency.Parse(cmd.String("currency"))
	if err != nil {
		return "", 0, err
	}
	value, err := currency.ToMinorUnits(amount, int32(cmd.Int("minor-unit")))
	if err != nil {
		return "", 0, err
	}
	if value > int64(math.MaxInt32) || value < int64(math.MinInt32) {
		return "", 0, fmt.Errorf("amount is too large to convert into minor units")
	}
	return parsedCurrency, int(value), nil
}

func renderReaderCheckout(appCtx *app.Context, w io.Writer, data *sumup.GetReaderCheckoutResponseData) error {
	if data == nil {
		return nil
	}

	majorAmount := float64(data.TotalAmount.Value) / math.Pow10(data.TotalAmount.MinorUnit)
	amountText := fmt.Sprintf("%.*f %s", data.TotalAmount.MinorUnit, majorAmount, data.TotalAmount.Currency)
	if parsedCurrency, err := currency.Parse(data.TotalAmount.Currency); err == nil {
		amountText = currency.Format(majorAmount, parsedCurrency)
	}

	details := display.NewDetailsBuilder().
		AddID(data.CheckoutID).
		Add("Status", attribute.Styled(string(data.Status))).
		Add("Payment Type", attribute.Styled(string(data.PaymentType))).
		Add("Amount", attribute.Styled(amountText)).
		Add("Client Transaction ID", attribute.Styled(data.ClientTransactionID)).
		Add("Created At", attribute.Styled(util.TimeOrDash(appCtx, &data.CreatedAt))).
		Add("Updated At", attribute.Styled(util.TimeOrDash(appCtx, &data.UpdatedAt))).
		Add("Reader Serial", attribute.Styled(data.ReaderSerialNumber)).
		Add("Firmware", attribute.Styled(data.ReaderFirmwareVersion))
	if cardType := data.CardType.Value(); cardType != nil {
		details.Add("Card Type", attribute.Styled(string(*cardType)))
	}
	if installments := data.Installments.Value(); installments != nil {
		details.Add("Installments", attribute.Styled(*installments))
	}
	if paymentStatus := data.PaymentStatus.Value(); paymentStatus != nil {
		details.Add("Payment Status", attribute.Styled(*paymentStatus))
	}
	if data.PaymentFailureReason != nil {
		if reason := data.PaymentFailureReason.Value(); reason != nil && *reason != "" {
			details.Add("Failure Reason", attribute.Styled(*reason))
		}
	}
	if validUntil := data.ValidUntil.Value(); validUntil != nil {
		details.Add("Valid Until", attribute.Styled(util.TimeOrDash(appCtx, validUntil)))
	}

	return details.Render(w)
}

package receipts

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "receipts",
		Usage: "Commands for retrieving transaction receipts.",
		Commands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "Get a receipt by transaction ID.",
				Action:    getReceipt,
				ArgsUsage: "<transaction-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the transaction. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.IntFlag{
						Name:  "transaction-event-id",
						Usage: "Transaction event ID for refund sumup.",
					},
				},
			},
		},
	}
}

func getReceipt(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	transactionID, err := util.RequireSingleArg(cmd, "transaction ID")
	if err != nil {
		return err
	}
	params := sumup.ReceiptsGetParams{
		Mid: merchantCode,
	}
	if cmd.IsSet("transaction-event-id") {
		value := cmd.Int("transaction-event-id")
		params.TxEventID = &value
	}

	receipt, err := appCtx.Client.Receipts.Get(ctx, transactionID, params)
	if err != nil {
		return fmt.Errorf("retrieve receipt: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, receipt)
	}

	return renderReceipt(appCtx, appCtx.Output, receipt)
}

func renderReceipt(appCtx *app.Context, w io.Writer, receipt *sumup.Receipt) error {
	sections := make([]display.Section, 0, 4)

	if transaction := receipt.TransactionData; transaction != nil {
		sections = append(sections, display.Section{
			Title: "Transaction",
			Pairs: []attribute.KeyValue{
				attribute.OptionalString("Code", transaction.TransactionCode),
				attribute.OptionalString("Status", transaction.Status),
				attribute.OptionalString("Payment Type", transaction.PaymentType),
				attribute.Attribute("Amount", attribute.Styled(receiptAmount(transaction))),
				attribute.Attribute("Timestamp", attribute.Styled(util.TimeOrDash(appCtx, transaction.Timestamp))),
				attribute.OptionalString("Entry Mode", transaction.EntryMode),
				attribute.OptionalString("Verification", transaction.VerificationMethod),
				attribute.Attribute("Card", attribute.Styled(receiptCard(transaction))),
			},
		})
	} else {
		sections = append(sections, display.Section{
			Title: "Transaction",
			Lines: []string{"-"},
		})
	}

	if merchant := receipt.MerchantData; merchant != nil {
		pairs := make([]attribute.KeyValue, 0, 5)
		lines := []string{}
		if profile := merchant.MerchantProfile; profile != nil {
			pairs = append(pairs, attribute.OptionalString("Name", profile.BusinessName))
			pairs = append(pairs, attribute.OptionalString("Code", profile.MerchantCode))
			if address := profile.Address; address != nil {
				if formatted := formatAddress(address); formatted != "" {
					pairs = append(pairs, attribute.Attribute("Address", attribute.Styled(formatted)))
				}
			}
			pairs = append(pairs, attribute.OptionalString("Email", profile.Email))
		} else {
			lines = append(lines, "Merchant profile unavailable")
		}
		if merchant.Locale != nil && *merchant.Locale != "" {
			pairs = append(pairs, attribute.Attribute("Locale", attribute.Styled(*merchant.Locale)))
		}
		sections = append(sections, display.Section{
			Title: "Merchant",
			Pairs: pairs,
			Lines: lines,
		})
	}

	if acquirer := receipt.AcquirerData; acquirer != nil {
		sections = append(sections, display.Section{
			Title: "Acquirer",
			Pairs: []attribute.KeyValue{
				attribute.OptionalString("Terminal ID", acquirer.Tid),
				attribute.OptionalString("Authorization Code", acquirer.AuthorizationCode),
				attribute.OptionalString("Return Code", acquirer.ReturnCode),
				attribute.OptionalString("Local Time", acquirer.LocalTime),
			},
		})
	}

	if transaction := receipt.TransactionData; transaction != nil && len(transaction.Events) > 0 {
		lines := make([]string, 0, len(transaction.Events))
		for _, event := range transaction.Events {
			lines = append(lines, fmt.Sprintf("- %s %s", enumValue(event.Type), enumValue(event.Status)))
		}
		sections = append(sections, display.Section{
			Title: fmt.Sprintf("Events (%d)", len(transaction.Events)),
			Lines: lines,
		})
	}

	return display.RenderSections(w, sections)
}

func receiptAmount(transaction *sumup.ReceiptTransaction) string {
	switch {
	case transaction.Amount != nil && transaction.Currency != nil:
		return fmt.Sprintf("%s %s", *transaction.Amount, *transaction.Currency)
	case transaction.Amount != nil:
		return *transaction.Amount
	default:
		return "-"
	}
}

func receiptCard(transaction *sumup.ReceiptTransaction) string {
	card := transaction.Card
	if card == nil {
		return "-"
	}
	switch {
	case card.Type != nil && card.Last4Digits != nil:
		return fmt.Sprintf("%s (****%s)", *card.Type, *card.Last4Digits)
	case card.Type != nil:
		return *card.Type
	case card.Last4Digits != nil:
		return fmt.Sprintf("****%s", *card.Last4Digits)
	default:
		return "-"
	}
}

func formatAddress(address *sumup.ReceiptMerchantDataMerchantProfileAddress) string {
	parts := []string{}
	if address.AddressLine1 != nil && *address.AddressLine1 != "" {
		parts = append(parts, *address.AddressLine1)
	}
	if address.City != nil && *address.City != "" {
		parts = append(parts, *address.City)
	}
	if address.PostCode != nil && *address.PostCode != "" {
		parts = append(parts, *address.PostCode)
	}
	if address.Country != nil && *address.Country != "" {
		parts = append(parts, *address.Country)
	}
	return strings.Join(parts, ", ")
}

func enumValue[T ~string](value *T) string {
	if value == nil {
		return "-"
	}
	return string(*value)
}

package receipts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/receipts"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "receipts",
		Usage: "Placeholder for the receipts API resource.",
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
						Usage: "Transaction event ID for refund receipts.",
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
	params := receipts.GetReceiptParams{
		Mid: merchantCode,
	}
	if cmd.IsSet("transaction-event-id") {
		value := cmd.Int("transaction-event-id")
		params.TxEventId = &value
	}

	receipt, err := appCtx.Client.Receipts.Get(ctx, transactionID, params)
	if err != nil {
		return fmt.Errorf("retrieve receipt: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(receipt)
	}

	renderReceipt(receipt)
	return nil
}

func renderReceipt(receipt *receipts.Receipt) {
	if transaction := receipt.TransactionData; transaction != nil {
		fmt.Println("Transaction")
		display.DataList([]attribute.KeyValue{
			attribute.Attribute("Code", attribute.Styled(util.StringOrDefault(transaction.TransactionCode, "-"))),
			attribute.Attribute("Status", attribute.Styled(util.StringOrDefault(transaction.Status, "-"))),
			attribute.Attribute("Payment Type", attribute.Styled(util.StringOrDefault(transaction.PaymentType, "-"))),
			attribute.Attribute("Amount", attribute.Styled(receiptAmount(transaction))),
			attribute.Attribute("Timestamp", attribute.Styled(timePointerToString(transaction.Timestamp))),
			attribute.Attribute("Entry Mode", attribute.Styled(util.StringOrDefault(transaction.EntryMode, "-"))),
			attribute.Attribute("Verification", attribute.Styled(util.StringOrDefault(transaction.VerificationMethod, "-"))),
			attribute.Attribute("Card", attribute.Styled(receiptCard(transaction))),
		})
	} else {
		fmt.Println("Transaction: -")
	}

	if merchant := receipt.MerchantData; merchant != nil {
		fmt.Println("\nMerchant")
		pairs := make([]attribute.KeyValue, 0, 5)
		if profile := merchant.MerchantProfile; profile != nil {
			pairs = append(pairs, attribute.Attribute("Name", attribute.Styled(util.StringOrDefault(profile.BusinessName, "-"))))
			pairs = append(pairs, attribute.Attribute("Code", attribute.Styled(util.StringOrDefault(profile.MerchantCode, "-"))))
			if address := profile.Address; address != nil {
				if formatted := formatAddress(address); formatted != "" {
					pairs = append(pairs, attribute.Attribute("Address", attribute.Styled(formatted)))
				}
			}
			pairs = append(pairs, attribute.Attribute("Email", attribute.Styled(util.StringOrDefault(profile.Email, "-"))))
		} else {
			fmt.Println("Merchant profile unavailable")
		}
		if merchant.Locale != nil && *merchant.Locale != "" {
			pairs = append(pairs, attribute.Attribute("Locale", attribute.Styled(*merchant.Locale)))
		}
		display.DataList(pairs)
	}

	if acquirer := receipt.AcquirerData; acquirer != nil {
		fmt.Println("\nAcquirer")
		display.DataList([]attribute.KeyValue{
			attribute.Attribute("Terminal ID", attribute.Styled(util.StringOrDefault(acquirer.Tid, "-"))),
			attribute.Attribute("Authorization Code", attribute.Styled(util.StringOrDefault(acquirer.AuthorizationCode, "-"))),
			attribute.Attribute("Return Code", attribute.Styled(util.StringOrDefault(acquirer.ReturnCode, "-"))),
			attribute.Attribute("Local Time", attribute.Styled(util.StringOrDefault(acquirer.LocalTime, "-"))),
		})
	}

	if transaction := receipt.TransactionData; transaction != nil {
		if len(transaction.Events) > 0 {
			fmt.Printf("\nEvents (%d)\n", len(transaction.Events))
			for _, event := range transaction.Events {
				fmt.Printf("  - %s %s\n", enumValue(event.Type), enumValue(event.Status))
			}
		}
	}
}

func receiptAmount(transaction *receipts.ReceiptTransaction) string {
	switch {
	case transaction.Amount != nil && transaction.Currency != nil:
		return fmt.Sprintf("%s %s", *transaction.Amount, *transaction.Currency)
	case transaction.Amount != nil:
		return *transaction.Amount
	default:
		return "-"
	}
}

func receiptCard(transaction *receipts.ReceiptTransaction) string {
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

func formatAddress(address *receipts.ReceiptMerchantDataMerchantProfileAddress) string {
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

func timePointerToString(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return value.UTC().Format(time.RFC3339)
}

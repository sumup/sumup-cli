package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/currency"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "transactions",
		Usage: "Placeholder for the transactions API resource.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List transactions for a merchant.",
				Action: listTransactions,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose transactions should be listed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of transactions to return.",
					},
					&cli.StringFlag{
						Name:  "changes-since",
						Usage: "Only return transactions modified at or after the timestamp (RFC3339).",
					},
					&cli.StringFlag{
						Name:  "newest-ref",
						Usage: "Return transactions whose event reference IDs are smaller than this value.",
					},
					&cli.StringFlag{
						Name:  "newest-time",
						Usage: "Return transactions created before this timestamp (RFC3339).",
					},
					&cli.StringFlag{
						Name:  "oldest-ref",
						Usage: "Return transactions whose event reference IDs are greater than this value.",
					},
					&cli.StringFlag{
						Name:  "oldest-time",
						Usage: "Return transactions created at or after this timestamp (RFC3339).",
					},
					&cli.StringFlag{
						Name:  "order",
						Usage: "Order in which results should be returned (e.g. asc, desc).",
					},
					&cli.StringSliceFlag{
						Name:  "payment-type",
						Usage: "Filter by payment type. May be specified multiple times.",
					},
					&cli.StringSliceFlag{
						Name:  "status",
						Usage: "Filter by transaction status. May be specified multiple times.",
					},
					&cli.StringFlag{
						Name:  "transaction-code",
						Usage: "Retrieve only the transaction matching the specified code.",
					},
					&cli.StringSliceFlag{
						Name:  "type",
						Usage: "Filter by transaction type. May be specified multiple times.",
					},
					&cli.StringSliceFlag{
						Name:  "user",
						Usage: "Filter by user email. May be specified multiple times.",
					},
				},
			},
			{
				Name:      "get",
				Usage:     "Get a specific transaction.",
				Action:    getTransaction,
				ArgsUsage: "[transaction-id]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the transaction. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:  "internal-id",
						Usage: "Lookup by internal transaction ID.",
					},
					&cli.StringFlag{
						Name:  "transaction-code",
						Usage: "Lookup by transaction code.",
					},
					&cli.StringFlag{
						Name:  "foreign-transaction-id",
						Usage: "Lookup by foreign transaction ID.",
					},
					&cli.StringFlag{
						Name:  "client-transaction-id",
						Usage: "Lookup by client transaction ID.",
					},
				},
			},
			{
				Name:      "refund",
				Usage:     "Refund a transaction fully or partially.",
				Action:    refundTransaction,
				ArgsUsage: "<transaction-id>",
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  "amount",
						Usage: "Optional partial refund amount in major units.",
					},
				},
			},
		},
	}
}

func listTransactions(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	params := sumup.TransactionsListParams{}
	if cmd.IsSet("limit") {
		value := cmd.Int("limit")
		params.Limit = &value
	}
	if ts, err := parseRFC3339Flag(cmd, "changes-since"); err != nil {
		return err
	} else if ts != nil {
		params.ChangesSince = ts
	}
	if cmd.IsSet("newest-ref") {
		value := cmd.String("newest-ref")
		params.NewestRef = &value
	}
	if ts, err := parseRFC3339Flag(cmd, "newest-time"); err != nil {
		return err
	} else if ts != nil {
		params.NewestTime = ts
	}
	if cmd.IsSet("oldest-ref") {
		value := cmd.String("oldest-ref")
		params.OldestRef = &value
	}
	if ts, err := parseRFC3339Flag(cmd, "oldest-time"); err != nil {
		return err
	} else if ts != nil {
		params.OldestTime = ts
	}
	if cmd.IsSet("order") {
		value := cmd.String("order")
		params.Order = &value
	}
	if values := cmd.StringSlice("payment-type"); len(values) > 0 {
		types := make([]sumup.PaymentType, 0, len(values))
		for _, v := range values {
			if v == "" {
				continue
			}
			types = append(types, sumup.PaymentType(v))
		}
		if len(types) > 0 {
			params.PaymentTypes = types
		}
	}
	if values := cmd.StringSlice("status"); len(values) > 0 {
		params.Statuses = values
	}
	if cmd.IsSet("transaction-code") {
		value := cmd.String("transaction-code")
		params.TransactionCode = &value
	}
	if values := cmd.StringSlice("type"); len(values) > 0 {
		params.Types = values
	}
	if values := cmd.StringSlice("user"); len(values) > 0 {
		params.Users = values
	}

	response, err := appCtx.Client.Transactions.List(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	items := response.Items
	if items == nil {
		items = []sumup.TransactionHistory{}
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(items)
	}

	rows := make([][]attribute.Value, 0, len(items))
	for _, tx := range items {
		rows = append(rows, []attribute.Value{
			attribute.OptionalStringValue(tx.ID),
			attribute.OptionalStringValue(tx.TransactionCode),
			attribute.ValueOf(currency.FormatPointers(tx.Amount, tx.Currency)),
			attribute.OptionalValue(tx.Status),
			attribute.OptionalValue(tx.PaymentType),
			attribute.ValueOf(util.TimeOrDash(appCtx, tx.Timestamp)),
		})
	}

	display.RenderTable(
		"Transactions",
		[]string{"ID", "Code", "Amount", "Status", "Payment Type", "Created At"},
		rows,
	)
	return nil
}

func getTransaction(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	params := sumup.TransactionsGetParams{}
	lookupCount := 0
	if cmd.Args().Len() > 0 {
		transactionID := cmd.Args().Get(0)
		params.ID = &transactionID
		lookupCount++
	}
	if value := cmd.String("internal-id"); value != "" {
		params.InternalID = &value
		lookupCount++
	}
	if value := cmd.String("transaction-code"); value != "" {
		params.TransactionCode = &value
		lookupCount++
	}
	if value := cmd.String("foreign-transaction-id"); value != "" {
		params.ForeignTransactionID = &value
		lookupCount++
	}
	if value := cmd.String("client-transaction-id"); value != "" {
		params.ClientTransactionID = &value
		lookupCount++
	}
	if lookupCount == 0 {
		return fmt.Errorf("provide a transaction ID argument or one lookup flag")
	}
	if lookupCount > 1 {
		return fmt.Errorf("provide exactly one transaction lookup")
	}

	transaction, err := appCtx.Client.Transactions.Get(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("retrieve transaction: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(transaction)
	}

	renderTransactionDetails(appCtx, transaction)
	return nil
}

func refundTransaction(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	transactionID, err := util.RequireSingleArg(cmd, "transaction ID")
	if err != nil {
		return err
	}

	body := sumup.TransactionsRefundParams{}
	if cmd.IsSet("amount") {
		value := float32(cmd.Float64("amount"))
		body.Amount = &value
	}

	if err := appCtx.Client.Transactions.Refund(ctx, transactionID, body); err != nil {
		return fmt.Errorf("refund transaction: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(map[string]string{"status": "refunded"})
	}

	return nil
}

func renderTransactionDetails(appCtx *app.Context, transaction *sumup.TransactionFull) {
	status := "-"
	if transaction.Status != nil && *transaction.Status != "" {
		status = string(*transaction.Status)
	}
	paymentType := "-"
	if transaction.PaymentType != nil && *transaction.PaymentType != "" {
		paymentType = string(*transaction.PaymentType)
	}

	display.DataList([]attribute.KeyValue{
		attribute.ID(util.StringOrDefault(transaction.ID, "-")),
		attribute.Attribute("Status", attribute.Styled(status)),
		attribute.OptionalString("Code", transaction.TransactionCode),
		attribute.Attribute("Amount", attribute.Styled(currency.FormatPointers(transaction.Amount, transaction.Currency))),
		attribute.OptionalString("Merchant", transaction.MerchantCode),
		attribute.Attribute("Payment Type", attribute.Styled(paymentType)),
		attribute.Attribute("Card", attribute.Styled(transactionCardLabel(transaction.Card))),
		attribute.OptionalString("Description", transaction.ProductSummary),
		attribute.Attribute("Created At", attribute.Styled(util.TimeOrDash(appCtx, transaction.Timestamp))),
	})
}

func transactionCardLabel(card *sumup.CardResponse) string {
	if card == nil {
		return "-"
	}
	var parts []string
	if card.Type != nil && *card.Type != "" {
		parts = append(parts, string(*card.Type))
	}
	if card.Last4Digits != nil && *card.Last4Digits != "" {
		parts = append(parts, fmt.Sprintf("(****%s)", *card.Last4Digits))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

func parseRFC3339Flag(cmd *cli.Command, name string) (*time.Time, error) {
	if !cmd.IsSet(name) {
		return nil, nil
	}
	value := cmd.String(name)
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for --%s: %w", name, err)
	}
	return &parsed, nil
}

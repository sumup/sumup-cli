package customers

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/customers"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "customers",
		Usage: "Commands for managing customers.",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "List saved payment instruments for a customer.",
				Action:    listPaymentInstruments,
				ArgsUsage: "<customer-id>",
			},
		},
	}
}

func listPaymentInstruments(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	customerID, err := util.RequireSingleArg(cmd, "customer ID")
	if err != nil {
		return err
	}
	instruments, err := appCtx.Client.Customers.ListPaymentInstruments(ctx, customerID)
	if err != nil {
		return fmt.Errorf("list customer payment instruments: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(instruments)
	}

	rows := make([][]string, 0, len(*instruments))
	for _, instrument := range *instruments {
		rows = append(rows, []string{
			util.StringOrDefault(instrument.Token, "-"),
			paymentInstrumentType(&instrument),
			lastFour(&instrument),
			util.BoolLabel(instrument.Active),
			util.TimeOrDash(appCtx, instrument.CreatedAt),
		})
	}

	display.RenderTable("Payment Instruments", []string{"Token", "Type", "Last 4", "Active", "Created At"}, rows)
	return nil
}

func paymentInstrumentType(instrument *customers.PaymentInstrumentResponse) string {
	if instrument.Type != nil {
		value := string(*instrument.Type)
		if value != "" {
			return value
		}
	}
	if instrument.Card != nil && instrument.Card.Type != nil {
		value := string(*instrument.Card.Type)
		if value != "" {
			return value
		}
	}
	return "-"
}

func lastFour(instrument *customers.PaymentInstrumentResponse) string {
	if instrument.Card != nil && instrument.Card.Last4Digits != nil && *instrument.Card.Last4Digits != "" {
		return *instrument.Card.Last4Digits
	}
	return "-"
}

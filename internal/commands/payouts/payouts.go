package payouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/datetime"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "payouts",
		Usage: "Placeholder for the payouts API resource.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List payouts for a merchant.",
				Action: listPayouts,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose payouts should be listed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:     "start-date",
						Usage:    "Start date (inclusive) in YYYY-MM-DD format.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "end-date",
						Usage:    "End date (inclusive) in YYYY-MM-DD format.",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of payouts to return.",
					},
					&cli.StringFlag{
						Name:  "order",
						Usage: "Sort payouts in ascending or descending order (asc, desc).",
					},
				},
			},
		},
	}
}

func listPayouts(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	startDate, err := parseDateArg(cmd.String("start-date"))
	if err != nil {
		return err
	}
	endDate, err := parseDateArg(cmd.String("end-date"))
	if err != nil {
		return err
	}
	params := sumup.PayoutsListParams{
		StartDate: startDate,
		EndDate:   endDate,
	}
	if cmd.IsSet("limit") {
		value := cmd.Int("limit")
		params.Limit = &value
	}
	if value := cmd.String("order"); value != "" {
		order := strings.ToLower(value)
		if order != "asc" && order != "desc" {
			return fmt.Errorf("invalid order %q, expected asc or desc", value)
		}
		params.Order = &order
	}

	payoutList, err := appCtx.Client.Payouts.List(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("list payouts: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(payoutList)
	}

	rows := make([][]attribute.Value, 0, len(*payoutList))
	for _, payout := range *payoutList {
		fee := attribute.OptionalValue(payout.Fee)
		if payout.Fee != nil {
			fee = attribute.ValueOf(fmt.Sprintf("%.2f", *payout.Fee))
		}

		rows = append(rows, []attribute.Value{
			attribute.OptionalValue(payout.ID),
			attribute.OptionalValue(payout.Date),
			attribute.ValueOf(payoutAmount(payout)),
			fee,
			attribute.OptionalValue(payout.Status),
			attribute.OptionalValue(payout.Type),
			attribute.OptionalStringValue(payout.Reference),
		})
	}

	display.RenderTable(
		"Payouts",
		[]string{"ID", "Date", "Amount", "Fee", "Status", "Type", "Reference"},
		rows,
	)
	return nil
}

func parseDateArg(value string) (datetime.Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return datetime.Date{}, fmt.Errorf("invalid date %q: %w", value, err)
	}
	return datetime.Date{Time: parsed}, nil
}

func payoutAmount(payout sumup.FinancialPayout) string {
	if payout.Amount == nil {
		return "-"
	}
	if payout.Currency == nil || *payout.Currency == "" {
		return fmt.Sprintf("%.2f", *payout.Amount)
	}
	return fmt.Sprintf("%.2f %s", *payout.Amount, *payout.Currency)
}

package payouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/datetime"
	"github.com/sumup/sumup-go/payouts"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
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
						Name:     "merchant-code",
						Usage:    "Merchant code whose payouts should be listed.",
						Sources:  cli.EnvVars("SUMUP_MERCHANT_CODE"),
						Required: true,
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
	startDate, err := parseDateArg(cmd.String("start-date"))
	if err != nil {
		return err
	}
	endDate, err := parseDateArg(cmd.String("end-date"))
	if err != nil {
		return err
	}
	params := payouts.ListPayoutsV1Params{
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

	payoutList, err := appCtx.Client.Payouts.List(ctx, cmd.String("merchant-code"), params)
	if err != nil {
		return fmt.Errorf("list payouts: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(payoutList)
	}

	rows := make([][]string, 0, len(*payoutList))
	for _, payout := range *payoutList {
		rows = append(rows, []string{
			intPointerToString(payout.ID),
			dateOrDash(payout.Date),
			payoutAmount(payout),
			floatPointerToString(payout.Fee),
			enumOrDash(payout.Status),
			enumOrDash(payout.Type),
			util.StringOrDefault(payout.Reference, "-"),
		})
	}

	display.RenderTable("Payouts", []string{"ID", "Date", "Amount", "Fee", "Status", "Type", "Reference"}, rows)
	return nil
}

func parseDateArg(value string) (datetime.Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return datetime.Date{}, fmt.Errorf("invalid date %q: %w", value, err)
	}
	return datetime.Date{Time: parsed}, nil
}

func intPointerToString(value *int) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *value)
}

func floatPointerToString(value *float32) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *value)
}

func enumOrDash[T ~string](value *T) string {
	if value == nil {
		return "-"
	}
	return string(*value)
}

func dateOrDash(value *datetime.Date) string {
	if value == nil {
		return "-"
	}
	return value.String()
}

func payoutAmount(payout payouts.FinancialPayout) string {
	if payout.Amount == nil {
		return "-"
	}
	if payout.Currency == nil || *payout.Currency == "" {
		return fmt.Sprintf("%.2f", *payout.Amount)
	}
	return fmt.Sprintf("%.2f %s", *payout.Amount, *payout.Currency)
}

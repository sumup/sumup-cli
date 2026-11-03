package sales

import (
	"context"
	"fmt"
	"time"

	sumup "github.com/sumup/sumup-go"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sales",
		Usage: "List, retrieve, and search fully-paid sales from the last two years.",
		Commands: []*cli.Command{
			apicommands.Bind("ListSalesV10", &cli.Command{
				Name: "list", Usage: "List sales, newest event first.",
				Action: listSales, Flags: pageFlags(),
			}),
			apicommands.Bind("GetSaleV10", &cli.Command{
				Name: "get", Usage: "Get a sale, including its items and payments.",
				ArgsUsage: "<sale-id>", Action: getSale,
				Flags: []cli.Flag{merchantFlag()},
			}),
			apicommands.Bind("SearchSalesV10", &cli.Command{
				Name: "search", Usage: "Search sales, newest close date first.",
				Description: "Filters from different categories combine with AND; repeated values within a category combine with OR. List and search pagination tokens are not interchangeable.",
				Action:      searchSales,
				Flags: append(pageFlags(),
					&cli.StringFlag{Name: "start-date", Usage: "Sale created on or after this RFC3339 timestamp (inclusive)."},
					&cli.StringFlag{Name: "end-date", Usage: "Sale created on or before this RFC3339 timestamp (inclusive)."},
					&cli.StringSliceFlag{Name: "payment-status", Usage: "Payment outcome filter (PAID, CHARGED_BACK). Repeat for multiple values."},
					&cli.StringSliceFlag{Name: "sale-status", Usage: "Sale lifecycle filter (CLOSED). Repeat for multiple values."},
					&cli.StringSliceFlag{Name: "sales-channel", Usage: "Sales channel filter (e.g. POS, SOLO, ONLINE_STORE). Repeat for multiple values."},
				),
			}),
		},
	}
}

func merchantFlag() cli.Flag {
	return &cli.StringFlag{
		Name: "merchant-code", Usage: "Merchant code that owns the sales. Falls back to context.",
		Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
	}
}

func pageFlags() []cli.Flag {
	return []cli.Flag{
		merchantFlag(),
		&cli.IntFlag{Name: "pagination-limit", Usage: "Maximum results per page (1-25; API default: 10)."},
		&cli.StringFlag{Name: "pagination-token", Usage: "Opaque next-page token from the previous response of the same command."},
	}
}

func pagination(cmd *cli.Command) (limit *int, token *string, err error) {
	if cmd.IsSet("pagination-limit") {
		value := cmd.Int("pagination-limit")
		if value < 1 || value > 25 {
			return nil, nil, fmt.Errorf("pagination-limit must be between 1 and 25")
		}
		limit = new(value)
	}
	if cmd.IsSet("pagination-token") {
		token = new(cmd.String("pagination-token"))
	}
	return limit, token, nil
}

func listSales(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchant, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	limit, token, err := pagination(cmd)
	if err != nil {
		return err
	}
	page, err := appCtx.Client.Sales.ListSalesV10(ctx, merchant, sumup.SalesListSalesV10Params{
		PaginationLimit: limit, PaginationToken: token,
	})
	if err != nil {
		return fmt.Errorf("list sales: %w", err)
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, page)
	}
	rows := make([][]attribute.Value, 0, len(page.Items))
	for _, sale := range page.Items {
		rows = append(rows, saleRow(appCtx, sale.ID, sale.Name, sale.State, sale.ClosedAt, sale.Price.GrossAmount.Value, sale.Price.GrossAmount.Currency))
	}
	return renderPage(appCtx, rows, page.PaginationToken)
}

func searchSales(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchant, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	limit, token, err := pagination(cmd)
	if err != nil {
		return err
	}
	params := sumup.SalesSearchSalesV10Params{PaginationLimit: limit, PaginationToken: token}
	for _, field := range []struct {
		name   string
		target **time.Time
	}{
		{"start-date", &params.StartDate}, {"end-date", &params.EndDate},
	} {
		if cmd.IsSet(field.name) {
			value, err := time.Parse(time.RFC3339, cmd.String(field.name))
			if err != nil {
				return fmt.Errorf("invalid %s: %w", field.name, err)
			}
			*field.target = new(value)
		}
	}
	if params.StartDate != nil && params.EndDate != nil && params.StartDate.After(*params.EndDate) {
		return fmt.Errorf("start-date must be on or before end-date")
	}
	for _, value := range cmd.StringSlice("payment-status") {
		params.PaymentStatuses = append(params.PaymentStatuses, sumup.PaymentStatusFilterV10(value))
	}
	for _, value := range cmd.StringSlice("sale-status") {
		params.SaleStatuses = append(params.SaleStatuses, sumup.SaleStateV10(value))
	}
	for _, value := range cmd.StringSlice("sales-channel") {
		params.SalesChannels = append(params.SalesChannels, sumup.SalesChannelFilterV10(value))
	}
	page, err := appCtx.Client.Sales.SearchSalesV10(ctx, merchant, params)
	if err != nil {
		return fmt.Errorf("search sales: %w", err)
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, page)
	}
	rows := make([][]attribute.Value, 0, len(page.Items))
	for _, sale := range page.Items {
		rows = append(rows, saleRow(appCtx, sale.ID, sale.Name, sale.State, sale.ClosedAt, sale.TotalAmount.Value, sale.TotalAmount.Currency))
	}
	return renderPage(appCtx, rows, page.PaginationToken)
}

func saleRow(appCtx *app.Context, id string, name *string, state sumup.SaleStateV10, closed time.Time, amount int, currency string) []attribute.Value {
	return []attribute.Value{
		attribute.ValueOf(id), attribute.ValueOf(util.StringOrDefault(name, "-")),
		attribute.ValueOf(state), attribute.ValueOf(util.TimeOrDash(appCtx, &closed)),
		attribute.ValueOf(amount), attribute.ValueOf(currency),
	}
}

func renderPage(appCtx *app.Context, rows [][]attribute.Value, token *string) error {
	if err := display.RenderTableWithOptions(appCtx.Output,
		[]string{"ID", "Name", "State", "Closed", "Amount (minor units)", "Currency"}, rows,
		display.TableOptions{Title: "Sales", EmptyText: "No items to display", IdentifierColumns: []int{0}},
	); err != nil {
		return err
	}
	if token != nil && *token != "" {
		_, err := fmt.Fprintf(appCtx.Output, "Next page: --pagination-token %s\n", *token)
		return err
	}
	return nil
}

func getSale(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchant, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	id, err := util.RequireSingleArg(cmd, "sale ID")
	if err != nil {
		return err
	}
	sale, err := appCtx.Client.Sales.GetSaleV10(ctx, merchant, id)
	if err != nil {
		return fmt.Errorf("get sale: %w", err)
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, sale)
	}
	sections := []display.Section{{Title: "Sale", Pairs: []attribute.KeyValue{
		attribute.ID(sale.ID),
		attribute.OptionalString("Name", sale.Name),
		attribute.Attribute("State", attribute.ValueOf(sale.State)),
		attribute.Attribute("Closed", attribute.ValueOf(util.TimeOrDash(appCtx, &sale.ClosedAt))),
		attribute.Attribute("Gross amount (minor units)", attribute.ValueOf(fmt.Sprintf("%d %s", sale.Price.GrossAmount.Value, sale.Price.GrossAmount.Currency))),
		attribute.Attribute("Channel", attribute.ValueOf(sale.Source.Channel)),
		attribute.OptionalString("Note", sale.Note),
	}}}
	items := display.Section{Title: "Items"}
	for _, item := range sale.Items {
		items.Lines = append(items.Lines, fmt.Sprintf("%s: %g × %s", item.ID, item.Quantity, item.Name))
	}
	payments := display.Section{Title: "Payments (minor units)"}
	for _, payment := range sale.Payments {
		payments.Lines = append(payments.Lines, fmt.Sprintf("%s: %s %d %s (%s)", payment.Reference, payment.Method, payment.Amount, sale.Currency, payment.Status))
	}
	return display.RenderSections(appCtx.Output, append(sections, items, payments))
}

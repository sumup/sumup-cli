package customers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/datetime"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "customers",
		Usage: "Commands for managing customers.",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create a customer.",
				Action: createCustomer,
				Flags:  customerDetailsFlags(),
			},
			{
				Name:      "get",
				Usage:     "Get a customer by ID.",
				Action:    getCustomer,
				ArgsUsage: "<customer-id>",
			},
			{
				Name:      "update",
				Usage:     "Update customer details.",
				Action:    updateCustomer,
				ArgsUsage: "<customer-id>",
				Flags:     customerDetailsFlags(),
			},
			{
				Name:  "payment-instruments",
				Usage: "Manage stored payment instruments for a customer.",
				Commands: []*cli.Command{
					{
						Name:      "list",
						Usage:     "List stored payment instruments for a customer.",
						Action:    listPaymentInstruments,
						ArgsUsage: "<customer-id>",
					},
					{
						Name:      "deactivate",
						Usage:     "Deactivate a stored payment instrument.",
						Action:    deactivatePaymentInstrument,
						ArgsUsage: "<customer-id> <token>",
					},
				},
			},
		},
	}
}

func customerDetailsFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "first-name",
			Usage: "Customer first name.",
		},
		&cli.StringFlag{
			Name:  "last-name",
			Usage: "Customer last name.",
		},
		&cli.StringFlag{
			Name:  "email",
			Usage: "Customer email address.",
		},
		&cli.StringFlag{
			Name:  "phone",
			Usage: "Customer phone number.",
		},
		&cli.StringFlag{
			Name:  "tax-id",
			Usage: "Customer tax identifier.",
		},
		&cli.StringFlag{
			Name:  "birth-date",
			Usage: "Customer birth date in YYYY-MM-DD format.",
		},
		&cli.StringFlag{
			Name:  "address-line-1",
			Usage: "Address line 1.",
		},
		&cli.StringFlag{
			Name:  "address-line-2",
			Usage: "Address line 2.",
		},
		&cli.StringFlag{
			Name:  "address-city",
			Usage: "Address city.",
		},
		&cli.StringFlag{
			Name:  "address-postal-code",
			Usage: "Address postal code.",
		},
		&cli.StringFlag{
			Name:  "address-state",
			Usage: "Address state.",
		},
		&cli.StringFlag{
			Name:  "address-country",
			Usage: "Address country code (ISO 3166-1 alpha-2).",
		},
	}
}

func createCustomer(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	personalDetails, _, err := customerDetailsFromFlags(cmd)
	if err != nil {
		return err
	}

	body := sumup.CustomersCreateParams{
		PersonalDetails: personalDetails,
	}
	customer, err := appCtx.Client.Customers.Create(ctx, body)
	if err != nil {
		return fmt.Errorf("create customer: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, customer)
	}

	if err := message.Success(appCtx.StatusOutput, "Customer created"); err != nil {
		return err
	}
	return renderCustomer(appCtx.Output, customer)
}

func getCustomer(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	customerID, err := util.RequireSingleArg(cmd, "customer ID")
	if err != nil {
		return err
	}

	customer, err := appCtx.Client.Customers.Get(ctx, customerID)
	if err != nil {
		return fmt.Errorf("get customer: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, customer)
	}

	return renderCustomer(appCtx.Output, customer)
}

func updateCustomer(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	customerID, err := util.RequireSingleArg(cmd, "customer ID")
	if err != nil {
		return err
	}

	personalDetails, changedCount, err := customerDetailsFromFlags(cmd)
	if err != nil {
		return err
	}
	if changedCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	body := sumup.CustomersUpdateParams{
		PersonalDetails: personalDetails,
	}
	customer, err := appCtx.Client.Customers.Update(ctx, customerID, body)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, customer)
	}

	if err := message.Success(appCtx.StatusOutput, "Customer updated"); err != nil {
		return err
	}
	return renderCustomer(appCtx.Output, customer)
}

func listPaymentInstruments(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("expected exactly 1 argument: customer ID")
	}

	customerID := cmd.Args().Get(0)
	instruments, err := appCtx.Client.Customers.ListPaymentInstruments(ctx, customerID)
	if err != nil {
		return fmt.Errorf("list payment instruments: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, instruments)
	}

	paymentInstruments := util.SliceOrEmpty(instruments)
	rows := make([][]attribute.Value, 0, len(paymentInstruments))
	for _, instrument := range paymentInstruments {
		rows = append(rows, []attribute.Value{
			attribute.OptionalStringValue(instrument.Token),
			attribute.OptionalValue(instrument.Type),
			attribute.ValueOf(paymentInstrumentCardLabel(instrument.Card)),
			attribute.ValueOf(util.BoolLabel(instrument.Active)),
			attribute.ValueOf(util.TimeOrDash(appCtx, instrument.CreatedAt)),
		})
	}

	return display.RenderTable(
		appCtx.Output,
		"Payment Instruments",
		[]string{"Token", "Type", "Card", "Active", "Created At"},
		rows,
	)
}

func deactivatePaymentInstrument(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	if cmd.Args().Len() != 2 {
		return fmt.Errorf("expected exactly 2 arguments: customer ID and token")
	}

	customerID := cmd.Args().Get(0)
	token := cmd.Args().Get(1)
	if err := appCtx.Client.Customers.DeactivatePaymentInstrument(ctx, customerID, token); err != nil {
		return fmt.Errorf("deactivate payment instrument: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, map[string]string{"status": "deactivated"})
	}

	return message.Success(appCtx.StatusOutput, "Payment instrument deactivated")
}

func customerDetailsFromFlags(cmd *cli.Command) (*sumup.PersonalDetails, int, error) {
	details := &sumup.PersonalDetails{}
	changedCount := 0

	if value := cmd.String("first-name"); value != "" {
		details.FirstName = &value
		changedCount++
	}
	if value := cmd.String("last-name"); value != "" {
		details.LastName = &value
		changedCount++
	}
	if value := cmd.String("email"); value != "" {
		details.Email = &value
		changedCount++
	}
	if value := cmd.String("phone"); value != "" {
		details.Phone = &value
		changedCount++
	}
	if value := cmd.String("tax-id"); value != "" {
		details.TaxID = &value
		changedCount++
	}
	if value := cmd.String("birth-date"); value != "" {
		parsedDate, err := parseDate(value)
		if err != nil {
			return nil, 0, err
		}
		details.BirthDate = parsedDate
		changedCount++
	}

	var address sumup.AddressLegacy
	addressChanged := false
	if value := cmd.String("address-line-1"); value != "" {
		address.Line1 = &value
		addressChanged = true
		changedCount++
	}
	if value := cmd.String("address-line-2"); value != "" {
		address.Line2 = &value
		addressChanged = true
		changedCount++
	}
	if value := cmd.String("address-city"); value != "" {
		address.City = &value
		addressChanged = true
		changedCount++
	}
	if value := cmd.String("address-postal-code"); value != "" {
		address.PostalCode = &value
		addressChanged = true
		changedCount++
	}
	if value := cmd.String("address-state"); value != "" {
		address.State = &value
		addressChanged = true
		changedCount++
	}
	if value := cmd.String("address-country"); value != "" {
		address.Country = &value
		addressChanged = true
		changedCount++
	}

	if addressChanged {
		details.Address = &address
	}
	if changedCount == 0 {
		return nil, 0, nil
	}

	return details, changedCount, nil
}

func parseDate(value string) (*datetime.Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", value, err)
	}
	date := datetime.Date{Time: parsed}
	return &date, nil
}

func renderCustomer(w io.Writer, customer *sumup.Customer) error {
	if customer == nil {
		return nil
	}

	details := []attribute.KeyValue{
		attribute.Attribute("Customer ID", attribute.Styled(customer.CustomerID)),
	}
	if customer.PersonalDetails == nil {
		return display.DataList(w, details)
	}

	personal := customer.PersonalDetails
	details = append(details, attribute.OptionalString("First Name", personal.FirstName))
	details = append(details, attribute.OptionalString("Last Name", personal.LastName))
	details = append(details, attribute.OptionalString("Email", personal.Email))
	details = append(details, attribute.OptionalString("Phone", personal.Phone))
	details = append(details, attribute.OptionalString("Tax ID", personal.TaxID))
	if personal.BirthDate != nil {
		birthDate := personal.BirthDate.Format(time.DateOnly)
		details = append(details, attribute.Attribute("Birth Date", attribute.Styled(birthDate)))
	} else {
		details = append(details, attribute.Attribute("Birth Date", attribute.Styled("-")))
	}
	details = append(details, attribute.Attribute("Address", attribute.Styled(formatAddress(personal.Address))))
	return display.DataList(w, details)
}

func formatAddress(address *sumup.AddressLegacy) string {
	if address == nil {
		return "-"
	}

	parts := make([]string, 0, 6)
	if address.Line1 != nil && *address.Line1 != "" {
		parts = append(parts, *address.Line1)
	}
	if address.Line2 != nil && *address.Line2 != "" {
		parts = append(parts, *address.Line2)
	}
	if address.City != nil && *address.City != "" {
		parts = append(parts, *address.City)
	}
	if address.PostalCode != nil && *address.PostalCode != "" {
		parts = append(parts, *address.PostalCode)
	}
	if address.State != nil && *address.State != "" {
		parts = append(parts, *address.State)
	}
	if address.Country != nil && *address.Country != "" {
		parts = append(parts, *address.Country)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

func paymentInstrumentCardLabel(card *sumup.PaymentInstrumentResponseCard) string {
	if card == nil {
		return "-"
	}

	parts := make([]string, 0, 2)
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

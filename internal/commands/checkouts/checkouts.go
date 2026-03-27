package checkouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/datetime"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/currency"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "checkouts",
		Usage: "Commands related to hosted sumup.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List checkout resources.",
				Action: listCheckouts,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "checkout-reference",
						Usage: "Filter results by checkout reference.",
					},
				},
			},
			{
				Name:  "create",
				Usage: "Create a new checkout resource.",
				Description: `Examples:
  sumup checkouts create --reference order-123 --amount 10 --currency EUR --merchant-code M123
  sumup checkouts create --reference ticket-42 --amount 29.99 --currency EUR --merchant-code M123 --description "Ticket" --return-url https://example.com/return`,
				Action: createCheckout,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "reference",
						Usage:    "Checkout reference that must be unique per merchant.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "amount",
						Usage:    "Amount to be charged.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "currency",
						Usage:    fmt.Sprintf("Currency for the checkout amount. Supported: %s", strings.Join(currency.Supported(), ", ")),
						Required: true,
					},
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the checkout. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Short description that will appear in the dashboard.",
					},
					&cli.StringFlag{
						Name:  "return-url",
						Usage: "URL SumUp should redirect to after payment.",
					},
					&cli.StringFlag{
						Name:  "redirect-url",
						Usage: "Optional URL for redirecting the payer after 3DS flows.",
					},
					&cli.StringFlag{
						Name:  "customer-id",
						Usage: "Attach the checkout to an existing customer.",
					},
					&cli.StringFlag{
						Name:  "purpose",
						Usage: "Optional purpose for the checkout.",
					},
				},
			},
			{
				Name:      "deactivate",
				Usage:     "Deactivate a checkout by ID.",
				Action:    deactivateCheckout,
				ArgsUsage: "<checkout-id>",
			},
			{
				Name:      "get",
				Usage:     "Get a checkout by ID.",
				Action:    getCheckout,
				ArgsUsage: "<checkout-id>",
			},
			{
				Name:   "payment-methods",
				Usage:  "List available payment methods for a merchant.",
				Action: listPaymentMethods,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the checkout. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:  "amount",
						Usage: "Optional amount filter.",
					},
					&cli.StringFlag{
						Name:  "currency",
						Usage: fmt.Sprintf("Optional currency filter. Supported: %s", strings.Join(currency.Supported(), ", ")),
					},
				},
			},
			{
				Name:      "process",
				Usage:     "Process a checkout.",
				Action:    processCheckout,
				ArgsUsage: "<checkout-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "payment-type",
						Usage:    "Payment type to use when processing the checkout.",
						Required: true,
					},
					&cli.StringFlag{Name: "customer-id", Usage: "Customer ID for tokenized payments."},
					&cli.StringFlag{Name: "token", Usage: "Saved payment instrument token."},
					&cli.IntFlag{Name: "installments", Usage: "Installment count for supported regions."},
					&cli.StringFlag{Name: "first-name", Usage: "Customer first name."},
					&cli.StringFlag{Name: "last-name", Usage: "Customer last name."},
					&cli.StringFlag{Name: "email", Usage: "Customer email."},
					&cli.StringFlag{Name: "phone", Usage: "Customer phone."},
					&cli.StringFlag{Name: "tax-id", Usage: "Customer tax ID."},
					&cli.StringFlag{Name: "birth-date", Usage: "Customer birth date in YYYY-MM-DD format."},
				},
			},
		},
	}
}

func listCheckouts(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	params := sumup.CheckoutsListParams{}
	if ref := cmd.String("checkout-reference"); ref != "" {
		params.CheckoutReference = &ref
	}

	checkoutList, err := appCtx.Client.Checkouts.List(ctx, params)
	if err != nil {
		return fmt.Errorf("list checkouts: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, checkoutList)
	}

	rows := make([][]attribute.Value, 0, len(*checkoutList))
	for _, checkout := range *checkoutList {
		status := "-"
		if checkout.Status != nil {
			status = string(*checkout.Status)
		}
		rows = append(rows, []attribute.Value{
			attribute.OptionalStringValue(checkout.ID),
			attribute.OptionalStringValue(checkout.CheckoutReference),
			attribute.ValueOf(currency.FormatPointers(checkout.Amount, checkout.Currency)),
			attribute.ValueOf(status),
			attribute.OptionalStringValue(checkout.MerchantCode),
			attribute.ValueOf(util.TimeOrDash(appCtx, checkout.Date)),
		})
	}

	display.RenderTable(
		appCtx.Output,
		"Checkouts",
		[]string{"ID", "Reference", "Amount", "Status", "Merchant", "Created At"},
		rows,
	)
	return nil
}

func createCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	parsedCurrency, err := currency.Parse(cmd.String("currency"))
	if err != nil {
		return err
	}

	body := sumup.CheckoutsCreateParams{
		CheckoutReference: cmd.String("reference"),
		Amount:            0,
		Currency:          parsedCurrency,
		MerchantCode:      merchantCode,
	}
	body.Amount, err = currency.ParseMajorUnitsForCurrency32(cmd.String("amount"), parsedCurrency)
	if err != nil {
		return err
	}

	if value := cmd.String("description"); value != "" {
		body.Description = &value
	}
	if value := cmd.String("return-url"); value != "" {
		body.ReturnURL = &value
	}
	if value := cmd.String("redirect-url"); value != "" {
		body.RedirectURL = &value
	}
	if value := cmd.String("customer-id"); value != "" {
		body.CustomerID = &value
	}
	if value := cmd.String("purpose"); value != "" {
		purpose := sumup.CheckoutCreateRequestPurpose(value)
		body.Purpose = &purpose
	}

	checkout, err := appCtx.Client.Checkouts.Create(ctx, body)
	if err != nil {
		return fmt.Errorf("create checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, checkout)
	}

	message.Success(appCtx.StatusOutput, "Checkout created")
	details := make([]attribute.KeyValue, 0, 5)
	if checkout.ID != nil {
		details = append(details, attribute.ID(*checkout.ID))
	}
	details = append(details, attribute.Attribute("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "N/A"))))
	details = append(details, attribute.Attribute("Amount", attribute.Styled(currency.FormatPointers(checkout.Amount, checkout.Currency))))
	if checkout.Status != nil {
		details = append(details, attribute.Attribute("Status", attribute.Styled(string(*checkout.Status))))
	}
	if checkout.Description != nil && *checkout.Description != "" {
		details = append(details, attribute.Attribute("Description", attribute.Styled(*checkout.Description)))
	}
	display.DataList(appCtx.Output, details)
	return nil
}

func deactivateCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	checkoutID, err := util.RequireSingleArg(cmd, "checkout ID")
	if err != nil {
		return err
	}
	checkout, err := appCtx.Client.Checkouts.Deactivate(ctx, checkoutID)
	if err != nil {
		return fmt.Errorf("deactivate checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, checkout)
	}

	message.Success(appCtx.StatusOutput, "Checkout deactivated")
	details := make([]attribute.KeyValue, 0, 4)
	if checkout.ID != nil {
		details = append(details, attribute.ID(*checkout.ID))
	}
	details = append(details, attribute.Attribute("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "N/A"))))
	if checkout.Status != nil {
		details = append(details, attribute.Attribute("Status", attribute.Styled(string(*checkout.Status))))
	}
	if checkout.ValidUntil != nil {
		if validUntil := checkout.ValidUntil.Value(); validUntil != nil {
			details = append(details, attribute.Attribute("Valid Until", attribute.Styled(validUntil.UTC().Format(time.RFC3339))))
		}
	}
	display.DataList(appCtx.Output, details)
	return nil
}

func getCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	checkoutID, err := util.RequireSingleArg(cmd, "checkout ID")
	if err != nil {
		return err
	}

	checkout, err := appCtx.Client.Checkouts.Get(ctx, checkoutID)
	if err != nil {
		return fmt.Errorf("get checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, checkout)
	}

	renderCheckout(appCtx, checkout)
	return nil
}

func listPaymentMethods(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	params := sumup.CheckoutsListAvailablePaymentMethodsParams{}
	if cmd.IsSet("amount") {
		if cmd.String("currency") == "" {
			return fmt.Errorf("--currency is required when --amount is set")
		}
	}
	if value := cmd.String("currency"); value != "" {
		parsedCurrency, err := currency.Parse(value)
		if err != nil {
			return err
		}
		if amount := cmd.String("amount"); amount != "" {
			parsedAmount, err := currency.ParseMajorUnitsForCurrency64(amount, parsedCurrency)
			if err != nil {
				return err
			}
			params.Amount = &parsedAmount
		}
		c := string(parsedCurrency)
		params.Currency = &c
	}

	methods, err := appCtx.Client.Checkouts.ListAvailablePaymentMethods(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("list checkout payment methods: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, methods.AvailablePaymentMethods)
	}

	rows := make([][]attribute.Value, 0, len(methods.AvailablePaymentMethods))
	for _, method := range methods.AvailablePaymentMethods {
		rows = append(rows, []attribute.Value{attribute.ValueOf(method.ID)})
	}

	display.RenderTableWithOptions(appCtx.Output, []string{"ID"}, rows, display.TableOptions{
		Title:              "Checkout Payment Methods",
		EmptyText:          "No payment methods available",
		HighlightIDColumns: true,
	})
	return nil
}

func processCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	checkoutID, err := util.RequireSingleArg(cmd, "checkout ID")
	if err != nil {
		return err
	}

	body := sumup.CheckoutsProcessParams{
		PaymentType: sumup.ProcessCheckoutPaymentType(cmd.String("payment-type")),
	}
	if customerID := cmd.String("customer-id"); customerID != "" {
		body.CustomerID = &customerID
	}
	if token := cmd.String("token"); token != "" {
		body.Token = &token
	}
	if cmd.IsSet("installments") {
		value := cmd.Int("installments")
		body.Installments = &value
	}
	if details, changedCount, err := checkoutPersonalDetailsFromFlags(cmd); err != nil {
		return err
	} else if changedCount > 0 {
		body.PersonalDetails = details
	}

	response, err := appCtx.Client.Checkouts.Process(ctx, checkoutID, body)
	if err != nil {
		return fmt.Errorf("process checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	if response.CheckoutSuccess != nil {
		message.Success(appCtx.StatusOutput, "Checkout processed")
		renderCheckout(appCtx, response.CheckoutSuccess)
		return nil
	}

	if response.CheckoutAccepted != nil {
		message.Success(appCtx.StatusOutput, "Checkout accepted")
		if response.CheckoutAccepted.NextStep != nil {
			display.DataList(appCtx.Output, []attribute.KeyValue{
				attribute.OptionalString("Method", response.CheckoutAccepted.NextStep.Method),
				attribute.OptionalString("URL", response.CheckoutAccepted.NextStep.URL),
				attribute.OptionalString("Redirect URL", response.CheckoutAccepted.NextStep.RedirectURL),
			})
			return nil
		}
	}
	return nil
}

func renderCheckout(appCtx *app.Context, checkout *sumup.CheckoutSuccess) {
	if checkout == nil {
		return
	}

	details := []attribute.KeyValue{}
	if checkout.ID != nil {
		details = append(details, attribute.ID(*checkout.ID))
	}
	details = append(details,
		attribute.Attribute("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "-"))),
		attribute.Attribute("Amount", attribute.Styled(currency.FormatPointers(checkout.Amount, checkout.Currency))),
		attribute.OptionalString("Merchant", checkout.MerchantCode),
		attribute.OptionalString("Merchant Name", checkout.MerchantName),
		attribute.OptionalString("Description", checkout.Description),
		attribute.OptionalString("Status", enumString(checkout.Status)),
		attribute.OptionalString("Transaction ID", checkout.TransactionID),
		attribute.OptionalString("Transaction Code", checkout.TransactionCode),
		attribute.Attribute("Created At", attribute.Styled(util.TimeOrDash(appCtx, checkout.Date))),
	)
	display.DataList(appCtx.Output, details)
}

func checkoutPersonalDetailsFromFlags(cmd *cli.Command) (*sumup.PersonalDetails, int, error) {
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
		parsedDate, err := time.Parse(time.DateOnly, value)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid birth date %q: %w", value, err)
		}
		date := datetime.Date{Time: parsedDate}
		details.BirthDate = &date
		changedCount++
	}

	if changedCount == 0 {
		return nil, 0, nil
	}
	return details, changedCount, nil
}

func enumString[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	text := string(*value)
	return &text
}

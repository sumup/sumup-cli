package checkouts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/nullable"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/currency"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "checkouts",
		Usage: "Commands related to hosted sumup.",
		Commands: []*cli.Command{
			apicommands.Bind("ListCheckouts", &cli.Command{
				Name:   "list",
				Usage:  "List checkout resources.",
				Action: listCheckouts,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "checkout-reference",
						Usage: "Filter results by checkout reference.",
					},
				},
			}),
			apicommands.Bind("CreateCheckout", &cli.Command{
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
						Name:  "valid-until",
						Usage: "Expiration timestamp in RFC3339 format.",
					},
					&cli.StringFlag{
						Name:  "purpose",
						Usage: "Optional purpose for the checkout.",
					},
					&cli.BoolFlag{
						Name:  "hosted-checkout",
						Usage: "Enable the SumUp-hosted checkout page and return its URL.",
					},
				},
			}),
			apicommands.Bind("DeactivateCheckout", &cli.Command{
				Name:      "deactivate",
				Usage:     "Deactivate a checkout by ID.",
				Action:    deactivateCheckout,
				ArgsUsage: "<checkout-id>",
			}),
			apicommands.Bind("CreateApplePaySession", &cli.Command{
				Name:      "apple-pay-session",
				Usage:     "Create an Apple Pay merchant session for a checkout.",
				Action:    createApplePaySession,
				ArgsUsage: "<checkout-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "context",
						Usage:    "Hostname requesting the Apple Pay session.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Apple Pay validation URL from the browser.",
						Required: true,
					},
				},
			}),
			apicommands.Bind("GetCheckout", &cli.Command{
				Name:      "get",
				Usage:     "Get a checkout by ID.",
				Action:    getCheckout,
				ArgsUsage: "<checkout-id>",
			}),
			apicommands.Bind("UpdateCheckout", &cli.Command{
				Name:  "update",
				Usage: "Update a checkout by ID.",
				Description: `Examples:
  sumup checkouts update checkout-123 --description "Updated ticket"
  sumup checkouts update checkout-123 --amount 29.99 --currency EUR --valid-until 2026-07-12T15:04:05Z`,
				Action:    updateCheckout,
				ArgsUsage: "<checkout-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "reference",
						Usage: "Updated checkout reference.",
					},
					&cli.StringFlag{
						Name:  "amount",
						Usage: "Updated amount to be charged.",
					},
					&cli.StringFlag{
						Name:  "currency",
						Usage: fmt.Sprintf("Updated checkout currency. Supported: %s", strings.Join(currency.Supported(), ", ")),
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Updated short description.",
					},
					&cli.StringFlag{
						Name:  "customer-id",
						Usage: "Updated customer ID.",
					},
					&cli.StringFlag{
						Name:  "valid-until",
						Usage: "Updated expiration timestamp in RFC3339 format.",
					},
					&cli.BoolFlag{
						Name:  "clear-valid-until",
						Usage: "Clear the checkout expiration timestamp.",
					},
				},
			}),
			apicommands.Bind("GetPaymentMethods", &cli.Command{
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
			}),
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

	checkouts := util.SliceOrEmpty(checkoutList)
	rows := make([][]attribute.Value, 0, len(checkouts))
	for _, checkout := range checkouts {
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

	return display.RenderTableWithOptions(appCtx.Output, []string{"ID", "Reference", "Amount", "Status", "Merchant", "Created At"}, rows, display.TableOptions{
		Title:             "Checkouts",
		EmptyText:         "No items to display",
		IdentifierColumns: []int{0},
	})
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
	if value := cmd.String("valid-until"); value != "" {
		validUntil, err := parseRFC3339Time(value, "valid-until")
		if err != nil {
			return err
		}
		body.ValidUntil = nullable.Value(validUntil)
	}
	if value := cmd.String("purpose"); value != "" {
		purpose := sumup.CheckoutCreateRequestPurpose(value)
		body.Purpose = &purpose
	}
	if cmd.Bool("hosted-checkout") {
		body.HostedCheckout = &sumup.HostedCheckout{Enabled: true}
	}

	checkout, err := appCtx.Client.Checkouts.Create(ctx, body)
	if err != nil {
		return fmt.Errorf("create checkout: %w", err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      checkout,
		SuccessMessage: "Checkout created",
		Details:        checkoutMutationDetails(appCtx, checkout),
	})
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

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      checkout,
		SuccessMessage: "Checkout deactivated",
		Details:        checkoutMutationDetails(appCtx, checkout),
	})
}

func createApplePaySession(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	checkoutID, err := util.RequireSingleArg(cmd, "checkout ID")
	if err != nil {
		return err
	}

	body := sumup.CheckoutsCreateApplePaySessionParams{
		Context: cmd.String("context"),
		Target:  cmd.String("target"),
	}
	session, err := appCtx.Client.Checkouts.CreateApplePaySession(ctx, checkoutID, body)
	if err != nil {
		return fmt.Errorf("create Apple Pay session: %w", err)
	}
	if session == nil {
		return nil
	}

	rawSession := json.RawMessage(*session)
	return display.PrintJSON(appCtx.Output, rawSession)
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

	return renderCheckout(appCtx, checkout)
}

func updateCheckout(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	checkoutID, err := util.RequireSingleArg(cmd, "checkout ID")
	if err != nil {
		return err
	}

	body := sumup.CheckoutsUpdateParams{}
	changedCount := 0
	if cmd.IsSet("reference") {
		value := cmd.String("reference")
		body.CheckoutReference = &value
		changedCount++
	}
	var parsedCurrency sumup.Currency
	if cmd.IsSet("currency") {
		parsedCurrency, err = currency.Parse(cmd.String("currency"))
		if err != nil {
			return err
		}
		body.Currency = &parsedCurrency
		changedCount++
	}
	if cmd.IsSet("amount") {
		var amount float32
		if body.Currency != nil {
			amount, err = currency.ParseMajorUnitsForCurrency32(cmd.String("amount"), *body.Currency)
		} else {
			amount, err = currency.ParseMajorUnits32(cmd.String("amount"))
		}
		if err != nil {
			return err
		}
		body.Amount = &amount
		changedCount++
	}
	if cmd.IsSet("description") {
		value := cmd.String("description")
		body.Description = &value
		changedCount++
	}
	if cmd.IsSet("customer-id") {
		value := cmd.String("customer-id")
		body.CustomerID = &value
		changedCount++
	}
	if cmd.Bool("clear-valid-until") && cmd.String("valid-until") != "" {
		return fmt.Errorf("--clear-valid-until cannot be used with --valid-until")
	}
	if cmd.Bool("clear-valid-until") {
		body.ValidUntil = nullable.Null[time.Time]()
		changedCount++
	} else if value := cmd.String("valid-until"); value != "" {
		validUntil, err := parseRFC3339Time(value, "valid-until")
		if err != nil {
			return err
		}
		body.ValidUntil = nullable.Value(validUntil)
		changedCount++
	}
	if changedCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	checkout, err := appCtx.Client.Checkouts.Update(ctx, checkoutID, body)
	if err != nil {
		return fmt.Errorf("update checkout: %w", err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      checkout,
		SuccessMessage: "Checkout updated",
		Details:        checkoutMutationDetails(appCtx, checkout),
	})
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

	return display.RenderTableWithOptions(appCtx.Output, []string{"ID"}, rows, display.TableOptions{
		Title:             "Checkout Payment Methods",
		EmptyText:         "No payment methods available",
		IdentifierColumns: []int{0},
	})
}

func renderCheckout(appCtx *app.Context, checkout *sumup.CheckoutSuccess) error {
	if checkout == nil {
		return nil
	}

	details := display.NewDetailsBuilder()
	if checkout.ID != nil {
		details.AddID(*checkout.ID)
	}
	details.
		Add("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "-"))).
		Add("Amount", attribute.Styled(currency.FormatPointers(checkout.Amount, checkout.Currency))).
		AddOptionalString("Merchant", checkout.MerchantCode).
		AddOptionalString("Merchant Name", checkout.MerchantName).
		AddOptionalString("Description", checkout.Description).
		AddOptionalString("Hosted Checkout URL", checkout.HostedCheckoutURL).
		AddOptionalString("Status", enumString(checkout.Status)).
		AddOptionalString("Transaction ID", checkout.TransactionID).
		AddOptionalString("Transaction Code", checkout.TransactionCode).
		Add("Created At", attribute.Styled(util.TimeOrDash(appCtx, checkout.Date)))
	return details.Render(appCtx.Output)
}

func checkoutMutationDetails(appCtx *app.Context, checkout *sumup.Checkout) []attribute.KeyValue {
	details := display.NewDetailsBuilder()
	if checkout == nil {
		return details.Pairs()
	}
	if checkout.ID != nil {
		details.AddID(*checkout.ID)
	}
	details.
		Add("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "N/A"))).
		Add("Amount", attribute.Styled(currency.FormatPointers(checkout.Amount, checkout.Currency))).
		AddOptionalString("Merchant", checkout.MerchantCode).
		AddOptionalString("Description", checkout.Description).
		AddOptionalString("Hosted Checkout URL", checkout.HostedCheckoutURL).
		AddOptionalString("Status", enumString(checkout.Status))
	if checkout.ValidUntil != nil {
		if validUntil := checkout.ValidUntil.Value(); validUntil != nil {
			details.Add("Valid Until", attribute.Styled(util.TimeOrDash(appCtx, validUntil)))
		}
	}
	return details.Pairs()
}

func parseRFC3339Time(value string, label string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s %q: expected RFC3339 timestamp: %w", label, value, err)
	}
	return parsed, nil
}

func enumString[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	text := string(*value)
	return &text
}

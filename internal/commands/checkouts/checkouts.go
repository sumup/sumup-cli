package checkouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/checkouts"

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
		Usage: "Commands related to hosted checkouts.",
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
  sumup-cli checkouts create --reference order-123 --amount 10 --currency EUR --merchant-code M123
  sumup-cli checkouts create --reference ticket-42 --amount 29.99 --currency EUR --merchant-code M123 --description "Ticket" --return-url https://example.com/return`,
				Action: createCheckout,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "reference",
						Usage:    "Checkout reference that must be unique per merchant.",
						Required: true,
					},
					&cli.Float64Flag{
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
		},
	}
}

func listCheckouts(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	params := checkouts.ListCheckoutsParams{}
	if ref := cmd.String("checkout-reference"); ref != "" {
		params.CheckoutReference = &ref
	}

	checkoutList, err := appCtx.Client.Checkouts.List(ctx, params)
	if err != nil {
		return fmt.Errorf("list checkouts: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(checkoutList)
	}

	rows := make([][]string, 0, len(*checkoutList))
	for _, checkout := range *checkoutList {
		status := "-"
		if checkout.Status != nil {
			status = string(*checkout.Status)
		}
		rows = append(rows, []string{
			util.StringOrDefault(checkout.ID, "-"),
			util.StringOrDefault(checkout.CheckoutReference, "-"),
			currency.FormatPointers(checkout.Amount, checkout.Currency),
			status,
			util.StringOrDefault(checkout.MerchantCode, "-"),
			util.TimeOrDash(appCtx, checkout.Date),
		})
	}

	display.RenderTable("Checkouts", []string{"ID", "Reference", "Amount", "Status", "Merchant", "Created At"}, rows)
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

	body := checkouts.CreateCheckoutBody{
		CheckoutReference: cmd.String("reference"),
		Amount:            float32(cmd.Float64("amount")),
		Currency:          parsedCurrency,
		MerchantCode:      merchantCode,
	}

	if value := cmd.String("description"); value != "" {
		body.Description = &value
	}
	if value := cmd.String("return-url"); value != "" {
		body.ReturnUrl = &value
	}
	if value := cmd.String("redirect-url"); value != "" {
		body.RedirectUrl = &value
	}
	if value := cmd.String("customer-id"); value != "" {
		body.CustomerId = &value
	}
	if value := cmd.String("purpose"); value != "" {
		purpose := checkouts.CreateCheckoutBodyPurpose(value)
		body.Purpose = &purpose
	}

	checkout, err := appCtx.Client.Checkouts.Create(ctx, body)
	if err != nil {
		return fmt.Errorf("create checkout: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(checkout)
	}

	message.Success("Checkout created")
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
	display.DataList(details)
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
		return display.PrintJSON(checkout)
	}

	message.Success("Checkout deactivated")
	details := make([]attribute.KeyValue, 0, 4)
	if checkout.ID != nil {
		details = append(details, attribute.ID(*checkout.ID))
	}
	details = append(details, attribute.Attribute("Reference", attribute.Styled(util.StringOrDefault(checkout.CheckoutReference, "N/A"))))
	if checkout.Status != nil {
		details = append(details, attribute.Attribute("Status", attribute.Styled(string(*checkout.Status))))
	}
	if checkout.ValidUntil != nil {
		details = append(details, attribute.Attribute("Valid Until", attribute.Styled(checkout.ValidUntil.UTC().Format(time.RFC3339))))
	}
	display.DataList(details)
	return nil
}

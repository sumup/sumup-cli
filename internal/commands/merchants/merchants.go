package merchants

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "merchants",
		Usage: "Commands related to merchant accounts.",
		Commands: []*cli.Command{
			apicommands.Bind("GetMerchant", &cli.Command{
				Name:   "get",
				Usage:  "Get merchant information.",
				Action: getMerchant,
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code to retrieve information for. Falls back to context."),
					versionFlag(),
				},
			}),
			{
				Name:  "persons",
				Usage: "Commands related to people associated with a merchant.",
				Commands: []*cli.Command{
					apicommands.Bind("ListPersons", &cli.Command{
						Name:   "list",
						Usage:  "List people associated with a merchant.",
						Action: listPersons,
						Flags: []cli.Flag{
							merchantCodeFlag("Merchant code whose people should be listed. Falls back to context."),
							versionFlag(),
						},
					}),
					apicommands.Bind("GetPerson", &cli.Command{
						Name:      "get",
						Usage:     "Get a person associated with a merchant.",
						ArgsUsage: "<person-id>",
						Action:    getPerson,
						Flags: []cli.Flag{
							merchantCodeFlag("Merchant code associated with the person. Falls back to context."),
							versionFlag(),
						},
					}),
				},
			},
		},
	}
}

func merchantCodeFlag(usage string) cli.Flag {
	return &cli.StringFlag{
		Name:    "merchant-code",
		Usage:   usage,
		Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
	}
}

func versionFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "version",
		Usage: "Resource version to retrieve. Currently only latest is supported.",
	}
}

func getMerchant(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	version, err := resourceVersion(cmd)
	if err != nil {
		return err
	}

	merchant, err := appCtx.Client.Merchants.Get(ctx, merchantCode, sumup.MerchantsGetParams{Version: version})
	if err != nil {
		return fmt.Errorf("get merchant: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, merchant)
	}

	return renderMerchant(appCtx, merchant)
}

func listPersons(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	version, err := resourceVersion(cmd)
	if err != nil {
		return err
	}

	response, err := appCtx.Client.Merchants.ListPersons(ctx, merchantCode, sumup.MerchantsListPersonsParams{Version: version})
	if err != nil {
		return fmt.Errorf("list persons: %w", err)
	}
	if response == nil {
		return nil
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response.Items)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for index := range response.Items {
		person := &response.Items[index]
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(person.ID),
			attribute.ValueOf(personName(person)),
			attribute.ValueOf(personRelationships(person)),
			attribute.OptionalStringValue(person.UserID),
		})
	}

	return display.RenderTableWithOptions(appCtx.Output, []string{"Person", "Name", "Relationships", "User ID"}, rows, display.TableOptions{
		Title:             "Persons",
		EmptyText:         "No items to display",
		IdentifierColumns: []int{0},
	})
}

func getPerson(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	personID, err := util.RequireSingleArg(cmd, "person ID")
	if err != nil {
		return err
	}
	version, err := resourceVersion(cmd)
	if err != nil {
		return err
	}

	person, err := appCtx.Client.Merchants.GetPerson(ctx, merchantCode, personID, sumup.MerchantsGetPersonParams{Version: version})
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, person)
	}
	return renderPerson(appCtx.Output, person)
}

func resourceVersion(cmd *cli.Command) (*string, error) {
	if !cmd.IsSet("version") {
		return nil, nil
	}

	version := strings.TrimSpace(cmd.String("version"))
	if version != "latest" {
		return nil, fmt.Errorf("version must be latest")
	}
	return &version, nil
}

func renderMerchant(appCtx *app.Context, merchant *sumup.Merchant) error {
	if merchant == nil {
		return nil
	}

	details := []attribute.KeyValue{
		attribute.Attribute("Merchant Code", attribute.Styled(merchant.MerchantCode)),
		attribute.OptionalString("Alias", merchant.Alias),
		attribute.Attribute("Country", attribute.Styled(merchant.Country)),
		attribute.Attribute("Default Currency", attribute.Styled(merchant.DefaultCurrency)),
		attribute.Attribute("Default Locale", attribute.Styled(merchant.DefaultLocale)),
		attribute.Attribute("Sandbox", attribute.Styled(util.BoolLabel(merchant.Sandbox))),
		attribute.OptionalString("Organization ID", merchant.OrganizationID),
		attribute.Attribute("Created At", attribute.Styled(util.TimeOrDash(appCtx, &merchant.CreatedAt))),
		attribute.Attribute("Updated At", attribute.Styled(util.TimeOrDash(appCtx, &merchant.UpdatedAt))),
	}

	if merchant.BusinessProfile != nil {
		details = append(details,
			attribute.OptionalString("Business Name", merchant.BusinessProfile.Name),
			attribute.OptionalString("Business Email", merchant.BusinessProfile.Email),
			attribute.Optional("Business Phone", merchant.BusinessProfile.PhoneNumber),
			attribute.OptionalString("Business Website", merchant.BusinessProfile.Website),
		)
	}

	return display.DataList(appCtx.Output, details)
}

func renderPerson(w io.Writer, person *sumup.Person) error {
	if person == nil {
		return nil
	}

	return display.DataList(w, []attribute.KeyValue{
		attribute.ID(person.ID),
		attribute.OptionalString("Given Name", person.GivenName),
		attribute.OptionalString("Middle Name", person.MiddleName),
		attribute.OptionalString("Family Name", person.FamilyName),
		attribute.Attribute("Relationships", attribute.Styled(personRelationships(person))),
		attribute.OptionalString("User ID", person.UserID),
		attribute.Optional("Citizenship", person.Citizenship),
		attribute.Attribute("Birthdate", attribute.Styled(personBirthdate(person))),
		attribute.Optional("Phone Number", person.PhoneNumber),
		attribute.Optional("Version", person.Version),
		attribute.Optional("Change Status", person.ChangeStatus),
	})
}

func personName(person *sumup.Person) string {
	if person == nil {
		return "-"
	}

	parts := make([]string, 0, 3)
	for _, part := range []*string{person.GivenName, person.MiddleName, person.FamilyName} {
		if part != nil && strings.TrimSpace(*part) != "" {
			parts = append(parts, strings.TrimSpace(*part))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

func personRelationships(person *sumup.Person) string {
	if person == nil || len(person.Relationships) == 0 {
		return "-"
	}
	return strings.Join(person.Relationships, ", ")
}

func personBirthdate(person *sumup.Person) string {
	if person == nil || person.Birthdate == nil {
		return "-"
	}
	return person.Birthdate.Format(time.DateOnly)
}

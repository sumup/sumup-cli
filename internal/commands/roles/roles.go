package roles

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "roles",
		Usage: "Commands for managing roles.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List available roles.",
				Action: listRoles,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose roles should be listed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			},
		},
	}
}

func listRoles(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	response, err := appCtx.Client.Roles.List(ctx, merchantCode)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response.Items)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for _, role := range response.Items {
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(role.ID),
			attribute.ValueOf(role.Name),
			attribute.ValueOf(roleDescription(role)),
		})
	}

	return display.RenderTableWithOptions(appCtx.Output, []string{"Role", "Name", "Description"}, rows, display.TableOptions{
		Title:             "Roles",
		EmptyText:         "No items to display",
		IdentifierColumns: []int{0},
	})
}

func roleDescription(role sumup.Role) string {
	if role.Description == nil || *role.Description == "" {
		return "-"
	}

	return *role.Description
}

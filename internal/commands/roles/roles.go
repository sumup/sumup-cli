package roles

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "roles",
		Usage: "Commands for managing merchant roles.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List roles for a merchant.",
				Action: listRoles,
				Flags:  merchantCodeFlags(),
			},
			{
				Name:      "get",
				Usage:     "Get a role by ID.",
				Action:    getRole,
				ArgsUsage: "<role-id>",
				Flags:     merchantCodeFlags(),
			},
			{
				Name:   "create",
				Usage:  "Create a custom role.",
				Action: createRole,
				Flags: append(
					merchantCodeFlags(),
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Role name.",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Role description.",
					},
					&cli.StringSliceFlag{
						Name:     "permission",
						Usage:    "Permission granted by the role. Repeat the flag to provide multiple permissions.",
						Required: true,
					},
				),
			},
			{
				Name:      "update",
				Usage:     "Update a custom role.",
				Action:    updateRole,
				ArgsUsage: "<role-id>",
				Flags: append(
					merchantCodeFlags(),
					&cli.StringFlag{
						Name:  "name",
						Usage: "Updated role name.",
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Updated role description.",
					},
					&cli.StringSliceFlag{
						Name:  "permission",
						Usage: "Updated set of permissions. Repeat the flag to provide multiple permissions.",
					},
				),
			},
			{
				Name:      "delete",
				Usage:     "Delete a custom role.",
				Action:    deleteRole,
				ArgsUsage: "<role-id>",
				Flags:     merchantCodeFlags(),
			},
		},
	}
}

func merchantCodeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "merchant-code",
			Usage:   "Merchant code that owns the roles. Falls back to context.",
			Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
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
		return display.PrintJSON(response.Items)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for _, role := range response.Items {
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(role.ID),
			attribute.ValueOf(role.Name),
			attribute.OptionalStringValue(role.Description),
			attribute.ValueOf(strings.Join(role.Permissions, ", ")),
			attribute.ValueOf(boolLabel(role.IsPredefined)),
		})
	}

	display.RenderTable(
		"Roles",
		[]string{"ID", "Name", "Description", "Permissions", "Predefined"},
		rows,
	)
	return nil
}

func getRole(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	roleID, err := util.RequireSingleArg(cmd, "role ID")
	if err != nil {
		return err
	}

	role, err := appCtx.Client.Roles.Get(ctx, merchantCode, roleID)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(role)
	}

	renderRole(role)
	return nil
}

func createRole(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	body := sumup.RolesCreateParams{
		Name:        cmd.String("name"),
		Permissions: cmd.StringSlice("permission"),
	}
	if description := cmd.String("description"); description != "" {
		body.Description = &description
	}

	role, err := appCtx.Client.Roles.Create(ctx, merchantCode, body)
	if err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(role)
	}

	message.Success("Role created")
	renderRole(role)
	return nil
}

func updateRole(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	roleID, err := util.RequireSingleArg(cmd, "role ID")
	if err != nil {
		return err
	}

	body := sumup.RolesUpdateParams{}
	changeCount := 0

	if name := cmd.String("name"); name != "" {
		body.Name = &name
		changeCount++
	}
	if description := cmd.String("description"); description != "" {
		body.Description = &description
		changeCount++
	}
	if permissions := cmd.StringSlice("permission"); len(permissions) > 0 {
		body.Permissions = permissions
		changeCount++
	}
	if changeCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	role, err := appCtx.Client.Roles.Update(ctx, merchantCode, roleID, body)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(role)
	}

	message.Success("Role updated")
	renderRole(role)
	return nil
}

func deleteRole(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	roleID, err := util.RequireSingleArg(cmd, "role ID")
	if err != nil {
		return err
	}

	if err := appCtx.Client.Roles.Delete(ctx, merchantCode, roleID); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(map[string]string{"status": "deleted"})
	}

	message.Success("Role deleted")
	return nil
}

func renderRole(role *sumup.Role) {
	if role == nil {
		return
	}

	display.DataList([]attribute.KeyValue{
		attribute.ID(role.ID),
		attribute.Attribute("Name", attribute.Styled(role.Name)),
		attribute.Attribute("Description", attribute.Styled(util.StringOrDefault(role.Description, "-"))),
		attribute.Attribute("Permissions", attribute.Styled(strings.Join(role.Permissions, ", "))),
		attribute.Attribute("Predefined", attribute.Styled(boolLabel(role.IsPredefined))),
		attribute.Attribute("Created At", attribute.Styled(role.CreatedAt.UTC().Format(time.RFC3339))),
		attribute.Attribute("Updated At", attribute.Styled(role.UpdatedAt.UTC().Format(time.RFC3339))),
	})
}

func boolLabel(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

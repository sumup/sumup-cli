package roles

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "roles",
		Usage: "Commands for managing roles.",
		Commands: []*cli.Command{
			apicommands.Bind("ListMerchantRoles", &cli.Command{
				Name:   "list",
				Usage:  "List available roles.",
				Action: listRoles,
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code whose roles should be listed. Falls back to context."),
				},
			}),
			apicommands.Bind("CreateMerchantRole", &cli.Command{
				Name:   "create",
				Usage:  "Create a custom role.",
				Action: createRole,
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code that will own the role. Falls back to context."),
					&cli.StringFlag{
						Name:     "name",
						Usage:    "User-defined role name.",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "permission",
						Usage:    "Permission granted by the role. Repeat for multiple permissions.",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "User-defined role description.",
					},
				},
			}),
			apicommands.Bind("GetMerchantRole", &cli.Command{
				Name:      "get",
				Usage:     "Get a custom role by ID.",
				Action:    getRole,
				ArgsUsage: "<role-id>",
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code that owns the role. Falls back to context."),
				},
			}),
			apicommands.Bind("UpdateMerchantRole", &cli.Command{
				Name:      "update",
				Usage:     "Update a custom role.",
				Action:    updateRole,
				ArgsUsage: "<role-id>",
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code that owns the role. Falls back to context."),
					&cli.StringFlag{
						Name:  "name",
						Usage: "Updated role name.",
					},
					&cli.StringSliceFlag{
						Name:  "permission",
						Usage: "Complete updated permission list. Repeat for multiple permissions.",
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Updated role description.",
					},
				},
			}),
			apicommands.Bind("DeleteMerchantRole", &cli.Command{
				Name:      "delete",
				Usage:     "Delete a custom role.",
				Action:    deleteRole,
				ArgsUsage: "<role-id>",
				Flags: []cli.Flag{
					merchantCodeFlag("Merchant code that owns the role. Falls back to context."),
				},
			}),
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

func createRole(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	permissions, err := rolePermissions(cmd)
	if err != nil {
		return err
	}

	body := sumup.RolesCreateParams{
		Name:        cmd.String("name"),
		Permissions: permissions,
	}
	if cmd.IsSet("description") {
		description := cmd.String("description")
		body.Description = &description
	}

	role, err := appCtx.Client.Roles.Create(ctx, merchantCode, body)
	if err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      role,
		SuccessMessage: "Role created",
		RenderHuman: func(w io.Writer) error {
			return renderRole(w, role)
		},
	})
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
		return display.PrintJSON(appCtx.Output, role)
	}
	return renderRole(appCtx.Output, role)
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
	changedCount := 0
	if cmd.IsSet("name") {
		name := cmd.String("name")
		body.Name = &name
		changedCount++
	}
	if cmd.IsSet("permission") {
		body.Permissions, err = rolePermissions(cmd)
		if err != nil {
			return err
		}
		changedCount++
	}
	if cmd.IsSet("description") {
		description := cmd.String("description")
		body.Description = &description
		changedCount++
	}
	if changedCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	role, err := appCtx.Client.Roles.Update(ctx, merchantCode, roleID, body)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      role,
		SuccessMessage: "Role updated",
		RenderHuman: func(w io.Writer) error {
			return renderRole(w, role)
		},
	})
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
		return display.PrintJSON(appCtx.Output, map[string]string{"status": "deleted"})
	}
	return message.Success(appCtx.StatusOutput, "Role deleted")
}

func rolePermissions(cmd *cli.Command) ([]string, error) {
	permissions := cmd.StringSlice("permission")
	for index, permission := range permissions {
		permissions[index] = strings.TrimSpace(permission)
		if permissions[index] == "" {
			return nil, fmt.Errorf("permission cannot be empty")
		}
	}
	if len(permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}
	return permissions, nil
}

func renderRole(w io.Writer, role *sumup.Role) error {
	if role == nil {
		return nil
	}

	permissions := "-"
	if len(role.Permissions) > 0 {
		permissions = strings.Join(role.Permissions, ", ")
	}
	return display.DataList(w, []attribute.KeyValue{
		attribute.ID(role.ID),
		attribute.Attribute("Name", attribute.Styled(role.Name)),
		attribute.OptionalString("Description", role.Description),
		attribute.Attribute("Permissions", attribute.Styled(permissions)),
		attribute.Attribute("Predefined", attribute.Styled(util.BoolLabel(&role.IsPredefined))),
	})
}

func roleDescription(role sumup.Role) string {
	if role.Description == nil || *role.Description == "" {
		return "-"
	}

	return *role.Description
}

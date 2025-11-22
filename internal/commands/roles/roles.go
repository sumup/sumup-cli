package roles

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
)

type role struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "roles",
		Usage: "Commands for managing roles.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List available roles.",
				Action: listRoles,
				Flags:  []cli.Flag{},
			},
		},
	}
}

func listRoles(_ context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	roles := defaultRoles()
	if appCtx.JSONOutput {
		return display.PrintJSON(roles)
	}

	rows := make([][]string, 0, len(roles))
	for _, role := range roles {
		rows = append(rows, []string{role.Name, role.DisplayName, role.Description})
	}

	display.RenderTable("Roles", []string{"Role", "Display Name", "Description"}, rows)
	return nil
}

func defaultRoles() []role {
	return []role{
		{
			Name:        "role_owner",
			DisplayName: "Owner",
			Description: "Full administrative access to the merchant account",
		},
		{
			Name:        "role_admin",
			DisplayName: "Admin",
			Description: "Administrative access with some restrictions",
		},
		{
			Name:        "role_employee",
			DisplayName: "Employee",
			Description: "Standard employee access for daily operations",
		},
		{
			Name:        "role_manager",
			DisplayName: "Manager",
			Description: "Management access with elevated permissions",
		},
		{
			Name:        "role_cashier",
			DisplayName: "Cashier",
			Description: "Limited access for point-of-sale operations",
		},
	}
}

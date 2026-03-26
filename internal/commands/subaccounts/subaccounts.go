package subaccounts

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
		Name:  "subaccounts",
		Usage: "Commands for the deprecated subaccounts API.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List subaccounts for the authenticated merchant.",
				Action: listSubaccounts,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "include-primary",
						Usage: "Include the primary user in results.",
					},
					&cli.StringFlag{
						Name:  "query",
						Usage: "Filter by email substring.",
					},
				},
			},
			{
				Name:      "get",
				Usage:     "Get a subaccount by numeric ID.",
				Action:    getSubaccount,
				ArgsUsage: "<operator-id>",
			},
			{
				Name:   "create",
				Usage:  "Create a subaccount.",
				Action: createSubaccount,
				Flags:  append(accountFlags(), permissionFlags("permission-")...),
			},
			{
				Name:      "update",
				Usage:     "Update a subaccount.",
				Action:    updateSubaccount,
				ArgsUsage: "<operator-id>",
				Flags: append([]cli.Flag{
					&cli.StringFlag{
						Name:  "username",
						Usage: "Updated login email.",
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "Updated password.",
					},
					&cli.StringFlag{
						Name:  "nickname",
						Usage: "Updated nickname.",
					},
					&cli.BoolFlag{
						Name:  "disabled",
						Usage: "Disable the subaccount.",
					},
					&cli.BoolFlag{
						Name:  "enabled",
						Usage: "Enable the subaccount.",
					},
				}, permissionFlags("permission-")...),
			},
		},
	}
}

func accountFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "username",
			Usage:    "Login email for the subaccount.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Usage:    "Password for the subaccount.",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "nickname",
			Usage: "Optional nickname.",
		},
	}
}

func permissionFlags(prefix string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{Name: prefix + "create-moto-payments", Usage: "Grant create MOTO payments permission."},
		&cli.BoolFlag{Name: prefix + "create-referral", Usage: "Grant create referral permission."},
		&cli.BoolFlag{Name: prefix + "full-transaction-history-view", Usage: "Grant full transaction history permission."},
		&cli.BoolFlag{Name: prefix + "refund-transactions", Usage: "Grant refund transactions permission."},
	}
}

func listSubaccounts(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	params := sumup.SubaccountsListSubAccountsParams{}
	if cmd.Bool("include-primary") {
		value := true
		params.IncludePrimary = &value
	}
	if query := cmd.String("query"); query != "" {
		params.Query = &query
	}

	response, err := appCtx.Client.Subaccounts.ListSubAccounts(ctx, params)
	if err != nil {
		return fmt.Errorf("list subaccounts: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(response)
	}

	rows := make([][]attribute.Value, 0, len(*response))
	for _, account := range *response {
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(account.ID),
			attribute.ValueOf(account.Username),
			attribute.ValueOf(nullableString(account.Nickname)),
			attribute.ValueOf(account.AccountType),
			attribute.ValueOf(boolLabel(account.Disabled)),
		})
	}

	display.RenderTable(
		"Subaccounts",
		[]string{"ID", "Username", "Nickname", "Type", "Disabled"},
		rows,
	)
	return nil
}

func getSubaccount(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	operatorID, err := parseOperatorID(cmd)
	if err != nil {
		return err
	}

	account, err := appCtx.Client.Subaccounts.CompatGetOperator(ctx, operatorID)
	if err != nil {
		return fmt.Errorf("get subaccount: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(account)
	}

	renderSubaccount(account)
	return nil
}

func createSubaccount(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	body := sumup.SubaccountsCreateSubAccountParams{
		Username: cmd.String("username"),
		Password: cmd.String("password"),
	}
	if nickname := cmd.String("nickname"); nickname != "" {
		body.Nickname = &nickname
	}
	if permissions := createPermissions(cmd); permissions != nil {
		body.Permissions = permissions
	}

	account, err := appCtx.Client.Subaccounts.CreateSubAccount(ctx, body)
	if err != nil {
		return fmt.Errorf("create subaccount: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(account)
	}

	message.Success("Subaccount created")
	renderSubaccount(account)
	return nil
}

func updateSubaccount(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	operatorID, err := parseOperatorID(cmd)
	if err != nil {
		return err
	}

	body := sumup.SubaccountsUpdateSubAccountParams{}
	changeCount := 0
	if username := cmd.String("username"); username != "" {
		body.Username = &username
		changeCount++
	}
	if password := cmd.String("password"); password != "" {
		body.Password = &password
		changeCount++
	}
	if nickname := cmd.String("nickname"); nickname != "" {
		body.Nickname = &nickname
		changeCount++
	}
	if cmd.Bool("disabled") && cmd.Bool("enabled") {
		return fmt.Errorf("use either --disabled or --enabled, not both")
	}
	if cmd.Bool("disabled") {
		value := true
		body.Disabled = &value
		changeCount++
	}
	if cmd.Bool("enabled") {
		value := false
		body.Disabled = &value
		changeCount++
	}
	if permissions := updatePermissions(cmd); permissions != nil {
		body.Permissions = permissions
		changeCount++
	}
	if changeCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	account, err := appCtx.Client.Subaccounts.UpdateSubAccount(ctx, operatorID, body)
	if err != nil {
		return fmt.Errorf("update subaccount: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(account)
	}

	message.Success("Subaccount updated")
	renderSubaccount(account)
	return nil
}

func parseOperatorID(cmd *cli.Command) (int32, error) {
	value, err := util.RequireSingleArg(cmd, "operator ID")
	if err != nil {
		return 0, err
	}
	var id int32
	if _, err := fmt.Sscanf(value, "%d", &id); err != nil {
		return 0, fmt.Errorf("invalid operator ID %q", value)
	}
	return id, nil
}

func createPermissions(cmd *cli.Command) *sumup.SubaccountsCreateSubAccountParamsPermissions {
	var permissions sumup.SubaccountsCreateSubAccountParamsPermissions
	changed := false
	if cmd.Bool("permission-create-moto-payments") {
		value := true
		permissions.CreateMotoPayments = &value
		changed = true
	}
	if cmd.Bool("permission-create-referral") {
		value := true
		permissions.CreateReferral = &value
		changed = true
	}
	if cmd.Bool("permission-full-transaction-history-view") {
		value := true
		permissions.FullTransactionHistoryView = &value
		changed = true
	}
	if cmd.Bool("permission-refund-transactions") {
		value := true
		permissions.RefundTransactions = &value
		changed = true
	}
	if !changed {
		return nil
	}
	return &permissions
}

func updatePermissions(cmd *cli.Command) *sumup.SubaccountsUpdateSubAccountParamsPermissions {
	var permissions sumup.SubaccountsUpdateSubAccountParamsPermissions
	changed := false
	if cmd.Bool("permission-create-moto-payments") {
		value := true
		permissions.CreateMotoPayments = &value
		changed = true
	}
	if cmd.Bool("permission-create-referral") {
		value := true
		permissions.CreateReferral = &value
		changed = true
	}
	if cmd.Bool("permission-full-transaction-history-view") {
		value := true
		permissions.FullTransactionHistoryView = &value
		changed = true
	}
	if cmd.Bool("permission-refund-transactions") {
		value := true
		permissions.RefundTransactions = &value
		changed = true
	}
	if !changed {
		return nil
	}
	return &permissions
}

func renderSubaccount(account *sumup.Operator) {
	if account == nil {
		return
	}

	display.DataList([]attribute.KeyValue{
		attribute.ID(account.ID),
		attribute.Attribute("Username", attribute.Styled(account.Username)),
		attribute.Attribute("Nickname", attribute.Styled(nullableString(account.Nickname))),
		attribute.Attribute("Type", attribute.Styled(account.AccountType)),
		attribute.Attribute("Disabled", attribute.Styled(boolLabel(account.Disabled))),
		attribute.Attribute("Permissions", attribute.Styled(renderPermissions(account.Permissions))),
		attribute.Attribute("Created At", attribute.Styled(account.CreatedAt.UTC().Format(time.RFC3339))),
		attribute.Attribute("Updated At", attribute.Styled(account.UpdatedAt.UTC().Format(time.RFC3339))),
	})
}

func renderPermissions(permissions sumup.Permissions) string {
	parts := make([]string, 0, 5)
	if permissions.Admin {
		parts = append(parts, "admin")
	}
	if permissions.CreateMotoPayments {
		parts = append(parts, "create_moto_payments")
	}
	if permissions.CreateReferral {
		parts = append(parts, "create_referral")
	}
	if permissions.FullTransactionHistoryView {
		parts = append(parts, "full_transaction_history_view")
	}
	if permissions.RefundTransactions {
		parts = append(parts, "refund_transactions")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

func nullableString(value interface{ Value() *string }) string {
	if value == nil {
		return "-"
	}
	current := value.Value()
	if current == nil || *current == "" {
		return "-"
	}
	return *current
}

func boolLabel(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

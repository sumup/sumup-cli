package members

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/secret"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "members",
		Usage: "Commands related to merchant members.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List members attached to a merchant resource.",
				Action: listMembers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose members should be listed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.IntFlag{
						Name:  "offset",
						Usage: "Offset of the first member to return.",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of members to return.",
					},
					&cli.StringFlag{
						Name:  "email",
						Usage: "Filter members by email prefix.",
					},
					&cli.StringFlag{
						Name:  "user-id",
						Usage: "Filter by a specific user ID.",
					},
					&cli.StringFlag{
						Name:  "status",
						Usage: "Filter by membership status (accepted, pending, expired, disabled, unknown).",
					},
					&cli.StringSliceFlag{
						Name:  "role",
						Usage: "Filter by roles (repeat flag to provide multiple roles).",
					},
					&cli.BoolFlag{
						Name:  "scroll",
						Usage: "Skip counting results to speed up pagination.",
					},
				},
			},
			{
				Name:   "create",
				Usage:  "Create a merchant member.",
				Action: createMember,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code for the new member. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:     "email",
						Usage:    "Email for the new member.",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "password",
						Usage:    "Password for the managed user.",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "role",
						Usage:    "Roles to assign to the member (repeat flag for multiple roles).",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "nickname",
						Usage: "Nickname for the member.",
					},
				},
			},
			{
				Name:      "get",
				Usage:     "Get a member from the merchant account.",
				Action:    getMember,
				ArgsUsage: "<member-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the member. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			},
			{
				Name:      "update",
				Usage:     "Update a member in the merchant account.",
				Action:    updateMember,
				ArgsUsage: "<member-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code that owns the member. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringSliceFlag{
						Name:  "role",
						Usage: "Roles to assign to the member (repeat flag for multiple roles).",
					},
					&cli.StringFlag{
						Name:  "nickname",
						Usage: "Nickname for managed users.",
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "Password for managed users.",
					},
				},
			},
			{
				Name:   "invite",
				Usage:  "Invite a user to become a member of the merchant account.",
				Action: inviteMember,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code to invite member to. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
					&cli.StringFlag{
						Name:     "email",
						Usage:    "Email of the user to invite.",
						Required: true,
					},
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a member from the merchant account.",
				Action:    deleteMember,
				ArgsUsage: "<member-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "merchant-code",
						Usage:   "Merchant code whose member should be removed. Falls back to context.",
						Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
					},
				},
			},
		},
	}
}

func listMembers(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	params := sumup.MembersListParams{}
	if cmd.IsSet("offset") {
		value := cmd.Int("offset")
		params.Offset = &value
	}
	if cmd.IsSet("limit") {
		value := cmd.Int("limit")
		params.Limit = &value
	}
	if value := cmd.String("email"); value != "" {
		params.Email = &value
	}
	if value := cmd.String("user-id"); value != "" {
		params.UserID = &value
	}
	if roles := cmd.StringSlice("role"); len(roles) > 0 {
		params.Roles = roles
	}
	if cmd.Bool("scroll") {
		value := true
		params.Scroll = &value
	}
	if value := cmd.String("status"); value != "" {
		status, err := parseMembershipStatus(value)
		if err != nil {
			return err
		}
		params.Status = &status
	}

	response, err := appCtx.Client.Members.List(ctx, merchantCode, params)
	if err != nil {
		return fmt.Errorf("list members: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response.Items)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for _, member := range response.Items {
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(member.ID),
			attribute.ValueOf(memberEmail(member)),
			attribute.ValueOf(memberRoles(member.Roles)),
			attribute.ValueOf(membershipStatusLabel(member.Status)),
			attribute.ValueOf(member.CreatedAt.UTC().Format(time.RFC3339)),
		})
	}

	display.RenderTable(
		appCtx.Output,
		"Members",
		[]string{"ID", "Email", "Roles", "Status", "Created At"},
		rows,
	)
	return nil
}

func createMember(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	roles := cmd.StringSlice("role")
	if len(roles) == 0 {
		return fmt.Errorf("at least one --role is required")
	}

	isManaged := true
	password := secret.New(cmd.String("password"))

	body := sumup.MembersCreateParams{
		Email:         cmd.String("email"),
		IsManagedUser: &isManaged,
		Password:      &password,
		Roles:         roles,
	}
	if nickname := cmd.String("nickname"); nickname != "" {
		body.Nickname = &nickname
	}

	response, err := appCtx.Client.Members.Create(ctx, merchantCode, body)
	if err != nil {
		return fmt.Errorf("create member: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	message.Success("Member created")
	display.DataList(appCtx.Output, []attribute.KeyValue{
		attribute.ID(response.ID),
	})
	return nil
}

func inviteMember(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	body := sumup.MembersCreateParams{
		Email: cmd.String("email"),
		Roles: []string{"role_employee"},
	}

	response, err := appCtx.Client.Members.Create(ctx, merchantCode, body)
	if err != nil {
		return fmt.Errorf("invite member: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	message.Success("Member invited")
	display.DataList(appCtx.Output, []attribute.KeyValue{
		attribute.ID(response.ID),
	})
	return nil
}

func getMember(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	memberID, err := util.RequireSingleArg(cmd, "member ID")
	if err != nil {
		return err
	}

	member, err := appCtx.Client.Members.Get(ctx, merchantCode, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, member)
	}

	renderMember(appCtx.Output, member)
	return nil
}

func updateMember(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	memberID, err := util.RequireSingleArg(cmd, "member ID")
	if err != nil {
		return err
	}

	body := sumup.MembersUpdateParams{}
	changeCount := 0

	if roles := cmd.StringSlice("role"); len(roles) > 0 {
		body.Roles = roles
		changeCount++
	}

	var user *sumup.MembersUpdateParamsUser
	if nickname := cmd.String("nickname"); nickname != "" {
		if user == nil {
			user = &sumup.MembersUpdateParamsUser{}
		}
		user.Nickname = &nickname
		changeCount++
	}
	if password := cmd.String("password"); password != "" {
		if user == nil {
			user = &sumup.MembersUpdateParamsUser{}
		}
		p := secret.New(password)
		user.Password = &p
		changeCount++
	}
	if user != nil {
		body.User = user
	}

	if changeCount == 0 {
		return fmt.Errorf("no update fields provided")
	}

	member, err := appCtx.Client.Members.Update(ctx, merchantCode, memberID, body)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, member)
	}

	message.Success("Member updated")
	renderMember(appCtx.Output, member)
	return nil
}

func deleteMember(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}

	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}

	memberID, err := util.RequireSingleArg(cmd, "member ID")
	if err != nil {
		return err
	}
	if err := appCtx.Client.Members.Delete(ctx, merchantCode, memberID); err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	message.Success("Member deleted")
	return nil
}

func parseMembershipStatus(value string) (sumup.MembershipStatus, error) {
	switch strings.ToLower(value) {
	case "accepted":
		return sumup.MembershipStatusAccepted, nil
	case "pending":
		return sumup.MembershipStatusPending, nil
	case "expired":
		return sumup.MembershipStatusExpired, nil
	case "disabled":
		return sumup.MembershipStatusDisabled, nil
	case "unknown":
		return sumup.MembershipStatusUnknown, nil
	default:
		return "", fmt.Errorf("unsupported status %q", value)
	}
}

func memberEmail(member sumup.Member) string {
	if member.User != nil && member.User.Email != "" {
		return member.User.Email
	}
	if member.Invite != nil && member.Invite.Email != "" {
		return member.Invite.Email
	}
	return "-"
}

func memberRoles(roles []string) string {
	if len(roles) == 0 {
		return "-"
	}
	return strings.Join(roles, ", ")
}

func membershipStatusLabel(status sumup.MembershipStatus) string {
	switch status {
	case sumup.MembershipStatusAccepted:
		return "Accepted"
	case sumup.MembershipStatusPending:
		return "Pending"
	case sumup.MembershipStatusExpired:
		return "Expired"
	case sumup.MembershipStatusDisabled:
		return "Disabled"
	default:
		return "Unknown"
	}
}

func renderMember(w io.Writer, member *sumup.Member) {
	if member == nil {
		return
	}

	var createdAt, updatedAt string
	createdAt = member.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt = member.UpdatedAt.UTC().Format(time.RFC3339)

	details := []attribute.KeyValue{
		attribute.ID(member.ID),
		attribute.Attribute("Email", attribute.Styled(memberEmail(*member))),
		attribute.Attribute("Roles", attribute.Styled(memberRoles(member.Roles))),
		attribute.Attribute("Status", attribute.Styled(membershipStatusLabel(member.Status))),
		attribute.Attribute("Nickname", attribute.Styled(memberNickname(member))),
		attribute.Attribute("Created At", attribute.Styled(createdAt)),
		attribute.Attribute("Updated At", attribute.Styled(updatedAt)),
	}

	display.DataList(w, details)
}

func memberNickname(member *sumup.Member) string {
	if member != nil && member.User != nil && member.User.Nickname != nil && *member.User.Nickname != "" {
		return *member.User.Nickname
	}
	return "-"
}

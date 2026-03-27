package memberships

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "memberships",
		Usage: "Commands related to sumup.",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List memberships for the authenticated user.",
				Action: listMemberships,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "offset",
						Usage: "Offset of the first membership to return.",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of memberships to return.",
					},
					&cli.StringFlag{
						Name:  "kind",
						Usage: "Filter memberships by resource kind.",
					},
					&cli.StringFlag{
						Name:  "status",
						Usage: "Filter memberships by status.",
					},
					&cli.StringFlag{
						Name:  "resource-type",
						Usage: "Filter memberships by the resource type.",
					},
					&cli.StringFlag{
						Name:  "resource-name",
						Usage: "Filter memberships by resource name.",
					},
					&cli.BoolFlag{
						Name:  "sandbox",
						Usage: "Filter memberships to sandbox resources only.",
					},
				},
			},
		},
	}
}

func listMemberships(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	params := sumup.MembershipsListParams{}
	if cmd.IsSet("offset") {
		value := cmd.Int("offset")
		params.Offset = &value
	}
	if cmd.IsSet("limit") {
		value := cmd.Int("limit")
		params.Limit = &value
	}
	if value := cmd.String("kind"); value != "" {
		kind := sumup.ResourceType(value)
		params.Kind = &kind
	}
	if value := cmd.String("status"); value != "" {
		status, err := parseMembershipStatus(value)
		if err != nil {
			return err
		}
		params.Status = &status
	}
	if value := cmd.String("resource-type"); value != "" {
		resourceType := sumup.ResourceType(value)
		params.ResourceType = &resourceType
	}
	if value := cmd.String("resource-name"); value != "" {
		params.ResourceName = &value
	}
	if cmd.Bool("sandbox") {
		value := true
		params.ResourceAttributesSandbox = &value
	}

	response, err := appCtx.Client.Memberships.List(ctx, params)
	if err != nil {
		return fmt.Errorf("list memberships: %w", err)
	}

	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}

	rows := make([][]attribute.Value, 0, len(response.Items))
	for _, membership := range response.Items {
		rows = append(rows, []attribute.Value{
			attribute.ValueOf(membership.ID),
			attribute.ValueOf(membership.Resource.Name),
			attribute.ValueOf(string(membership.Resource.Type)),
			attribute.ValueOf(memberRoles(membership.Roles)),
			attribute.ValueOf(membershipStatusLabel(membership.Status)),
			attribute.ValueOf(util.TimeOrDash(appCtx, &membership.CreatedAt)),
		})
	}

	return display.RenderTableWithOptions(appCtx.Output, []string{"ID", "Resource", "Type", "Roles", "Status", "Created At"}, rows, display.TableOptions{
		Title:             "Memberships",
		EmptyText:         "No items to display",
		IdentifierColumns: []int{0},
	})
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

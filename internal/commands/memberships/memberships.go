package memberships

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-go/memberships"
	"github.com/sumup/sumup-go/shared"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/display"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "memberships",
		Usage: "Commands related to memberships.",
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
	params := memberships.ListMembershipsParams{}
	if cmd.IsSet("offset") {
		value := cmd.Int("offset")
		params.Offset = &value
	}
	if cmd.IsSet("limit") {
		value := cmd.Int("limit")
		params.Limit = &value
	}
	if value := cmd.String("kind"); value != "" {
		kind := memberships.ResourceType(value)
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
		resourceType := memberships.ResourceType(value)
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
		return display.PrintJSON(response)
	}

	rows := make([][]string, 0, len(response.Items))
	for _, membership := range response.Items {
		rows = append(rows, []string{
			membership.ID,
			membership.Resource.Name,
			string(membership.Resource.Type),
			memberRoles(membership.Roles),
			membershipStatusLabel(membership.Status),
			membership.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	display.RenderTable("Memberships", []string{"ID", "Resource", "Type", "Roles", "Status", "Created At"}, rows)
	return nil
}

func parseMembershipStatus(value string) (shared.MembershipStatus, error) {
	switch strings.ToLower(value) {
	case "accepted":
		return shared.MembershipStatusAccepted, nil
	case "pending":
		return shared.MembershipStatusPending, nil
	case "expired":
		return shared.MembershipStatusExpired, nil
	case "disabled":
		return shared.MembershipStatusDisabled, nil
	case "unknown":
		return shared.MembershipStatusUnknown, nil
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

func membershipStatusLabel(status shared.MembershipStatus) string {
	switch status {
	case shared.MembershipStatusAccepted:
		return "Accepted"
	case shared.MembershipStatusPending:
		return "Pending"
	case shared.MembershipStatusExpired:
		return "Expired"
	case shared.MembershipStatusDisabled:
		return "Disabled"
	default:
		return "Unknown"
	}
}

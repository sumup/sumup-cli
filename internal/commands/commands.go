package commands

import (
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/commands/checkouts"
	"github.com/sumup/sumup-cli/internal/commands/context"
	"github.com/sumup/sumup-cli/internal/commands/customers"
	"github.com/sumup/sumup-cli/internal/commands/members"
	"github.com/sumup/sumup-cli/internal/commands/memberships"
	"github.com/sumup/sumup-cli/internal/commands/merchants"
	"github.com/sumup/sumup-cli/internal/commands/payouts"
	"github.com/sumup/sumup-cli/internal/commands/readers"
	"github.com/sumup/sumup-cli/internal/commands/receipts"
	"github.com/sumup/sumup-cli/internal/commands/roles"
	"github.com/sumup/sumup-cli/internal/commands/transactions"
)

// All returns the list of resource commands exposed by the CLI.
func All() []*cli.Command {
	return []*cli.Command{
		checkouts.NewCommand(),
		context.NewCommand(),
		customers.NewCommand(),
		members.NewCommand(),
		memberships.NewCommand(),
		merchants.NewCommand(),
		payouts.NewCommand(),
		readers.NewCommand(),
		receipts.NewCommand(),
		roles.NewCommand(),
		transactions.NewCommand(),
	}
}

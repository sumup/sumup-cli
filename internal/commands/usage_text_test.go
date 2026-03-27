package commands_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/commands/payouts"
	"github.com/sumup/sumup-cli/internal/commands/receipts"
	"github.com/sumup/sumup-cli/internal/commands/transactions"
)

func TestCommandUsageDoesNotAdvertisePlaceholders(t *testing.T) {
	tests := []struct {
		name  string
		usage string
	}{
		{name: "transactions", usage: transactions.NewCommand().Usage},
		{name: "payouts", usage: payouts.NewCommand().Usage},
		{name: "receipts", usage: receipts.NewCommand().Usage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.usage)
			assert.NotEqual(t, "Placeholder for the transactions API resource.", tt.usage)
			assert.NotEqual(t, "Placeholder for the payouts API resource.", tt.usage)
			assert.NotEqual(t, "Placeholder for the receipts API resource.", tt.usage)
		})
	}
}

package commands_test

import (
	"testing"

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
		if tt.usage == "" {
			t.Fatalf("%s usage is empty", tt.name)
		}
		if tt.usage == "Placeholder for the transactions API resource." ||
			tt.usage == "Placeholder for the payouts API resource." ||
			tt.usage == "Placeholder for the receipts API resource." {
			t.Fatalf("%s usage still contains placeholder text: %q", tt.name, tt.usage)
		}
	}
}

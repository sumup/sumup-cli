package transactions

import (
	"context"
	"testing"
	"time"

	sumup "github.com/sumup/sumup-go"
	"github.com/urfave/cli/v3"
)

func TestTransactionsListParamsFromCommand(t *testing.T) {
	t.Run("maps CLI flags into request params", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{
			"sumup",
			"--limit", "5",
			"--changes-since", "2026-03-26T12:00:00Z",
			"--payment-type", "CARD",
			"--payment-type", "",
			"--status", "SUCCESSFUL",
			"--user", "user@example.com",
		}, listTransactionsFlags())

		params, err := transactionsListParamsFromCommand(cmd)
		if err != nil {
			t.Fatalf("transactionsListParamsFromCommand() error = %v", err)
		}

		if params.Limit == nil || *params.Limit != 5 {
			t.Fatalf("Limit = %v, want 5", params.Limit)
		}
		if params.ChangesSince == nil || !params.ChangesSince.Equal(time.Date(2026, time.March, 26, 12, 0, 0, 0, time.UTC)) {
			t.Fatalf("ChangesSince = %v, want 2026-03-26T12:00:00Z", params.ChangesSince)
		}
		if len(params.PaymentTypes) != 1 || params.PaymentTypes[0] != sumup.PaymentType("CARD") {
			t.Fatalf("PaymentTypes = %v, want [CARD]", params.PaymentTypes)
		}
		if len(params.Statuses) != 1 || params.Statuses[0] != "SUCCESSFUL" {
			t.Fatalf("Statuses = %v, want [SUCCESSFUL]", params.Statuses)
		}
		if len(params.Users) != 1 || params.Users[0] != "user@example.com" {
			t.Fatalf("Users = %v, want [user@example.com]", params.Users)
		}
	})

	t.Run("rejects invalid RFC3339 timestamps", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "--changes-since", "not-a-time"}, listTransactionsFlags())

		if _, err := transactionsListParamsFromCommand(cmd); err == nil {
			t.Fatal("transactionsListParamsFromCommand() error = nil, want non-nil")
		}
	})
}

func TestTransactionLookupParamsFromCommand(t *testing.T) {
	t.Run("accepts exactly one lookup", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "--transaction-code", "TX123"}, getTransactionFlags())

		params, err := transactionLookupParamsFromCommand(cmd)
		if err != nil {
			t.Fatalf("transactionLookupParamsFromCommand() error = %v", err)
		}
		if params.TransactionCode == nil || *params.TransactionCode != "TX123" {
			t.Fatalf("TransactionCode = %v, want TX123", params.TransactionCode)
		}
	})

	t.Run("rejects missing lookups", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup"}, getTransactionFlags())

		if _, err := transactionLookupParamsFromCommand(cmd); err == nil {
			t.Fatal("transactionLookupParamsFromCommand() error = nil, want non-nil")
		}
	})

	t.Run("rejects multiple lookups", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "txn-1", "--transaction-code", "TX123"}, getTransactionFlags())

		if _, err := transactionLookupParamsFromCommand(cmd); err == nil {
			t.Fatal("transactionLookupParamsFromCommand() error = nil, want non-nil")
		}
	})
}

func TestPaymentTypesFromStrings(t *testing.T) {
	got := paymentTypesFromStrings([]string{"CARD", "", "CASH"})
	if len(got) != 2 || got[0] != sumup.PaymentType("CARD") || got[1] != sumup.PaymentType("CASH") {
		t.Fatalf("paymentTypesFromStrings() = %v, want [CARD CASH]", got)
	}
}

func listTransactionsFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{Name: "limit"},
		&cli.StringFlag{Name: "changes-since"},
		&cli.StringFlag{Name: "newest-ref"},
		&cli.StringFlag{Name: "newest-time"},
		&cli.StringFlag{Name: "oldest-ref"},
		&cli.StringFlag{Name: "oldest-time"},
		&cli.StringFlag{Name: "order"},
		&cli.StringSliceFlag{Name: "payment-type"},
		&cli.StringSliceFlag{Name: "status"},
		&cli.StringFlag{Name: "transaction-code"},
		&cli.StringSliceFlag{Name: "type"},
		&cli.StringSliceFlag{Name: "user"},
	}
}

func getTransactionFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "internal-id"},
		&cli.StringFlag{Name: "transaction-code"},
		&cli.StringFlag{Name: "foreign-transaction-id"},
		&cli.StringFlag{Name: "client-transaction-id"},
	}
}

func runCommandForTest(t *testing.T, args []string, flags []cli.Flag) *cli.Command {
	t.Helper()

	var captured *cli.Command
	cmd := &cli.Command{
		Name:  "sumup",
		Flags: flags,
		Action: func(_ context.Context, cmd *cli.Command) error {
			captured = cmd
			return nil
		},
	}

	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	return captured
}

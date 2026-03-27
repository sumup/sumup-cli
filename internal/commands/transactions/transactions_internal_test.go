package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)
		require.NotNil(t, params.Limit)
		assert.Equal(t, 5, *params.Limit)
		require.NotNil(t, params.ChangesSince)
		assert.True(t, params.ChangesSince.Equal(time.Date(2026, time.March, 26, 12, 0, 0, 0, time.UTC)))
		assert.Equal(t, []sumup.PaymentType{sumup.PaymentType("CARD")}, params.PaymentTypes)
		assert.Equal(t, []string{"SUCCESSFUL"}, params.Statuses)
		assert.Equal(t, []string{"user@example.com"}, params.Users)
	})

	t.Run("rejects invalid RFC3339 timestamps", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "--changes-since", "not-a-time"}, listTransactionsFlags())

		_, err := transactionsListParamsFromCommand(cmd)
		require.Error(t, err)
	})
}

func TestTransactionLookupParamsFromCommand(t *testing.T) {
	t.Run("accepts exactly one lookup", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "--transaction-code", "TX123"}, getTransactionFlags())

		params, err := transactionLookupParamsFromCommand(cmd)
		require.NoError(t, err)
		require.NotNil(t, params.TransactionCode)
		assert.Equal(t, "TX123", *params.TransactionCode)
	})

	t.Run("rejects missing lookups", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup"}, getTransactionFlags())

		_, err := transactionLookupParamsFromCommand(cmd)
		require.Error(t, err)
	})

	t.Run("rejects multiple lookups", func(t *testing.T) {
		cmd := runCommandForTest(t, []string{"sumup", "txn-1", "--transaction-code", "TX123"}, getTransactionFlags())

		_, err := transactionLookupParamsFromCommand(cmd)
		require.Error(t, err)
	})
}

func TestPaymentTypesFromStrings(t *testing.T) {
	t.Run("filters empty values and preserves valid payment types", func(t *testing.T) {
		got := paymentTypesFromStrings([]string{"CARD", "", "CASH"})
		assert.Equal(t, []sumup.PaymentType{sumup.PaymentType("CARD"), sumup.PaymentType("CASH")}, got)
	})
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
		require.NoError(t, err)
	}

	return captured
}

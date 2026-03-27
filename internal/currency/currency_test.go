package currency_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/currency"
)

func TestParseMajorUnits32(t *testing.T) {
	t.Run("parses decimal strings without using float flags", func(t *testing.T) {
		got, err := currency.ParseMajorUnits32("10.10")

		require.NoError(t, err)
		assert.Equal(t, float32(10.10), got)
	})

	t.Run("rejects invalid decimal strings", func(t *testing.T) {
		_, err := currency.ParseMajorUnits32("12,34")
		require.Error(t, err)
	})
}

func TestParseMajorUnitsForCurrency32(t *testing.T) {
	t.Run("accepts valid scale for currency", func(t *testing.T) {
		got, err := currency.ParseMajorUnitsForCurrency32("29.99", sumup.CurrencyEUR)

		require.NoError(t, err)
		assert.Equal(t, float32(29.99), got)
	})

	t.Run("rejects too many decimal places for currency", func(t *testing.T) {
		_, err := currency.ParseMajorUnitsForCurrency32("29.999", sumup.CurrencyEUR)
		require.Error(t, err)
	})
}

func TestParseMajorUnitsForCurrency64(t *testing.T) {
	t.Run("rejects fractional minor units for zero-decimal currencies", func(t *testing.T) {
		_, err := currency.ParseMajorUnitsForCurrency64("10.5", sumup.CurrencyCLP)
		require.Error(t, err)
	})
}

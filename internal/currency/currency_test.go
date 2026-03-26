package currency_test

import (
	"testing"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/currency"
)

func TestParseMajorUnits32(t *testing.T) {
	t.Run("parses decimal strings without using float flags", func(t *testing.T) {
		got, err := currency.ParseMajorUnits32("10.10")
		if err != nil {
			t.Fatalf("ParseMajorUnits32() error = %v", err)
		}

		if got != float32(10.10) {
			t.Fatalf("ParseMajorUnits32() = %v, want %v", got, float32(10.10))
		}
	})

	t.Run("rejects invalid decimal strings", func(t *testing.T) {
		if _, err := currency.ParseMajorUnits32("12,34"); err == nil {
			t.Fatal("ParseMajorUnits32() error = nil, want non-nil")
		}
	})
}

func TestParseMajorUnitsForCurrency32(t *testing.T) {
	t.Run("accepts valid scale for currency", func(t *testing.T) {
		got, err := currency.ParseMajorUnitsForCurrency32("29.99", sumup.CurrencyEUR)
		if err != nil {
			t.Fatalf("ParseMajorUnitsForCurrency32() error = %v", err)
		}

		if got != float32(29.99) {
			t.Fatalf("ParseMajorUnitsForCurrency32() = %v, want %v", got, float32(29.99))
		}
	})

	t.Run("rejects too many decimal places for currency", func(t *testing.T) {
		if _, err := currency.ParseMajorUnitsForCurrency32("29.999", sumup.CurrencyEUR); err == nil {
			t.Fatal("ParseMajorUnitsForCurrency32() error = nil, want non-nil")
		}
	})
}

func TestParseMajorUnitsForCurrency64(t *testing.T) {
	t.Run("rejects fractional minor units for zero-decimal currencies", func(t *testing.T) {
		if _, err := currency.ParseMajorUnitsForCurrency64("10.5", sumup.CurrencyCLP); err == nil {
			t.Fatal("ParseMajorUnitsForCurrency64() error = nil, want non-nil")
		}
	})
}

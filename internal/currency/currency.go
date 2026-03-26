package currency

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	sumup "github.com/sumup/sumup-go"
)

type symbolPosition int

const (
	positionBefore symbolPosition = iota
	positionAfter
	positionAfterNoSpace
)

type currencyInfo struct {
	symbol   string
	decimals int32
	position symbolPosition
}

var infoByCurrency = map[sumup.Currency]currencyInfo{
	sumup.CurrencyBGN: {symbol: "лв", decimals: 2, position: positionAfter},
	sumup.CurrencyBRL: {symbol: "R$", decimals: 2, position: positionBefore},
	sumup.CurrencyCHF: {symbol: "CHF ", decimals: 2, position: positionBefore},
	sumup.CurrencyCLP: {symbol: "$", decimals: 0, position: positionBefore},
	sumup.CurrencyCZK: {symbol: "Kč", decimals: 2, position: positionAfter},
	sumup.CurrencyDKK: {symbol: "kr", decimals: 2, position: positionAfter},
	sumup.CurrencyEUR: {symbol: "€", decimals: 2, position: positionAfterNoSpace},
	sumup.CurrencyGBP: {symbol: "£", decimals: 2, position: positionBefore},
	sumup.CurrencyHRK: {symbol: "kn", decimals: 2, position: positionAfter},
	sumup.CurrencyHUF: {symbol: "Ft", decimals: 0, position: positionAfter},
	sumup.CurrencyNOK: {symbol: "kr", decimals: 2, position: positionAfter},
	sumup.CurrencyPLN: {symbol: "zł", decimals: 2, position: positionAfter},
	sumup.CurrencyRON: {symbol: "lei", decimals: 2, position: positionAfter},
	sumup.CurrencySEK: {symbol: "kr", decimals: 2, position: positionAfter},
	sumup.CurrencyUSD: {symbol: "$", decimals: 2, position: positionBefore},
}

var codeToCurrency = map[string]sumup.Currency{
	"BGN": sumup.CurrencyBGN,
	"BRL": sumup.CurrencyBRL,
	"CHF": sumup.CurrencyCHF,
	"CLP": sumup.CurrencyCLP,
	"CZK": sumup.CurrencyCZK,
	"DKK": sumup.CurrencyDKK,
	"EUR": sumup.CurrencyEUR,
	"GBP": sumup.CurrencyGBP,
	"HRK": sumup.CurrencyHRK,
	"HUF": sumup.CurrencyHUF,
	"NOK": sumup.CurrencyNOK,
	"PLN": sumup.CurrencyPLN,
	"RON": sumup.CurrencyRON,
	"SEK": sumup.CurrencySEK,
	"USD": sumup.CurrencyUSD,
}

// Format renders an amount with a currency symbol.
func Format(amount float64, currency sumup.Currency) string {
	info, ok := infoByCurrency[currency]
	if !ok {
		return fmt.Sprintf("%.*f %s", 2, amount, string(currency))
	}
	value := decimal.NewFromFloat(amount).StringFixed(info.decimals)
	switch info.position {
	case positionBefore:
		return info.symbol + value
	case positionAfter:
		return value + " " + info.symbol
	default:
		return value + info.symbol
	}
}

// FormatPointers renders optional amount and currency pointers.
func FormatPointers(amount *float32, currency *sumup.Currency) string {
	if amount == nil {
		return "-"
	}
	if currency == nil {
		return fmt.Sprintf("%.2f", *amount)
	}
	return Format(float64(*amount), *currency)
}

// Parse converts a string into a SumUp currency value.
func Parse(value string) (sumup.Currency, error) {
	normalized := strings.TrimSpace(strings.ToUpper(value))
	currency, ok := codeToCurrency[normalized]
	if !ok {
		return "", fmt.Errorf(
			"unsupported currency %q. Supported values: %s",
			value,
			strings.Join(Supported(), ", "),
		)
	}
	return currency, nil
}

// Supported returns the currency codes understood by the CLI.
func Supported() []string {
	return []string{
		"BGN",
		"BRL",
		"CHF",
		"CLP",
		"CZK",
		"DKK",
		"EUR",
		"GBP",
		"HRK",
		"HUF",
		"NOK",
		"PLN",
		"RON",
		"SEK",
		"USD",
	}
}

// Code returns the ISO code string representation of the currency.
func Code(currency sumup.Currency) string {
	return string(currency)
}

// ToMinorUnits converts a decimal-string amount into its minor units.
func ToMinorUnits(amount string, minorUnit int32) (int64, error) {
	decimalAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", amount, err)
	}
	factor := decimal.New(1, 0).Shift(minorUnit)
	scaled := decimalAmount.Mul(factor).Round(0)
	value := scaled.IntPart()
	return value, nil
}

// ParseMajorUnits32 converts a decimal amount string into float32 major units.
func ParseMajorUnits32(amount string) (float32, error) {
	normalized, err := normalizedAmountString(amount)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(normalized, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	return float32(value), nil
}

// ParseMajorUnitsForCurrency32 converts a decimal amount string into float32 major units
// after validating the scale for the given currency.
func ParseMajorUnitsForCurrency32(amount string, currency sumup.Currency) (float32, error) {
	normalized, err := normalizedAmountStringForCurrency(amount, currency)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(normalized, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	return float32(value), nil
}

// ParseMajorUnitsForCurrency64 converts a decimal amount string into float64 major units
// after validating the scale for the given currency.
func ParseMajorUnitsForCurrency64(amount string, currency sumup.Currency) (float64, error) {
	normalized, err := normalizedAmountStringForCurrency(amount, currency)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	return value, nil
}

func normalizedAmountString(amount string) (string, error) {
	decimalAmount, err := decimal.NewFromString(strings.TrimSpace(amount))
	if err != nil {
		return "", fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	return decimalAmount.String(), nil
}

func normalizedAmountStringForCurrency(amount string, currency sumup.Currency) (string, error) {
	decimalAmount, err := decimal.NewFromString(strings.TrimSpace(amount))
	if err != nil {
		return "", fmt.Errorf("invalid amount %q: %w", amount, err)
	}

	info, ok := infoByCurrency[currency]
	if ok && -decimalAmount.Exponent() > info.decimals {
		return "", fmt.Errorf("invalid amount %q for %s: supports at most %d decimal places", amount, currency, info.decimals)
	}

	return decimalAmount.StringFixedBank(max(-decimalAmount.Exponent(), 0)), nil
}

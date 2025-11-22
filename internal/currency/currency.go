package currency

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/sumup/sumup-go/shared"
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

var infoByCurrency = map[shared.Currency]currencyInfo{
	shared.CurrencyBGN: {symbol: "лв", decimals: 2, position: positionAfter},
	shared.CurrencyBRL: {symbol: "R$", decimals: 2, position: positionBefore},
	shared.CurrencyCHF: {symbol: "CHF ", decimals: 2, position: positionBefore},
	shared.CurrencyCLP: {symbol: "$", decimals: 0, position: positionBefore},
	shared.CurrencyCZK: {symbol: "Kč", decimals: 2, position: positionAfter},
	shared.CurrencyDKK: {symbol: "kr", decimals: 2, position: positionAfter},
	shared.CurrencyEUR: {symbol: "€", decimals: 2, position: positionAfterNoSpace},
	shared.CurrencyGBP: {symbol: "£", decimals: 2, position: positionBefore},
	shared.CurrencyHRK: {symbol: "kn", decimals: 2, position: positionAfter},
	shared.CurrencyHUF: {symbol: "Ft", decimals: 0, position: positionAfter},
	shared.CurrencyNOK: {symbol: "kr", decimals: 2, position: positionAfter},
	shared.CurrencyPLN: {symbol: "zł", decimals: 2, position: positionAfter},
	shared.CurrencyRON: {symbol: "lei", decimals: 2, position: positionAfter},
	shared.CurrencySEK: {symbol: "kr", decimals: 2, position: positionAfter},
	shared.CurrencyUSD: {symbol: "$", decimals: 2, position: positionBefore},
}

var codeToCurrency = map[string]shared.Currency{
	"BGN": shared.CurrencyBGN,
	"BRL": shared.CurrencyBRL,
	"CHF": shared.CurrencyCHF,
	"CLP": shared.CurrencyCLP,
	"CZK": shared.CurrencyCZK,
	"DKK": shared.CurrencyDKK,
	"EUR": shared.CurrencyEUR,
	"GBP": shared.CurrencyGBP,
	"HRK": shared.CurrencyHRK,
	"HUF": shared.CurrencyHUF,
	"NOK": shared.CurrencyNOK,
	"PLN": shared.CurrencyPLN,
	"RON": shared.CurrencyRON,
	"SEK": shared.CurrencySEK,
	"USD": shared.CurrencyUSD,
}

// Format renders an amount with a currency symbol.
func Format(amount float64, currency shared.Currency) string {
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
func FormatPointers(amount *float32, currency *shared.Currency) string {
	if amount == nil {
		return "-"
	}
	if currency == nil {
		return fmt.Sprintf("%.2f", *amount)
	}
	return Format(float64(*amount), *currency)
}

// Parse converts a string into a SumUp currency value.
func Parse(value string) (shared.Currency, error) {
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
func Code(currency shared.Currency) string {
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

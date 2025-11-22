package message

import "fmt"

const (
	resetColor  = "\033[0m"
	greenColor  = "\033[32m"
	yellowColor = "\033[33m"
	blueColor   = "\033[34m"
	redColor    = "\033[31m"
)

const (
	successSymbol = "✓"
	warnSymbol    = "⚠"
	notifySymbol  = "ℹ"
	errorSymbol   = "✖"
)

// Success prints a green success message prefixed with a check mark.
func Success(format string, args ...any) {
	printColored(greenColor, successSymbol, format, args...)
}

// Warn prints a yellow warning message prefixed with a caution sign.
func Warn(format string, args ...any) {
	printColored(yellowColor, warnSymbol, format, args...)
}

// Notify prints a blue informational message prefixed with an info sign.
func Notify(format string, args ...any) {
	printColored(blueColor, notifySymbol, format, args...)
}

// Error prints a red error message prefixed with a cross.
func Error(format string, args ...any) {
	printColored(redColor, errorSymbol, format, args...)
}

func printColored(colorCode, symbol, format string, args ...any) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	fmt.Printf("%s%s %s%s\n", colorCode, symbol, message, resetColor)
}

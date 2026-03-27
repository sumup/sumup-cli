package message

import (
	"fmt"
	"io"
	"os"

	"github.com/sumup/sumup-cli/internal/outpututil"
)

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
func Success(w io.Writer, format string, args ...any) error {
	return printColored(w, greenColor, successSymbol, format, args...)
}

// Warn prints a yellow warning message prefixed with a caution sign.
func Warn(w io.Writer, format string, args ...any) error {
	return printColored(w, yellowColor, warnSymbol, format, args...)
}

// Notify prints a blue informational message prefixed with an info sign.
func Notify(w io.Writer, format string, args ...any) error {
	return printColored(w, blueColor, notifySymbol, format, args...)
}

// Error prints a red error message prefixed with a cross.
func Error(w io.Writer, format string, args ...any) error {
	return printColored(w, redColor, errorSymbol, format, args...)
}

func printColored(w io.Writer, colorCode, symbol, format string, args ...any) error {
	out := outpututil.WriterOrDefault(w, os.Stdout)
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	if outpututil.SupportsColor(out) {
		return outpututil.Fprintf(out, os.Stdout, "%s%s %s%s\n", colorCode, symbol, message, resetColor)
	}

	return outpututil.Fprintf(out, os.Stdout, "%s %s\n", symbol, message)
}

package message

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
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
func Success(w io.Writer, format string, args ...any) {
	printColored(w, greenColor, successSymbol, format, args...)
}

// Warn prints a yellow warning message prefixed with a caution sign.
func Warn(w io.Writer, format string, args ...any) {
	printColored(w, yellowColor, warnSymbol, format, args...)
}

// Notify prints a blue informational message prefixed with an info sign.
func Notify(w io.Writer, format string, args ...any) {
	printColored(w, blueColor, notifySymbol, format, args...)
}

// Error prints a red error message prefixed with a cross.
func Error(w io.Writer, format string, args ...any) {
	printColored(w, redColor, errorSymbol, format, args...)
}

func printColored(w io.Writer, colorCode, symbol, format string, args ...any) {
	out := writerOrDefault(w, os.Stdout)
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	if supportsColor(out) {
		_, _ = fmt.Fprintf(out, "%s%s %s%s\n", colorCode, symbol, message, resetColor)
		return
	}

	_, _ = fmt.Fprintf(out, "%s %s\n", symbol, message)
}

func writerOrDefault(w io.Writer, fallback io.Writer) io.Writer {
	if w == nil {
		return fallback
	}
	return w
}

func supportsColor(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
}

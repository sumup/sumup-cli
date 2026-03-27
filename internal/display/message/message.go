package message

import (
	"fmt"
	"io"
	"os"
	"strings"

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
	out := writerOrDefault(w, os.Stdout)
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	if supportsColor(out) {
		_, err := fmt.Fprintf(out, "%s%s %s%s\n", colorCode, symbol, message, resetColor)
		return err
	}

	_, err := fmt.Fprintf(out, "%s %s\n", symbol, message)
	return err
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

	return shouldUseColor(term.IsTerminal(int(file.Fd())), os.Getenv("TERM"), os.Getenv("NO_COLOR"))
}

func shouldUseColor(isTerminal bool, termEnv, noColorEnv string) bool {
	if !isTerminal {
		return false
	}
	if strings.TrimSpace(noColorEnv) != "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(termEnv), "dumb") {
		return false
	}
	return true
}

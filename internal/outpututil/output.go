package outpututil

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func WriterOrDefault(w, fallback io.Writer) io.Writer {
	if w == nil {
		return fallback
	}
	return w
}

func Fprintln(w, fallback io.Writer, args ...any) error {
	_, err := fmt.Fprintln(WriterOrDefault(w, fallback), args...)
	return err
}

func Fprintf(w, fallback io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(WriterOrDefault(w, fallback), format, args...)
	return err
}

func SupportsColor(w io.Writer) bool {
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

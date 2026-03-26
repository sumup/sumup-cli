package display

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PrintJSON renders the value as pretty JSON.
func PrintJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	writeln(w, string(data))
	return nil
}

func writerOrStdout(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

func writeln(w io.Writer, args ...any) {
	_, _ = fmt.Fprintln(writerOrStdout(w), args...)
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(writerOrStdout(w), format, args...)
}

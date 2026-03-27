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
	return writeln(w, string(data))
}

func writerOrStdout(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

func writeln(w io.Writer, args ...any) error {
	_, err := fmt.Fprintln(writerOrStdout(w), args...)
	return err
}

func writef(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(writerOrStdout(w), format, args...)
	return err
}

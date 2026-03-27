package display

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sumup/sumup-cli/internal/outpututil"
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
	return outpututil.WriterOrDefault(w, os.Stdout)
}

func writeln(w io.Writer, args ...any) error {
	return outpututil.Fprintln(w, os.Stdout, args...)
}

func writef(w io.Writer, format string, args ...any) error {
	return outpututil.Fprintf(w, os.Stdout, format, args...)
}

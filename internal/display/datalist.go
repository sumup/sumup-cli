package display

import (
	"io"
	"strings"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

// DataList renders key/value pairs as "Key: Value" rows where keys are bold.
func DataList(w io.Writer, pairs []attribute.KeyValue) error {
	if len(pairs) == 0 {
		return nil
	}

	for _, pair := range pairs {
		if strings.TrimSpace(pair.Key.Text) == "" {
			continue
		}
		if err := writef(w, "%s: %s\n", pair.Key.String(), pair.Value.String()); err != nil {
			return err
		}
	}
	return nil
}

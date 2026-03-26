package display

import (
	"fmt"
	"io"
	"strings"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

// DataList renders key/value pairs as "Key: Value" rows where keys are bold.
func DataList(w io.Writer, pairs []attribute.KeyValue) {
	if len(pairs) == 0 {
		return
	}

	for _, pair := range pairs {
		if strings.TrimSpace(pair.Key.Text) == "" {
			continue
		}
		writef(w, "%s: %s\n", pair.Key.String(), pair.Value.String())
	}
}

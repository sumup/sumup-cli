package display

import (
	"fmt"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

// DataList renders key/value pairs as "Key: Value" rows where keys are bold.
func DataList(pairs []attribute.KeyValue) {
	if len(pairs) == 0 {
		return
	}

	for _, pair := range pairs {
		if pair.Key.V == "" {
			continue
		}
		fmt.Printf("%s: %s\n", pair.Key.String(), pair.Value.String())
	}
}

package display

import (
	"encoding/json"
	"fmt"
)

// PrintJSON renders the value as pretty JSON.
func PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

package display

import (
	"io"

	"github.com/sumup/sumup-cli/internal/display/attribute"
	"github.com/sumup/sumup-cli/internal/display/message"
)

// MutationResult describes the shared output flow for mutating commands.
type MutationResult struct {
	JSONValue      any
	SuccessMessage string
	Details        []attribute.KeyValue
	RenderHuman    func(io.Writer) error
}

// RenderMutation renders a mutating command result consistently for JSON and human output.
func RenderMutation(output, statusOutput io.Writer, jsonOutput bool, result MutationResult) error {
	if jsonOutput {
		return PrintJSON(output, result.JSONValue)
	}

	if result.SuccessMessage != "" {
		if err := message.Success(statusOutput, "%s", result.SuccessMessage); err != nil {
			return err
		}
	}

	if result.RenderHuman != nil {
		return result.RenderHuman(output)
	}

	return DataList(output, result.Details)
}

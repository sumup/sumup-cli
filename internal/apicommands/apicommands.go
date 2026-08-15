// Package apicommands connects CLI commands to the OpenAPI operations exposed
// by the pinned SumUp Go SDK.
package apicommands

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

//go:generate go run ../cmd/generate-operations -out catalog.gen.go

const operationIDMetadataKey = "sumup.openapi.operation-id"

// Parameter describes an OpenAPI operation parameter.
type Parameter struct {
	Name        string
	Location    string
	Description string
	Type        string
	Format      string
	Required    bool
}

// RequestBody describes the JSON request body accepted by an operation.
type RequestBody struct {
	Schema   string
	Required bool
}

// Operation describes one SDK method generated from an OpenAPI operation.
type Operation struct {
	ID          string
	Client      string
	SDKMethod   string
	HTTPMethod  string
	Path        string
	Summary     string
	Description string
	Parameters  []Parameter
	RequestBody *RequestBody
}

// Lookup returns the generated operation with the given OpenAPI operation ID.
func Lookup(operationID string) (Operation, bool) {
	for _, operation := range Operations {
		if operation.ID == operationID {
			return operation, true
		}
	}

	return Operation{}, false
}

// Bind records which OpenAPI operation a CLI command exposes and returns the
// command so the binding can live next to the command definition.
func Bind(operationID string, command *cli.Command) *cli.Command {
	if command == nil {
		panic("cannot bind an OpenAPI operation to a nil command")
	}
	if _, ok := Lookup(operationID); !ok {
		panic(fmt.Sprintf("unknown OpenAPI operation %q", operationID))
	}
	if command.Metadata == nil {
		command.Metadata = make(map[string]any)
	}
	command.Metadata[operationIDMetadataKey] = operationID

	return command
}

// OperationID returns the OpenAPI operation ID bound to a CLI command.
func OperationID(command *cli.Command) (string, bool) {
	if command == nil || command.Metadata == nil {
		return "", false
	}

	operationID, ok := command.Metadata[operationIDMetadataKey].(string)
	return operationID, ok && operationID != ""
}

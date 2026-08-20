// Package codesamples generates the versioned CLI sample catalog consumed by
// the SumUp developer portal.
package codesamples

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/commands"
)

const (
	catalogSchemaVersion = 1
	cliModule            = "github.com/sumup/sumup-cli"
	cliLanguage          = "bash"
)

// Catalog is the versioned JSON contract consumed by documentation sites.
type Catalog struct {
	SchemaVersion  int      `json:"schemaVersion"`
	Language       string   `json:"language"`
	SDK            SDK      `json:"sdk"`
	OpenAPIVersion string   `json:"openAPIVersion"`
	Samples        []Sample `json:"samples"`
}

// SDK identifies the CLI package used by every generated sample. The field
// name is retained for compatibility with the shared portal catalog schema.
type SDK struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

// Sample is one complete CLI invocation for an OpenAPI operation.
type Sample struct {
	ID          string `json:"id"`
	OperationID string `json:"operationId"`
	Example     string `json:"example,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	HTTPMethod  string `json:"httpMethod"`
	Path        string `json:"path"`
	Source      string `json:"sample"`
}

type boundCommand struct {
	path    string
	command *cli.Command
}

type commandInvocation struct {
	path      []string
	arguments []string
	flags     []sampleFlag
}

type sampleFlag struct {
	name    string
	value   string
	boolean bool
}

var argumentPattern = regexp.MustCompile(`[<\[]([a-z0-9-]+)[>\]]`)

// optionalSampleFlags adds a representative field where an otherwise valid
// command would not show what a create or update request changes.
var optionalSampleFlags = map[string]map[string]string{
	"checkouts update": {
		"description": "Updated order",
	},
	"customers create": {
		"email": "customer@example.com",
	},
	"customers update": {
		"email": "updated-customer@example.com",
	},
	"members update": {
		"role": "role_employee",
	},
	"roles update": {
		"name": "Payment reviewer",
	},
}

var flagSampleValues = map[string]string{
	"amount":                "10.00",
	"client-transaction-id": "19e12390-72cf-4f9f-80b5-b0c8a67fa43f",
	"context":               "example.com",
	"currency":              "EUR",
	"email":                 "member@example.com",
	"end-date":              "2026-01-31",
	"merchant-code":         "$SUMUP_MERCHANT_CODE",
	"name":                  "Example",
	"pairing-code":          "4WLFDSBF",
	"password":              "$MEMBER_PASSWORD",
	"payment-type":          "card",
	"permission":            "members_access",
	"reference":             "order-123",
	"role":                  "role_employee",
	"start-date":            "2026-01-01",
	"target":                "https://apple-pay-gateway-cert.apple.com/paymentservices/startSession",
}

var argumentSampleValues = map[string]string{
	"checkout-id":    "$CHECKOUT_ID",
	"customer-id":    "$CUSTOMER_ID",
	"member-id":      "$MEMBER_ID",
	"person-id":      "$PERSON_ID",
	"reader-id":      "$READER_ID",
	"role-id":        "$ROLE_ID",
	"token":          "$PAYMENT_INSTRUMENT_TOKEN",
	"transaction-id": "$TRANSACTION_ID",
}

// Generate builds a deterministic sample catalog for the current CLI command
// tree and generated OpenAPI operation catalog.
func Generate(cliVersion string) (*Catalog, error) {
	cliVersion = strings.TrimSpace(cliVersion)
	if cliVersion == "" {
		return nil, errors.New("cli version is required")
	}

	commandsByOperation := boundCommandsByOperation(commands.All())
	samples := make([]Sample, 0, len(apicommands.Operations))
	for _, operation := range apicommands.Operations {
		command, err := commandForOperation(operation.ID, commandsByOperation[operation.ID])
		if err != nil {
			return nil, err
		}
		source, err := renderCommand(command)
		if err != nil {
			return nil, fmt.Errorf("generate sample for %q: %w", operation.ID, err)
		}
		samples = append(samples, Sample{
			ID:          operation.ID,
			OperationID: operation.ID,
			Summary:     operation.Summary,
			Description: operation.Description,
			HTTPMethod:  operation.HTTPMethod,
			Path:        operation.Path,
			Source:      source,
		})
	}

	slices.SortFunc(samples, func(a, b Sample) int {
		return strings.Compare(a.ID, b.ID)
	})

	return &Catalog{
		SchemaVersion: catalogSchemaVersion,
		Language:      cliLanguage,
		SDK: SDK{
			Module:  cliModule,
			Version: cliVersion,
		},
		OpenAPIVersion: apicommands.CatalogOpenAPIVersion,
		Samples:        samples,
	}, nil
}

func boundCommandsByOperation(resourceCommands []*cli.Command) map[string][]boundCommand {
	result := make(map[string][]boundCommand)
	var walk func([]string, *cli.Command)
	walk = func(parents []string, command *cli.Command) {
		path := append(parents, command.Name)
		if len(command.Commands) == 0 {
			if operationID, ok := apicommands.OperationID(command); ok {
				result[operationID] = append(result[operationID], boundCommand{
					path:    strings.Join(path, " "),
					command: command,
				})
			}
			return
		}
		for _, child := range command.Commands {
			walk(path, child)
		}
	}
	for _, command := range resourceCommands {
		walk(nil, command)
	}
	return result
}

func commandForOperation(operationID string, candidates []boundCommand) (boundCommand, error) {
	switch len(candidates) {
	case 0:
		return boundCommand{}, fmt.Errorf("OpenAPI operation %q has no CLI command", operationID)
	case 1:
		return candidates[0], nil
	}

	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		paths = append(paths, candidate.path)
	}
	slices.Sort(paths)
	return boundCommand{}, fmt.Errorf("OpenAPI operation %q has multiple CLI commands (%s)", operationID, strings.Join(paths, ", "))
}

func renderCommand(bound boundCommand) (string, error) {
	invocation, err := buildInvocation(bound)
	if err != nil {
		return "", err
	}
	return invocation.source(), nil
}

func buildInvocation(bound boundCommand) (commandInvocation, error) {
	invocation := commandInvocation{path: strings.Fields(bound.path)}
	for _, match := range argumentPattern.FindAllStringSubmatch(bound.command.ArgsUsage, -1) {
		value, ok := argumentSampleValues[match[1]]
		if !ok {
			return commandInvocation{}, fmt.Errorf("no sample value for argument %q", match[1])
		}
		invocation.arguments = append(invocation.arguments, value)
	}

	for _, flag := range bound.command.Flags {
		name := flag.Names()[0]
		value, optional := optionalSampleFlags[bound.path][name]
		required := false
		if requirement, ok := flag.(interface{ IsRequired() bool }); ok {
			required = requirement.IsRequired()
		}
		if !required && name != "merchant-code" && !optional {
			continue
		}

		if boolean, ok := flag.(interface{ IsBoolFlag() bool }); ok && boolean.IsBoolFlag() {
			invocation.flags = append(invocation.flags, sampleFlag{name: name, boolean: true})
			continue
		}
		if !optional {
			var ok bool
			value, ok = flagSampleValues[name]
			if !ok {
				return commandInvocation{}, fmt.Errorf("no sample value for flag --%s", name)
			}
		}
		invocation.flags = append(invocation.flags, sampleFlag{name: name, value: value})
	}
	return invocation, nil
}

func (invocation commandInvocation) source() string {
	words := append([]string{"sumup"}, invocation.path...)
	for _, argument := range invocation.arguments {
		words = append(words, shellQuote(argument))
	}
	var source strings.Builder
	source.WriteString(strings.Join(words, " "))
	for _, flag := range invocation.flags {
		source.WriteString(" \\\n  ")
		source.WriteString("--")
		source.WriteString(flag.name)
		if !flag.boolean {
			source.WriteByte(' ')
			source.WriteString(shellQuote(flag.value))
		}
	}
	source.WriteByte('\n')
	return source.String()
}

func (invocation commandInvocation) args() []string {
	args := append([]string{"sumup"}, invocation.path...)
	args = append(args, invocation.arguments...)
	for _, flag := range invocation.flags {
		args = append(args, "--"+flag.name)
		if !flag.boolean {
			args = append(args, flag.value)
		}
	}
	return args
}

func shellQuote(value string) string {
	if strings.HasPrefix(value, "$") {
		return `"` + value + `"`
	}
	return strconv.Quote(value)
}

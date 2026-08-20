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
	"github.com/sumup/sumup-cli/internal/currency"
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

// Generate builds a deterministic sample catalog for the current CLI command
// tree and generated OpenAPI operation catalog.
func Generate(cliVersion string) (*Catalog, error) {
	cliVersion = strings.TrimSpace(cliVersion)
	if cliVersion == "" {
		return nil, errors.New("cli version is required")
	}

	spec, err := loadPinnedSpec()
	if err != nil {
		return nil, err
	}

	commandsByOperation := boundCommandsByOperation(commands.All())
	samples := make([]Sample, 0, len(apicommands.Operations))
	for _, operation := range apicommands.Operations {
		command, err := commandForOperation(operation.ID, commandsByOperation[operation.ID])
		if err != nil {
			return nil, err
		}
		example := spec.exampleFor(operation.HTTPMethod, operation.Path)
		source, err := renderCommand(spec, command, example)
		if err != nil {
			return nil, fmt.Errorf("generate sample for %q: %w", operation.ID, err)
		}
		summary := operation.Summary
		if example.summary != "" {
			summary = example.summary
		}
		description := operation.Description
		if example.description != "" {
			description = example.description
		}
		samples = append(samples, Sample{
			ID:          operation.ID,
			OperationID: operation.ID,
			Example:     example.name,
			Summary:     summary,
			Description: description,
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

func renderCommand(document *openAPIDocument, bound boundCommand, example operationExample) (string, error) {
	invocation, err := buildInvocation(document, bound, example)
	if err != nil {
		return "", err
	}
	return invocation.source(), nil
}

func buildInvocation(document *openAPIDocument, bound boundCommand, example operationExample) (commandInvocation, error) {
	invocation := commandInvocation{path: strings.Fields(bound.path)}
	for _, match := range argumentPattern.FindAllStringSubmatch(bound.command.ArgsUsage, -1) {
		value, ok := argumentValue(document, example, match[1])
		if !ok {
			return commandInvocation{}, fmt.Errorf("no sample value for argument %q", match[1])
		}
		invocation.arguments = append(invocation.arguments, value)
	}

	flattened := flattenExample(example.body)
	for _, flag := range bound.command.Flags {
		name := flag.Names()[0]
		required := flagRequired(flag)
		_, inBody := lookupExample(name, flattened)
		allowSchema := required || name == "merchant-code" || !example.bodyProvided
		if !required && name != "merchant-code" && !inBody && example.bodyProvided {
			continue
		}

		values, ok := flagValues(document, example, flattened, name, required || name == "merchant-code", allowSchema)
		if !ok {
			if required {
				return commandInvocation{}, fmt.Errorf("no sample value for flag --%s", name)
			}
			continue
		}

		boolean := flagBoolean(flag)
		if boolean {
			if len(values) == 0 {
				continue
			}
			invocation.flags = append(invocation.flags, sampleFlag{name: name, boolean: true})
			continue
		}
		for _, value := range values {
			invocation.flags = append(invocation.flags, sampleFlag{name: name, value: value})
		}
	}
	return invocation, nil
}

func argumentValue(document *openAPIDocument, example operationExample, name string) (string, bool) {
	if value, ok := example.parameterExample(document, name); ok {
		formatted := formatExampleValue(value)
		if len(formatted) > 0 {
			return formatted[0], true
		}
	}
	parameter := example.parameter(name)
	if parameter != nil {
		if fallback := document.schemaFallback(parameter.Schema); fallback != nil {
			formatted := formatExampleValue(fallback)
			if len(formatted) > 0 {
				return formatted[0], true
			}
		}
	}
	return "", false
}

func flagValues(document *openAPIDocument, example operationExample, flattened map[string]any, name string, fromParameters, allowSchema bool) ([]string, bool) {
	if value, ok := lookupExample(name, flattened); ok {
		formatted := formatExampleValue(value)
		if name == "currency" {
			if supported, ok := supportedCurrencyValue(document, formatted); ok {
				return []string{supported}, true
			}
		} else if len(formatted) > 0 {
			return formatted, true
		}
	}
	if fromParameters {
		if value, ok := example.parameterExample(document, name); ok {
			formatted := formatExampleValue(value)
			if len(formatted) > 0 {
				return formatted, true
			}
		}
	}
	if allowSchema {
		if property := document.propertySchema(example.bodySchema, name); property != nil {
			if value, ok := document.schemaExample(property, nil); ok {
				formatted := formatExampleValue(value)
				if len(formatted) > 0 {
					return formatted, true
				}
			}
			if fromParameters {
				if fallback := document.schemaFallback(property); fallback != nil {
					formatted := formatExampleValue(fallback)
					if len(formatted) > 0 {
						return formatted, true
					}
				}
			}
		}
	}
	if fromParameters {
		parameter := example.parameter(name)
		if parameter != nil {
			if fallback := document.schemaFallback(parameter.Schema); fallback != nil {
				formatted := formatExampleValue(fallback)
				if len(formatted) > 0 {
					return formatted, true
				}
			}
		}
	}
	return nil, false
}

func supportedCurrencyValue(document *openAPIDocument, formatted []string) (string, bool) {
	if len(formatted) == 1 {
		if _, err := currency.Parse(formatted[0]); err == nil {
			return formatted[0], true
		}
	}
	if document.Components.Schemas == nil {
		return "", false
	}
	value, ok := document.schemaExample(document.Components.Schemas["Currency"], nil)
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	if _, err := currency.Parse(text); err != nil {
		return "", false
	}
	return text, true
}

func flagRequired(flag cli.Flag) bool {
	requirement, ok := flag.(interface{ IsRequired() bool })
	return ok && requirement.IsRequired()
}

func flagBoolean(flag cli.Flag) bool {
	boolean, ok := flag.(interface{ IsBoolFlag() bool })
	return ok && boolean.IsBoolFlag()
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

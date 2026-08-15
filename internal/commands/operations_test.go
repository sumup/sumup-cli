package commands

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/apicommands"
)

func TestOpenAPICatalogMatchesSDK(t *testing.T) {
	t.Parallel()

	clientType := reflect.TypeFor[sumup.Client]()
	sdkMethods := make([]string, 0)
	for index := range clientType.NumField() {
		field := clientType.Field(index)
		if !field.IsExported() {
			continue
		}
		for methodIndex := range field.Type.NumMethod() {
			method := field.Type.Method(methodIndex)
			sdkMethods = append(sdkMethods, field.Name+"."+method.Name)
		}
	}
	slices.Sort(sdkMethods)

	catalogMethods := make([]string, 0, len(apicommands.Operations))
	for _, operation := range apicommands.Operations {
		catalogMethods = append(catalogMethods, operation.Client+"."+operation.SDKMethod)
	}
	slices.Sort(catalogMethods)

	assert.Equal(t, sdkMethods, catalogMethods)
}

func TestCommandsCoverOpenAPICatalog(t *testing.T) {
	t.Parallel()

	apiGroups := make(map[string]struct{})
	for _, operation := range apicommands.Operations {
		apiGroups[strings.ToLower(operation.Client)] = struct{}{}
	}

	commandsByOperation := make(map[string][]string)
	unbound := make([]string, 0)
	walkLeafCommands(All(), func(path string, command *cli.Command) {
		group, _, _ := strings.Cut(path, " ")
		if _, ok := apiGroups[group]; !ok {
			return
		}

		operationID, ok := apicommands.OperationID(command)
		if !ok {
			unbound = append(unbound, path)
			return
		}
		commandsByOperation[operationID] = append(commandsByOperation[operationID], path)
	})
	slices.Sort(unbound)

	missing := make([]string, 0)
	for _, operation := range apicommands.Operations {
		if _, ok := commandsByOperation[operation.ID]; !ok {
			missing = append(missing, operation.Client+"."+operation.SDKMethod+" ("+operation.ID+")")
		}
	}
	slices.Sort(missing)

	duplicates := make([]string, 0)
	for operationID, paths := range commandsByOperation {
		if len(paths) < 2 {
			continue
		}
		slices.Sort(paths)
		duplicates = append(duplicates, operationID+": "+strings.Join(paths, ", "))
	}
	slices.Sort(duplicates)

	require.Empty(t, unbound, "API commands without an OpenAPI operation binding")
	require.Empty(t, duplicates, "OpenAPI operations exposed by more than one CLI command")
	assert.Empty(t, missing, "SDK operations without a CLI command")
}

func walkLeafCommands(commands []*cli.Command, visit func(string, *cli.Command)) {
	var walk func([]string, *cli.Command)
	walk = func(parents []string, command *cli.Command) {
		path := append(parents, command.Name)
		if len(command.Commands) == 0 {
			visit(strings.Join(path, " "), command)
			return
		}
		for _, child := range command.Commands {
			walk(path, child)
		}
	}

	for _, command := range commands {
		walk(nil, command)
	}
}

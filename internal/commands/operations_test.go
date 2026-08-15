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

	covered := make(map[string]struct{})
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
		covered[operationID] = struct{}{}
	})
	slices.Sort(unbound)

	missing := make([]string, 0)
	for _, operation := range apicommands.Operations {
		if _, ok := covered[operation.ID]; !ok {
			missing = append(missing, operation.Client+"."+operation.SDKMethod+" ("+operation.ID+")")
		}
	}
	slices.Sort(missing)

	require.Empty(t, unbound, "API commands without an OpenAPI operation binding")
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

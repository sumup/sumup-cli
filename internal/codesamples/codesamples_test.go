package codesamples

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	catalog, err := Generate("v1.2.3")

	require.NoError(t, err)
	assert.Equal(t, 1, catalog.SchemaVersion)
	assert.Equal(t, "bash", catalog.Language)
	assert.Equal(t, SDK{Module: "github.com/sumup/sumup-cli", Version: "v1.2.3"}, catalog.SDK)
	assert.Equal(t, apicommands.CatalogOpenAPIVersion, catalog.OpenAPIVersion)
	require.Len(t, catalog.Samples, len(apicommands.Operations))
	assert.True(t, slices.IsSortedFunc(catalog.Samples, func(a, b Sample) int {
		return strings.Compare(a.ID, b.ID)
	}))

	seen := make(map[string]struct{}, len(catalog.Samples))
	for _, sample := range catalog.Samples {
		assert.Equal(t, sample.ID, sample.OperationID)
		assert.NotEmpty(t, sample.HTTPMethod)
		assert.NotEmpty(t, sample.Path)
		assert.NotContains(t, sample.Source, "<")
		if _, ok := seen[sample.ID]; ok {
			t.Errorf("duplicate sample ID %q", sample.ID)
		}
		seen[sample.ID] = struct{}{}

		command := exec.CommandContext(t.Context(), "sh", "-n")
		command.Stdin = strings.NewReader(sample.Source)
		output, err := command.CombinedOutput()
		assert.NoError(t, err, "sample %q is not valid shell syntax:\n%s\n%s", sample.ID, sample.Source, output)
	}

	createCheckout := sampleByID(t, catalog.Samples, "CreateCheckout")
	assert.Equal(t, "Checkout", createCheckout.Example)
	assert.Equal(t, `sumup checkouts create \
  --reference "f00a8f74-b05d-4605-bd73-2a901bae5802" \
  --amount "10.1" \
  --currency "EUR" \
  --merchant-code "MH4H92C7" \
  --description "Purchase" \
  --redirect-url "https://sumup.com" \
  --valid-until "2020-02-29T10:56:56+00:00"
`, createCheckout.Source)
	assert.Equal(t, `sumup merchants persons get "pers_5AKFHN2KSK8D3TS79DJE3P3A2Z" \
  --merchant-code "MK10CL2A"
`, sampleByID(t, catalog.Samples, "GetPerson").Source)
	assert.Contains(t, sampleByID(t, catalog.Samples, "UpdateCheckout").Source, `--description "Updated purchase"`)
	assert.Contains(t, sampleByID(t, catalog.Samples, "CreateGoReaderCheckout").Source, "sumup readers go-checkout")
	assert.Contains(t, sampleByID(t, catalog.Samples, "CreateMerchantMember").Source, "sumup members create")
	assert.NotContains(t, sampleByID(t, catalog.Samples, "CreateMerchantMember").Source, "members invite")

	encodedSample, err := json.Marshal(sampleByID(t, catalog.Samples, "CreateCheckout"))
	require.NoError(t, err)
	assert.Contains(t, string(encodedSample), `"sample":`)
	assert.NotContains(t, string(encodedSample), `"source":`)
}

func TestGenerateDeterministic(t *testing.T) {
	t.Parallel()

	first, err := Generate("v1.2.3")
	require.NoError(t, err)
	second, err := Generate("v1.2.3")
	require.NoError(t, err)

	firstJSON, err := json.Marshal(first)
	require.NoError(t, err)
	secondJSON, err := json.Marshal(second)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(firstJSON, secondJSON))
}

func TestGenerateRequiresVersion(t *testing.T) {
	t.Parallel()

	_, err := Generate("  ")

	require.EqualError(t, err, "cli version is required")
}

func TestGeneratedInvocationsReachAPITransport(t *testing.T) {
	for _, operation := range apicommands.Operations {
		t.Run(operation.ID, func(t *testing.T) {
			resourceCommands := commands.All()
			commandsByOperation := boundCommandsByOperation(resourceCommands)
			bound, err := commandForOperation(operation.ID, commandsByOperation[operation.ID])
			require.NoError(t, err)
			spec, err := loadPinnedSpec()
			require.NoError(t, err)
			invocation, err := buildInvocation(spec, bound, spec.exampleFor(operation.HTTPMethod, operation.Path))
			require.NoError(t, err)

			transportError := errors.New("sample reached API transport")
			called := false
			httpClient := &http.Client{Transport: sampleRoundTripFunc(func(*http.Request) (*http.Response, error) {
				called = true
				return nil, transportError
			})}
			root := &cli.Command{
				Name:     "sumup",
				Commands: resourceCommands,
				Metadata: map[string]any{
					app.ContextKey: &app.Context{
						Client:       sumup.NewClient(sumupclient.WithClient(httpClient), sumupclient.WithBaseURL("https://api.test")),
						MerchantCode: "M123",
						Output:       io.Discard,
						StatusOutput: io.Discard,
					},
				},
			}

			err = root.Run(t.Context(), invocation.args())

			assert.True(t, called, "sample failed before making an API request: %v", err)
			require.ErrorIs(t, err, transportError)
		})
	}
}

func TestCommandForOperationRejectsDuplicates(t *testing.T) {
	t.Parallel()

	_, err := commandForOperation("ExampleOperation", []boundCommand{{path: "examples first"}, {path: "examples second"}})

	require.EqualError(t, err, `OpenAPI operation "ExampleOperation" has multiple CLI commands (examples first, examples second)`)
}

func sampleByID(t *testing.T, samples []Sample, id string) Sample {
	t.Helper()
	for _, sample := range samples {
		if sample.ID == id {
			return sample
		}
	}
	t.Fatalf("sample %q not found", id)
	return Sample{}
}

type sampleRoundTripFunc func(*http.Request) (*http.Response, error)

func (function sampleRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

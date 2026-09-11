package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunIncludesOnlyExplicitlyUnsupportedOperations(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	specPath := filepath.Join(directory, "sdk.json")
	unsupportedSpecPath := filepath.Join(directory, "openapi.json")
	outputPath := filepath.Join(directory, "catalog.go")
	require.NoError(t, os.WriteFile(specPath, []byte(`{
		"info":{"version":"1.0.0"},
		"paths":{"/widgets":{"get":{"operationId":"ListWidgets","tags":["Widgets"],"x-codegen":{"method_name":"list"}}}}
	}`), 0o600))
	require.NoError(t, os.WriteFile(unsupportedSpecPath, []byte(`{
		"info":{"version":"1.0.0"},
		"paths":{
			"/checkouts/{checkout_id}":{"put":{"operationId":"ProcessCheckout","tags":["Checkouts"],"x-codegen":{"method_name":"process"}}},
			"/other":{"get":{"operationId":"OtherOperation","tags":["Other"],"x-codegen":{"method_name":"get"}}},
			"/widgets":{"get":{"operationId":"ListWidgets","summary":"Must not override the SDK","tags":["Widgets"],"x-codegen":{"method_name":"list"}}}
		}
	}`), 0o600))

	require.NoError(t, run(outputPath, specPath, "v1.2.3", unsupportedSpecPath))
	generated, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	output := string(generated)
	assert.Regexp(t, `ID:\s+"ProcessCheckout"`, output)
	assert.Contains(t, output, "Unsupported: true")
	assert.Equal(t, 1, strings.Count(output, `"ListWidgets"`))
	assert.NotContains(t, output, "OtherOperation")
	assert.NotContains(t, output, "Must not override the SDK")
}

func TestParseOperations(t *testing.T) {
	t.Parallel()

	spec := []byte(`{
		"info":{"version":"1.2.3"},
		"paths":{
			"/widgets/{widget_id}":{
				"parameters":[{"name":"widget_id","in":"path","required":true,"schema":{"type":"string"}}],
				"post":{
					"operationId":"CreateWidget",
					"summary":"Create a widget",
					"description":"Creates one widget.",
					"tags":["Widgets"],
					"x-codegen":{"method_name":"create_widget"},
					"parameters":[{"name":"version","in":"query","schema":{"type":"string"}}],
					"requestBody":{"required":true,"content":{"application/json":{"schema":{"$ref":"#/components/schemas/WidgetRequest"}}}}
				}
			}
		}
	}`)

	document, operations, err := parseOperations(spec)

	require.NoError(t, err)
	assert.Equal(t, "1.2.3", document.Info.Version)
	require.Len(t, operations, 1)
	assert.Equal(t, operation{
		ID:          "CreateWidget",
		Client:      "Widgets",
		SDKMethod:   "CreateWidget",
		HTTPMethod:  "POST",
		Path:        "/widgets/{widget_id}",
		Summary:     "Create a widget",
		Description: "Creates one widget.",
		Parameters: []parameter{
			{Name: "widget_id", Location: "path", Type: "string", Required: true},
			{Name: "version", Location: "query", Type: "string"},
		},
		RequestBody: &requestBody{Schema: "WidgetRequest", Required: true},
	}, operations[0])
}

func TestRenderCatalog(t *testing.T) {
	t.Parallel()

	generated, err := renderCatalog("1.2.3", "v4.5.6", []byte("spec"), []operation{{
		ID:          "ListWidgets",
		Client:      "Widgets",
		SDKMethod:   "List",
		HTTPMethod:  "GET",
		Path:        "/widgets",
		Summary:     "List widgets",
		Description: "Returns every widget.",
	}})

	require.NoError(t, err)
	output := string(generated)
	assert.Contains(t, output, `CatalogSDKVersion     = "v4.5.6"`)
	assert.Regexp(t, `ID:\s+"ListWidgets"`, output)
	assert.Regexp(t, `SDKMethod:\s+"List"`, output)
	assert.Contains(t, output, `Description: "Returns every widget."`)
	assert.True(t, strings.HasSuffix(output, "}\n"))
}

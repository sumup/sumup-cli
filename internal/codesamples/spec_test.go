package codesamples

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaExampleFollowsRefChain(t *testing.T) {
	t.Parallel()

	document, err := parseSpec([]byte(`{
		"components": {
			"schemas": {
				"Currency": {"type": "string", "enum": ["BGN", "EUR"], "example": "EUR"},
				"Amount": {
					"allOf": [
						{"$ref": "#/components/schemas/Currency"}
					]
				}
			}
		}
	}`))
	require.NoError(t, err)

	value, ok := document.schemaExample(&specSchema{Ref: "#/components/schemas/Amount"}, nil)

	require.True(t, ok)
	assert.Equal(t, "EUR", value)
}

func TestExampleForPrefersFirstNamedRequestExample(t *testing.T) {
	t.Parallel()

	document, err := parseSpec([]byte(`{
		"paths": {
			"/widgets": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {"$ref": "#/components/schemas/Widget"},
								"examples": {
									"Zulu": {"value": {"name": "zulu"}},
									"Alpha": {"value": {"name": "alpha", "color": "blue"}}
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Widget": {
					"type": "object",
					"properties": {
						"name": {"type": "string", "example": "property-name"},
						"color": {"type": "string", "example": "property-color"}
					}
				}
			}
		}
	}`))
	require.NoError(t, err)

	example := document.exampleFor("POST", "/widgets")

	assert.Equal(t, "Alpha", example.name)
	assert.Equal(t, map[string]any{"name": "alpha", "color": "blue"}, example.body)
	assert.True(t, example.bodyProvided)
}

func TestExampleForWalksPropertyExamplesWhenRequestExampleIsMissing(t *testing.T) {
	t.Parallel()

	document, err := parseSpec([]byte(`{
		"paths": {
			"/widgets": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {"$ref": "#/components/schemas/Widget"}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Widget": {
					"type": "object",
					"properties": {
						"name": {"type": "string", "example": "Widget"},
						"nested": {"$ref": "#/components/schemas/Details"}
					}
				},
				"Details": {
					"type": "object",
					"properties": {
						"email": {"type": "string", "example": "user@example.com"}
					}
				}
			}
		}
	}`))
	require.NoError(t, err)

	example := document.exampleFor("POST", "/widgets")

	assert.False(t, example.bodyProvided)
	assert.Equal(t, map[string]any{
		"name":   "Widget",
		"nested": map[string]any{"email": "user@example.com"},
	}, example.body)
}

func TestLookupExampleMapsCLIFlagsOntoNestedProperties(t *testing.T) {
	t.Parallel()

	values := flattenExample(map[string]any{
		"checkout_reference": "ref-1",
		"personal_details": map[string]any{
			"email": "user@example.com",
			"address": map[string]any{
				"line_1": "Sample street",
			},
		},
	})

	reference, ok := lookupExample("reference", values)
	require.True(t, ok)
	assert.Equal(t, "ref-1", reference)

	email, ok := lookupExample("email", values)
	require.True(t, ok)
	assert.Equal(t, "user@example.com", email)

	line, ok := lookupExample("address-line-1", values)
	require.True(t, ok)
	assert.Equal(t, "Sample street", line)
}

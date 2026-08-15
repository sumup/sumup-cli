package roles_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/roles"
	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"
)

func TestNewCommand(t *testing.T) {
	t.Run("list renders roles returned by the api", func(t *testing.T) {
		cmd, out, _ := newCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles", request.URL.Path)
			return jsonResponse(http.StatusOK, `{"items":[{"id":"role_owner","name":"Owner","description":"Full administrative access to the merchant account","is_predefined":true,"permissions":[],"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"},{"id":"role_cashier","name":"Cashier","description":"Limited access for point-of-sale operations","is_predefined":true,"permissions":[],"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"}]}`)
		})

		err := cmd.Run(t.Context(), []string{"roles", "list", "--merchant-code", "M123"})

		require.NoError(t, err)
		assert.Equal(t, "Roles\nRole Name Description\nrole_owner Owner Full administrative access to the merchant account\nrole_cashier Cashier Limited access for point-of-sale operations", normalizeOutput(out.String()))
	})

	t.Run("create sends role fields and renders the created role", func(t *testing.T) {
		cmd, out, statusOut := newCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodPost, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles", request.URL.Path)

			var body sumup.RolesCreateParams
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			assert.Equal(t, "Supervisor", body.Name)
			assert.Equal(t, []string{"transactions.read", "payouts.read"}, body.Permissions)
			assert.Equal(t, "Can review payments", requireValue(t, body.Description))
			return jsonResponse(http.StatusCreated, `{"id":"role_supervisor","name":"Supervisor","description":"Can review payments","permissions":["transactions.read","payouts.read"],"is_predefined":false,"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"}`)
		})

		err := cmd.Run(t.Context(), []string{"roles", "create", "--merchant-code", "M123", "--name", "Supervisor", "--permission", "transactions.read", "--permission", "payouts.read", "--description", "Can review payments"})

		require.NoError(t, err)
		assert.Contains(t, normalizeOutput(statusOut.String()), "Role created")
		assert.Contains(t, normalizeOutput(out.String()), "role_supervisor")
		assert.Contains(t, normalizeOutput(out.String()), "transactions.read, payouts.read")
	})

	t.Run("get renders a role returned by the api", func(t *testing.T) {
		cmd, out, _ := newCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles/role_supervisor", request.URL.Path)
			return jsonResponse(http.StatusOK, `{"id":"role_supervisor","name":"Supervisor","permissions":["transactions.read"],"is_predefined":false,"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"}`)
		})

		err := cmd.Run(t.Context(), []string{"roles", "get", "role_supervisor"})

		require.NoError(t, err)
		assert.Contains(t, normalizeOutput(out.String()), "role_supervisor")
		assert.Contains(t, normalizeOutput(out.String()), "transactions.read")
	})

	t.Run("update sends only requested role fields", func(t *testing.T) {
		cmd, out, statusOut := newCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodPatch, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles/role_supervisor", request.URL.Path)

			var body map[string]any
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			assert.Equal(t, map[string]any{
				"name":        "Payment reviewer",
				"permissions": []any{"transactions.read"},
			}, body)
			return jsonResponse(http.StatusOK, `{"id":"role_supervisor","name":"Payment reviewer","permissions":["transactions.read"],"is_predefined":false,"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T11:00:00Z"}`)
		})

		err := cmd.Run(t.Context(), []string{"roles", "update", "role_supervisor", "--name", "Payment reviewer", "--permission", "transactions.read"})

		require.NoError(t, err)
		assert.Contains(t, normalizeOutput(statusOut.String()), "Role updated")
		assert.Contains(t, normalizeOutput(out.String()), "Payment reviewer")
	})

	t.Run("delete removes the requested role", func(t *testing.T) {
		cmd, _, statusOut := newCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodDelete, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles/role_supervisor", request.URL.Path)
			return jsonResponse(http.StatusOK, "")
		})

		err := cmd.Run(t.Context(), []string{"roles", "delete", "role_supervisor"})

		require.NoError(t, err)
		assert.Contains(t, normalizeOutput(statusOut.String()), "Role deleted")
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func newCommand(t *testing.T, handler func(*http.Request) *http.Response) (*cli.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer
	var statusOut bytes.Buffer
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return handler(request), nil
	})}
	cmd := roles.NewCommand()
	cmd.Metadata = map[string]any{
		app.ContextKey: &app.Context{
			Client:       sumup.NewClient(sumupclient.WithClient(httpClient), sumupclient.WithBaseURL("https://api.test")),
			MerchantCode: "M123",
			Output:       &out,
			StatusOutput: &statusOut,
		},
	}
	return cmd, &out, &statusOut
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func requireValue[T any](t *testing.T, value *T) T {
	t.Helper()
	require.NotNil(t, value)
	return *value
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func normalizeOutput(value string) string {
	value = ansiPattern.ReplaceAllString(strings.ReplaceAll(value, "\r\n", "\n"), "")
	lines := strings.Split(value, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.Join(normalized, "\n")
}

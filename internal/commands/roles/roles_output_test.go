package roles_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/roles"
	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"
)

func TestNewCommand(t *testing.T) {
	t.Run("list renders roles returned by the api", func(t *testing.T) {
		var out bytes.Buffer

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v0.1/merchants/M123/roles", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"items":[{"id":"role_owner","name":"Owner","description":"Full administrative access to the merchant account","is_predefined":true,"permissions":[],"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"},{"id":"role_cashier","name":"Cashier","description":"Limited access for point-of-sale operations","is_predefined":true,"permissions":[],"created_at":"2026-03-27T10:00:00Z","updated_at":"2026-03-27T10:00:00Z"}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		cmd := roles.NewCommand()
		cmd.Metadata = map[string]any{
			app.ContextKey: &app.Context{
				Client:       sumup.NewClient(sumupclient.WithBaseURL(server.URL)),
				MerchantCode: "M123",
				Output:       &out,
			},
		}

		err := cmd.Run(context.Background(), []string{"roles", "list", "--merchant-code", "M123"})

		require.NoError(t, err)
		assert.Equal(t, "Roles\nRole Name Description\nrole_owner Owner Full administrative access to the merchant account\nrole_cashier Cashier Limited access for point-of-sale operations", normalizeOutput(out.String()))
	})
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

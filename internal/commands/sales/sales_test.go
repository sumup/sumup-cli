package sales

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
)

const saleJSON = `{"id":"sale_123","name":"Coffee","state":"CLOSED","currency":"EUR","closed_at":"2026-11-01T10:00:00Z","price":{"gross_amount":{"value":925,"currency":"EUR"}},"items":[{"id":"line_1","name":"Cappuccino","quantity":2}],"payments":[{"reference":"TX123","method":"card","amount":975,"status":"SUCCESSFUL"}]}`

func TestListSales(t *testing.T) {
	cmd, out := testCommand(t, false, func(r *http.Request) *http.Response {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1.0/merchants/M123/sales", r.URL.Path)
		assert.Equal(t, "25", r.URL.Query().Get("pagination_limit"))
		assert.Equal(t, "opaque+/=", r.URL.Query().Get("pagination_token"))
		return response(http.StatusOK, `{"items":[`+saleJSON+`],"pagination_token":"next-token"}`)
	})
	require.NoError(t, cmd.Run(t.Context(), []string{"sales", "list", "--pagination-limit", "25", "--pagination-token", "opaque+/="}))
	for _, value := range []string{"sale_123", "Coffee", "925", "EUR", "minor units", "--pagination-token next-token"} {
		assert.Contains(t, out.String(), value)
	}
}

func TestGetSale(t *testing.T) {
	for _, jsonOutput := range []bool{false, true} {
		t.Run(map[bool]string{false: "human", true: "json"}[jsonOutput], func(t *testing.T) {
			cmd, out := testCommand(t, jsonOutput, func(r *http.Request) *http.Response {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1.0/merchants/OTHER/sales/sale_123", r.URL.Path)
				return response(http.StatusOK, saleJSON)
			})
			require.NoError(t, cmd.Run(t.Context(), []string{"sales", "get", "--merchant-code", "OTHER", "sale_123"}))
			if jsonOutput {
				var sale map[string]any
				require.NoError(t, json.Unmarshal(out.Bytes(), &sale))
				assert.Equal(t, "sale_123", sale["id"])
				assert.Len(t, sale["payments"], 1)
			} else {
				for _, value := range []string{"sale_123", "Cappuccino", "TX123", "975", "925"} {
					assert.Contains(t, out.String(), value)
				}
			}
		})
	}
}

func TestSearchSales(t *testing.T) {
	cmd, out := testCommand(t, true, func(r *http.Request) *http.Response {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1.0/merchants/M123/sales/search", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"pagination_limit": float64(10), "pagination_token": "search-token",
			"start_date": "2026-10-01T00:00:00Z", "end_date": "2026-11-01T00:00:00Z",
			"payment_statuses": []any{"PAID", "CHARGED_BACK"},
			"sale_statuses":    []any{"CLOSED"}, "sales_channels": []any{"POS", "SOLO"},
		}, body)
		return response(http.StatusOK, `{"items":[],"pagination_token":"next-search-token"}`)
	})
	require.NoError(t, cmd.Run(t.Context(), []string{"sales", "search",
		"--pagination-limit", "10", "--pagination-token", "search-token",
		"--start-date", "2026-10-01T00:00:00Z", "--end-date", "2026-11-01T00:00:00Z",
		"--payment-status", "PAID", "--payment-status", "CHARGED_BACK",
		"--sale-status", "CLOSED", "--sales-channel", "POS", "--sales-channel", "SOLO",
	}))
	assert.JSONEq(t, `{"items":[],"pagination_token":"next-search-token"}`, out.String())
}

func TestSalesRejectInvalidInput(t *testing.T) {
	for _, args := range [][]string{
		{"get"}, {"get", "one", "two"}, {"list", "--pagination-limit", "0"},
		{"search", "--pagination-limit", "26"}, {"search", "--start-date", "yesterday"},
		{"search", "--start-date", "2026-11-01T00:00:00Z", "--end-date", "2026-10-01T00:00:00Z"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd, _ := testCommand(t, false, func(*http.Request) *http.Response {
				t.Fatal("invalid input must not reach the API")
				return nil
			})
			require.Error(t, cmd.Run(t.Context(), append([]string{"sales"}, args...)))
		})
	}
}

func TestSalesAPIError(t *testing.T) {
	cmd, _ := testCommand(t, false, func(*http.Request) *http.Response {
		return response(http.StatusForbidden, `{"type":"about:blank","title":"Forbidden","status":403}`)
	})
	require.ErrorContains(t, cmd.Run(t.Context(), []string{"sales", "list"}), "list sales")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testCommand(t *testing.T, jsonOutput bool, handler func(*http.Request) *http.Response) (*cli.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("SUMUP_MERCHANT_CODE", "M123")
	var out bytes.Buffer
	cmd := NewCommand()
	cmd.Metadata = map[string]any{app.ContextKey: &app.Context{
		Client:       sumup.NewClient(sumupclient.WithClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return handler(r), nil })}), sumupclient.WithBaseURL("https://api.test")),
		MerchantCode: "M123", Output: &out, StatusOutput: io.Discard, JSONOutput: jsonOutput,
	}}
	return cmd, &out
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
}

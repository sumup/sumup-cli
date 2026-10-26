package readers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestReaderSettings(t *testing.T) {
	t.Parallel()

	t.Run("get renders menu lock state", func(t *testing.T) {
		t.Parallel()

		cmd, out := newReadersCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/readers/reader_42/settings", request.URL.Path)
			return readersJSONResponse(`{"data":{"menu_lock":{"enabled":true,"passcode":"0123"}}}`)
		})

		err := cmd.Run(t.Context(), []string{"readers", "settings", "get", "reader_42"})

		require.NoError(t, err)
		assert.Contains(t, out.String(), "true")
		assert.NotContains(t, out.String(), "0123")
	})

	t.Run("update preserves false and omits unspecified passcode", func(t *testing.T) {
		t.Parallel()

		cmd, out := newReadersCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodPatch, request.Method)
			assert.Equal(t, "/v0.1/merchants/M123/readers/reader_42/settings", request.URL.Path)
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"menu_lock":{"enabled":false}}`, string(body))
			return readersJSONResponse(`{"data":{"status":"applied","menu_lock":{"enabled":false}}}`)
		})

		err := cmd.Run(t.Context(), []string{"readers", "settings", "update", "--menu-lock-enabled=false", "reader_42"})

		require.NoError(t, err)
		assert.Contains(t, out.String(), "applied")
		assert.Contains(t, out.String(), "false")
	})

	t.Run("update preserves leading zeros in passcode", func(t *testing.T) {
		t.Parallel()

		cmd, _ := newReadersCommand(t, func(request *http.Request) *http.Response {
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"menu_lock":{"passcode":"0123"}}`, string(body))
			return readersJSONResponse(`{"data":{"status":"applied","menu_lock":{"passcode":"0123"}}}`)
		})

		err := cmd.Run(t.Context(), []string{"readers", "settings", "update", "--menu-lock-passcode", "0123", "reader_42"})

		require.NoError(t, err)
	})

	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "empty update", args: []string{"reader_42"}, want: "provide --menu-lock-enabled or --menu-lock-passcode"},
		{name: "invalid passcode", args: []string{"--menu-lock-passcode", "123", "reader_42"}, want: "menu lock passcode must contain exactly four digits"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd, _ := newReadersCommand(t, func(*http.Request) *http.Response {
				t.Fatal("API should not be called")
				return nil
			})

			err := cmd.Run(t.Context(), append([]string{"readers", "settings", "update"}, test.args...))

			require.EqualError(t, err, test.want)
		})
	}
}

func TestReaderCheckoutPayloads(t *testing.T) {
	t.Parallel()

	t.Run("affiliate transaction reference is optional", func(t *testing.T) {
		t.Parallel()

		cmd, _ := newReadersCommand(t, func(request *http.Request) *http.Response {
			var body sumup.ReadersCreateCheckoutParams
			require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
			affiliate := body.Affiliate.Value()
			require.NotNil(t, affiliate)
			assert.Equal(t, "com.example.app", affiliate.AppID)
			assert.Equal(t, "example-key", affiliate.Key)
			assert.Nil(t, affiliate.ForeignTransactionID)
			require.NotNil(t, body.PaymentType)
			assert.Equal(t, sumup.CreateCheckoutRequestPaymentTypePix, *body.PaymentType)
			response := readersJSONResponse(`{"data":{}}`)
			response.StatusCode = http.StatusCreated
			return response
		})

		err := cmd.Run(t.Context(), []string{"readers", "checkout", "--amount", "10", "--currency", "BRL", "--payment-type", "pix", "--affiliate-app-id", "com.example.app", "--affiliate-key", "example-key", "reader_42"})

		require.NoError(t, err)
	})

	t.Run("go checkout sends amount and affiliate", func(t *testing.T) {
		t.Parallel()

		cmd, _ := newReadersCommand(t, func(request *http.Request) *http.Response {
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"client_transaction_id":"order-123","total_amount":{"currency":"EUR","value":1000},"affiliate":{"app_id":"com.example.app","key":"example-key"}}`, string(body))
			return readersJSONResponse(`{"data":{}}`)
		})

		err := cmd.Run(t.Context(), []string{"readers", "go-checkout", "--amount", "10", "--currency", "EUR", "--client-transaction-id", "order-123", "--affiliate-app-id", "com.example.app", "--affiliate-key", "example-key", "reader_42"})

		require.NoError(t, err)
	})
}

type readersRoundTripFunc func(*http.Request) (*http.Response, error)

func (function readersRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func newReadersCommand(t *testing.T, handler func(*http.Request) *http.Response) (*cli.Command, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer
	httpClient := &http.Client{Transport: readersRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		return handler(request), nil
	})}
	cmd := NewCommand()
	cmd.Metadata = map[string]any{
		app.ContextKey: &app.Context{
			Client:       sumup.NewClient(sumupclient.WithClient(httpClient), sumupclient.WithBaseURL("https://api.test")),
			MerchantCode: "M123",
			Output:       &out,
			StatusOutput: io.Discard,
		},
	}
	return cmd, &out
}

func readersJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

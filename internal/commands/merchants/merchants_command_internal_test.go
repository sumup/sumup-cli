package merchants

import (
	"bytes"
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

func TestPersonCommands(t *testing.T) {
	t.Run("list sends the requested version and renders persons", func(t *testing.T) {
		cmd, out := newMerchantsCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/v1/merchants/M123/persons", request.URL.Path)
			assert.Equal(t, "latest", request.URL.Query().Get("version"))
			return merchantsJSONResponse(http.StatusOK, `{"items":[{"id":"person_123","given_name":"Ada","family_name":"Lovelace","relationships":["representative"],"user_id":"user_123"}]}`)
		})

		err := cmd.Run(t.Context(), []string{"merchants", "persons", "list", "--version", "latest"})

		require.NoError(t, err)
		for _, want := range []string{"Persons", "person_123", "Ada Lovelace", "representative", "user_123"} {
			assert.Contains(t, out.String(), want)
		}
	})

	t.Run("get renders the requested person", func(t *testing.T) {
		cmd, out := newMerchantsCommand(t, func(request *http.Request) *http.Response {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/v1/merchants/M123/persons/person_123", request.URL.Path)
			assert.Empty(t, request.URL.Query().Get("version"))
			return merchantsJSONResponse(http.StatusOK, `{"id":"person_123","given_name":"Ada","middle_name":"Byron","family_name":"Lovelace","birthdate":"1815-12-10","citizenship":"GB","phone_number":"+441234567890","relationships":["representative"]}`)
		})

		err := cmd.Run(t.Context(), []string{"merchants", "persons", "get", "person_123"})

		require.NoError(t, err)
		for _, want := range []string{"person_123", "Ada", "Byron", "Lovelace", "1815-12-10", "GB", "+441234567890", "representative"} {
			assert.Contains(t, out.String(), want)
		}
	})

	t.Run("rejects unsupported resource versions before calling the api", func(t *testing.T) {
		cmd, _ := newMerchantsCommand(t, func(*http.Request) *http.Response {
			t.Fatal("API should not be called")
			return nil
		})

		err := cmd.Run(t.Context(), []string{"merchants", "get", "--version", "draft"})

		require.EqualError(t, err, "version must be latest")
	})
}

type merchantsRoundTripFunc func(*http.Request) (*http.Response, error)

func (function merchantsRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func newMerchantsCommand(t *testing.T, handler func(*http.Request) *http.Response) (*cli.Command, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer
	httpClient := &http.Client{Transport: merchantsRoundTripFunc(func(request *http.Request) (*http.Response, error) {
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

func merchantsJSONResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

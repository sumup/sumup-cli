package tokens

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestCreateToken(t *testing.T) {
	t.Parallel()

	for _, jsonOutput := range []bool{false, true} {
		name := "requires JSON output before making a request"
		if jsonOutput {
			name = "JSON output includes token and expiry"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			called := false
			httpClient := &http.Client{Transport: tokensRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				called = true
				assert.Equal(t, http.MethodPost, request.Method)
				assert.Equal(t, "/v0.1/merchants/M123/tokens", request.URL.Path)
				body, err := io.ReadAll(request.Body)
				require.NoError(t, err)
				assert.JSONEq(t, `{"actions":["checkout"]}`, string(body))
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"access_token":"example-token","expires_in":900,"token_type":"Bearer"}`)),
				}, nil
			})}
			cmd := NewCommand()
			cmd.Metadata = map[string]any{
				app.ContextKey: &app.Context{
					Client:       sumup.NewClient(sumupclient.WithClient(httpClient), sumupclient.WithBaseURL("https://api.test")),
					MerchantCode: "M123",
					Output:       &out,
					StatusOutput: io.Discard,
					JSONOutput:   jsonOutput,
				},
			}

			err := cmd.Run(t.Context(), []string{"tokens", "create", "--action", "checkout"})

			if jsonOutput {
				require.NoError(t, err)
				require.True(t, called)
				assert.JSONEq(t, `{"access_token":"example-token","expires_in":900,"token_type":"Bearer"}`, out.String())
			} else {
				require.EqualError(t, err, "creating a merchant token requires --json to output the access token")
				assert.False(t, called)
				assert.Empty(t, out.String())
			}
		})
	}
}

type tokensRoundTripFunc func(*http.Request) (*http.Response, error)

func (function tokensRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

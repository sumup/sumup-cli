package transactions

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup/sumup-cli/internal/app"
)

func TestRenderRefundResult(t *testing.T) {
	t.Run("writes status message in human mode", func(t *testing.T) {
		var statusOut bytes.Buffer

		appCtx := &app.Context{StatusOutput: &statusOut}

		err := renderRefundResult(appCtx)

		require.NoError(t, err)
		assert.Contains(t, statusOut.String(), "Transaction refunded")
	})

	t.Run("prints json acknowledgement when requested", func(t *testing.T) {
		var out bytes.Buffer

		appCtx := &app.Context{
			JSONOutput: true,
			Output:     &out,
		}

		err := renderRefundResult(appCtx)

		require.NoError(t, err)
		assert.Contains(t, out.String(), `"status": "refunded"`)
	})
}

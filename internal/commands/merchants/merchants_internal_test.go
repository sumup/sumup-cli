package merchants

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestRenderMerchant(t *testing.T) {
	t.Run("renders merchant and business profile details", func(t *testing.T) {
		var out bytes.Buffer
		createdAt := time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)
		updatedAt := createdAt.Add(2 * time.Hour)
		alias := "Main account"
		businessName := "Example Shop"
		businessEmail := "shop@example.com"
		businessWebsite := "https://example.com"
		sandbox := true

		err := renderMerchant(&app.Context{Output: &out, ExactTimestamps: true}, &sumup.Merchant{
			MerchantCode:    "M123",
			Alias:           &alias,
			Country:         "DE",
			DefaultCurrency: "EUR",
			DefaultLocale:   "de-DE",
			Sandbox:         &sandbox,
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
			BusinessProfile: &sumup.BusinessProfile{
				Name:    &businessName,
				Email:   &businessEmail,
				Website: &businessWebsite,
			},
		})

		require.NoError(t, err)

		rendered := out.String()
		for _, want := range []string{"M123", "Main account", "Example Shop", "shop@example.com", "https://example.com"} {
			assert.Contains(t, rendered, want)
		}
	})
}

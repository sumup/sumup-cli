package merchants

import (
	"bytes"
	"strings"
	"testing"
	"time"

	sumup "github.com/sumup/sumup-go"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestRenderMerchant(t *testing.T) {
	var out bytes.Buffer
	createdAt := time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	alias := "Main account"
	businessName := "Example Shop"
	businessEmail := "shop@example.com"
	businessWebsite := "https://example.com"
	sandbox := true

	renderMerchant(&app.Context{Output: &out, ExactTimestamps: true}, &sumup.Merchant{
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

	rendered := out.String()
	for _, want := range []string{"M123", "Main account", "Example Shop", "shop@example.com", "https://example.com"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("renderMerchant() output = %q, want substring %q", rendered, want)
		}
	}
}

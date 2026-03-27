package members

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup/sumup-cli/internal/app"
	sumup "github.com/sumup/sumup-go"
)

func TestRenderDeleteMemberResult(t *testing.T) {
	t.Run("prints json acknowledgement when requested", func(t *testing.T) {
		var out bytes.Buffer

		appCtx := &app.Context{
			JSONOutput: true,
			Output:     &out,
		}

		err := renderDeleteMemberResult(appCtx)

		require.NoError(t, err)
		assert.Contains(t, out.String(), `"status": "deleted"`)
	})

	t.Run("writes status message in human mode", func(t *testing.T) {
		var statusOut bytes.Buffer

		appCtx := &app.Context{StatusOutput: &statusOut}

		err := renderDeleteMemberResult(appCtx)

		require.NoError(t, err)
		assert.Contains(t, statusOut.String(), "Member deleted")
	})
}

func TestRenderMember(t *testing.T) {
	t.Run("uses exact local timestamps when requested", func(t *testing.T) {
		var out bytes.Buffer
		createdAt := time.Date(2026, time.March, 27, 10, 0, 0, 0, time.UTC)
		updatedAt := createdAt.Add(2 * time.Hour)

		member := &sumup.Member{
			ID:        "member-1",
			Roles:     []string{"role_employee"},
			Status:    sumup.MembershipStatusAccepted,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			User:      &sumup.MembershipUser{Email: "member@example.com"},
		}

		err := renderMember(&app.Context{Output: &out, ExactTimestamps: true}, &out, member)

		require.NoError(t, err)
		assert.Contains(t, out.String(), createdAt.In(time.Local).Format(time.RFC3339))
		assert.Contains(t, out.String(), updatedAt.In(time.Local).Format(time.RFC3339))
	})
}

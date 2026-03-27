package members

import (
	"bytes"
	"regexp"
	"strings"
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
		assert.Equal(t, "{\n  \"status\": \"deleted\"\n}\n", out.String())
	})

	t.Run("writes status message in human mode", func(t *testing.T) {
		var statusOut bytes.Buffer

		appCtx := &app.Context{StatusOutput: &statusOut}

		err := renderDeleteMemberResult(appCtx)

		require.NoError(t, err)
		assert.Equal(t, "✓ Member deleted\n", statusOut.String())
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
		assert.Equal(t, "ID: member-1\nEmail: member@example.com\nRoles: role_employee\nStatus: Accepted\nNickname: -\nCreated At: "+createdAt.In(time.Local).Format(time.RFC3339)+"\nUpdated At: "+updatedAt.In(time.Local).Format(time.RFC3339), normalizeOutput(out.String()))
	})
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func normalizeOutput(value string) string {
	value = ansiPattern.ReplaceAllString(strings.ReplaceAll(value, "\r\n", "\n"), "")
	lines := strings.Split(value, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.Join(normalized, "\n")
}

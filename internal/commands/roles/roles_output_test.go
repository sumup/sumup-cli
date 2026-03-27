package roles_test

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/roles"
)

func TestNewCommand(t *testing.T) {
	t.Run("list renders the human output contract", func(t *testing.T) {
		var out bytes.Buffer

		cmd := roles.NewCommand()
		cmd.Commands[0].Metadata = map[string]any{
			app.ContextKey: &app.Context{Output: &out},
		}

		listCmd := cmd.Commands[0]

		err := listCmd.Action(context.Background(), listCmd)

		require.NoError(t, err)
		assert.Equal(t, "Roles\nRole Display Name Description\nrole_owner Owner Full administrative access to the merchant account\nrole_admin Admin Administrative access with some restrictions\nrole_employee Employee Standard employee access for daily operations\nrole_manager Manager Management access with elevated permissions\nrole_cashier Cashier Limited access for point-of-sale operations", normalizeOutput(out.String()))
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

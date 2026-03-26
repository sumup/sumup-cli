package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/config"
)

func TestGetMerchantCode(t *testing.T) {
	t.Run("prefers flag value over stored config", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)
		require.NoError(t, config.SetCurrentMerchantCode("MCONFIG"))

		got, err := runGetMerchantCode(t, []string{"sumup", "--merchant-code", "MFLAG"})
		require.NoError(t, err)
		assert.Equal(t, "MFLAG", got)
	})

	t.Run("falls back to stored config", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)
		require.NoError(t, config.SetCurrentMerchantCode("MCONFIG"))

		got, err := runGetMerchantCode(t, []string{"sumup"})
		require.NoError(t, err)
		assert.Equal(t, "MCONFIG", got)
	})

	t.Run("returns helpful error when no merchant code is available", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)

		got, err := runGetMerchantCode(t, []string{"sumup"})
		require.Error(t, err)
		assert.Empty(t, got)
		assert.ErrorContains(t, err, "merchant code is required")
	})
}

func runGetMerchantCode(t *testing.T, args []string) (string, error) {
	t.Helper()

	var got string
	cmd := &cli.Command{
		Name: "sumup",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "merchant-code"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			var err error
			got, err = app.GetMerchantCode(cmd, "merchant-code")
			return err
		},
	}

	err := cmd.Run(context.Background(), args)
	return got, err
}

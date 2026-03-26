package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup/sumup-cli/internal/config"
)

func TestConfig_Save(t *testing.T) {
	t.Run("persists config that can be loaded back", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)

		cfg := &config.Config{CurrentMerchantCode: "M123"}

		require.NoError(t, cfg.Save())

		loaded, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, cfg.CurrentMerchantCode, loaded.CurrentMerchantCode)
	})
}

func TestLoad(t *testing.T) {
	t.Run("returns empty config when file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tempDir)
		t.Setenv("HOME", tempDir)

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Empty(t, cfg.CurrentMerchantCode)
	})
}

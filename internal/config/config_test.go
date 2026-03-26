package config

import "testing"

func TestConfigSaveAndLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	cfg := &Config{CurrentMerchantCode: "M123"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.CurrentMerchantCode != cfg.CurrentMerchantCode {
		t.Fatalf("Load().CurrentMerchantCode = %q, want %q", loaded.CurrentMerchantCode, cfg.CurrentMerchantCode)
	}
}

func TestLoadMissingConfigReturnsEmptyConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CurrentMerchantCode != "" {
		t.Fatalf("Load().CurrentMerchantCode = %q, want empty", cfg.CurrentMerchantCode)
	}
}

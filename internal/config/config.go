package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the CLI configuration.
type Config struct {
	CurrentMerchantCode string `json:"current_merchant_code,omitempty"`
}

// configDir returns the platform-specific configuration directory.
func configDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
			if baseDir == "" {
				return "", fmt.Errorf("unable to determine config directory")
			}
			baseDir = filepath.Join(baseDir, "AppData", "Roaming")
		}
	case "darwin", "linux":
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("unable to determine home directory: %w", err)
			}
			baseDir = filepath.Join(home, ".config")
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return filepath.Join(baseDir, "sumup"), nil
}

// configPath returns the path to the configuration file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sumup.json"), nil
}

// Load reads the configuration from disk.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// GetCurrentMerchantCode returns the current merchant code from config.
func GetCurrentMerchantCode() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.CurrentMerchantCode, nil
}

// SetCurrentMerchantCode sets the current merchant code in config.
func SetCurrentMerchantCode(merchantCode string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.CurrentMerchantCode = merchantCode
	return cfg.Save()
}

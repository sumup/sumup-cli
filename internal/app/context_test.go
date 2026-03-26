package app

import (
	"context"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/config"
)

func TestNormalizeLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "trim and replace underscore", input: " en_US.UTF-8 ", want: "en-US"},
		{name: "strip modifier", input: "sr_RS@latin", want: "sr-RS"},
		{name: "plain locale", input: "de-DE", want: "de-DE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeLocale(tc.input); got != tc.want {
				t.Fatalf("normalizeLocale(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestGetMerchantCodePrefersFlag(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	if err := config.SetCurrentMerchantCode("MCONFIG"); err != nil {
		t.Fatalf("set current merchant code: %v", err)
	}

	cmd := &cli.Command{
		Name: "sumup",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "merchant-code"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			got, err := GetMerchantCode(cmd, "merchant-code")
			if err != nil {
				t.Fatalf("GetMerchantCode() unexpected error: %v", err)
			}
			if got != "MFLAG" {
				t.Fatalf("GetMerchantCode() = %q, want %q", got, "MFLAG")
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"sumup", "--merchant-code", "MFLAG"}); err != nil {
		t.Fatalf("run command: %v", err)
	}
}

func TestGetMerchantCodeFallsBackToConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", tempDir)
	if err := config.SetCurrentMerchantCode("MCONFIG"); err != nil {
		t.Fatalf("set current merchant code: %v", err)
	}

	cmd := &cli.Command{
		Name: "sumup",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "merchant-code"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			got, err := GetMerchantCode(cmd, "merchant-code")
			if err != nil {
				t.Fatalf("GetMerchantCode() unexpected error: %v", err)
			}
			if got != "MCONFIG" {
				t.Fatalf("GetMerchantCode() = %q, want %q", got, "MCONFIG")
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"sumup"}); err != nil {
		t.Fatalf("run command: %v", err)
	}
}

func TestGetMerchantCodeErrorsWhenUnset(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", tempDir)

	cmd := &cli.Command{
		Name: "sumup",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "merchant-code"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			_, err := GetMerchantCode(cmd, "merchant-code")
			if err == nil {
				t.Fatal("GetMerchantCode() error = nil, want non-nil")
			}
			if !strings.Contains(err.Error(), "merchant code is required") {
				t.Fatalf("GetMerchantCode() error = %q, want merchant code hint", err)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"sumup"}); err != nil {
		t.Fatalf("run command: %v", err)
	}
}

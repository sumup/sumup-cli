package util

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
)

func TestRequireSingleArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{name: "one arg", args: []string{"sumup", "abc"}, want: "abc"},
		{name: "missing arg", args: []string{"sumup"}, wantErr: "argument is required"},
		{name: "too many args", args: []string{"sumup", "abc", "def"}, wantErr: "unexpected extra arguments"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := &cli.Command{
				Name: "sumup",
				Action: func(_ context.Context, cmd *cli.Command) error {
					got, err := RequireSingleArg(cmd, "value")
					if tc.wantErr != "" {
						if err == nil {
							t.Fatalf("RequireSingleArg() error = nil, want %q", tc.wantErr)
						}
						if !strings.Contains(err.Error(), tc.wantErr) {
							t.Fatalf("RequireSingleArg() error = %q, want substring %q", err, tc.wantErr)
						}
						return nil
					}

					if err != nil {
						t.Fatalf("RequireSingleArg() unexpected error: %v", err)
					}
					if got != tc.want {
						t.Fatalf("RequireSingleArg() = %q, want %q", got, tc.want)
					}
					return nil
				},
			}

			if err := cmd.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("run command: %v", err)
			}
		})
	}
}

func TestStringOrDefault(t *testing.T) {
	t.Parallel()

	fallback := "fallback"
	if got := StringOrDefault(nil, fallback); got != fallback {
		t.Fatalf("StringOrDefault(nil) = %q, want %q", got, fallback)
	}

	empty := ""
	if got := StringOrDefault(&empty, fallback); got != fallback {
		t.Fatalf("StringOrDefault(empty) = %q, want %q", got, fallback)
	}

	value := "value"
	if got := StringOrDefault(&value, fallback); got != value {
		t.Fatalf("StringOrDefault(value) = %q, want %q", got, value)
	}
}

func TestBoolLabel(t *testing.T) {
	t.Parallel()

	if got := BoolLabel(nil); got != "-" {
		t.Fatalf("BoolLabel(nil) = %q, want %q", got, "-")
	}

	trueValue := true
	if got := BoolLabel(&trueValue); got != "Yes" {
		t.Fatalf("BoolLabel(true) = %q, want %q", got, "Yes")
	}

	falseValue := false
	if got := BoolLabel(&falseValue); got != "No" {
		t.Fatalf("BoolLabel(false) = %q, want %q", got, "No")
	}
}

func TestTimeOrDash(t *testing.T) {
	t.Parallel()

	if got := TimeOrDash(nil, nil); got != "-" {
		t.Fatalf("TimeOrDash(nil, nil) = %q, want %q", got, "-")
	}

	ts := time.Date(2026, time.March, 26, 12, 34, 56, 0, time.UTC)
	ctx := &app.Context{ExactTimestamps: true}
	if got := TimeOrDash(ctx, &ts); got != ts.In(time.Local).Format(time.RFC3339) {
		t.Fatalf("TimeOrDash(exact) = %q, want %q", got, ts.In(time.Local).Format(time.RFC3339))
	}

	relative := TimeOrDash(&app.Context{Locale: "en"}, &ts)
	if relative == "" || relative == "-" {
		t.Fatalf("TimeOrDash(relative) = %q, want non-empty relative string", relative)
	}
}

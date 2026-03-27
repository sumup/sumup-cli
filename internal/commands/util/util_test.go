package util_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := runRequireSingleArg(t, tt.args)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringOrDefault(t *testing.T) {
	t.Parallel()

	fallback := "fallback"

	t.Run("returns fallback for nil", func(t *testing.T) {
		assert.Equal(t, fallback, util.StringOrDefault(nil, fallback))
	})

	t.Run("returns fallback for empty string", func(t *testing.T) {
		empty := ""
		assert.Equal(t, fallback, util.StringOrDefault(&empty, fallback))
	})

	t.Run("returns provided value when present", func(t *testing.T) {
		value := "value"
		assert.Equal(t, value, util.StringOrDefault(&value, fallback))
	})
}

func TestBoolLabel(t *testing.T) {
	t.Parallel()

	t.Run("returns dash for nil", func(t *testing.T) {
		assert.Equal(t, "-", util.BoolLabel(nil))
	})

	t.Run("returns yes for true", func(t *testing.T) {
		trueValue := true
		assert.Equal(t, "Yes", util.BoolLabel(&trueValue))
	})

	t.Run("returns no for false", func(t *testing.T) {
		falseValue := false
		assert.Equal(t, "No", util.BoolLabel(&falseValue))
	})
}

func TestSliceOrEmpty(t *testing.T) {
	t.Parallel()

	t.Run("returns empty slice for nil pointer", func(t *testing.T) {
		var values *[]string
		assert.Empty(t, util.SliceOrEmpty(values))
	})

	t.Run("returns dereferenced slice when present", func(t *testing.T) {
		values := []string{"a", "b"}
		assert.Equal(t, values, util.SliceOrEmpty(&values))
	})
}

func TestTimeOrDash(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, time.March, 26, 12, 34, 56, 0, time.UTC)

	t.Run("returns dash for nil timestamp", func(t *testing.T) {
		assert.Equal(t, "-", util.TimeOrDash(nil, nil))
	})

	t.Run("formats exact timestamps in local time", func(t *testing.T) {
		ctx := &app.Context{ExactTimestamps: true}
		assert.Equal(t, ts.In(time.Local).Format(time.RFC3339), util.TimeOrDash(ctx, &ts))
	})

	t.Run("formats relative timestamps when exact mode is disabled", func(t *testing.T) {
		got := util.TimeOrDash(&app.Context{Locale: "en"}, &ts)
		assert.NotEmpty(t, got)
		assert.NotEqual(t, "-", got)
	})
}

func runRequireSingleArg(t *testing.T, args []string) (string, error) {
	t.Helper()

	var got string
	cmd := &cli.Command{
		Name: "sumup",
		Action: func(_ context.Context, cmd *cli.Command) error {
			var err error
			got, err = util.RequireSingleArg(cmd, "value")
			return err
		},
	}

	err := cmd.Run(context.Background(), args)
	return got, err
}

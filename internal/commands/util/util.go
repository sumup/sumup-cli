package util

import (
	"fmt"
	"time"

	"github.com/mergestat/timediff"
	"github.com/mergestat/timediff/locale"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/app"
)

func RequireSingleArg(cmd *cli.Command, label string) (string, error) {
	args := cmd.Args()
	if args.Len() == 0 {
		return "", fmt.Errorf("%s argument is required", label)
	}
	if args.Len() > 1 {
		return "", fmt.Errorf("unexpected extra arguments: %v", args.Slice()[1:])
	}
	value := args.Get(0)
	if value == "" {
		return "", fmt.Errorf("%s argument cannot be empty", label)
	}
	return value, nil
}

func StringOrDefault(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func TimeOrDash(appCtx *app.Context, value *time.Time) string {
	if value == nil {
		return "-"
	}

	if appCtx != nil && appCtx.ExactTimestamps {
		return value.In(time.Local).Format(time.RFC3339)
	}

	opts := make([]timediff.TimeDiffOption, 0, 1)
	if appCtx != nil && appCtx.Locale != "" {
		opts = append(opts, timediff.WithLocale(locale.Locale(appCtx.Locale)))
	}
	return timediff.TimeDiff(value.UTC(), opts...)
}

func BoolLabel(value *bool) string {
	if value == nil {
		return "-"
	}
	if *value {
		return "Yes"
	}
	return "No"
}

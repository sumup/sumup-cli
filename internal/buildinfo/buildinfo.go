package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	// Version is the semantic version of the CLI.
	Version = "dev"
	// Commit is the VCS revision used for the build.
	Commit = "unknown"
	// Date is the build timestamp in UTC (RFC3339 format).
	Date = "unknown"
)

func info() (version, commit, date string, dirty bool) {
	version = Version
	commit = Commit
	date = Date

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date, false
	}

	if (version == "dev" || version == "(devel)") && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = bi.Main.Version
	}

	for _, setting := range bi.Settings {
		switch setting.Key {
		case "vcs.revision":
			if commit == "unknown" && setting.Value != "" {
				commit = setting.Value
			}
		case "vcs.time":
			if date == "unknown" && setting.Value != "" {
				date = setting.Value
			}
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}

	return version, commit, date, dirty
}

// Short returns a concise version string suitable for --version output.
func Short() string {
	version, commit, _, dirty := info()

	var parts []string
	parts = append(parts, version)
	if commit != "" && commit != "unknown" {
		shortCommit := commit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		parts = append(parts, shortCommit)
	}
	if dirty {
		parts = append(parts, "dirty")
	}

	return strings.Join(parts, " ")
}

// Long returns full build details.
func Long() string {
	version, commit, date, _ := info()
	return fmt.Sprintf(
		"Version: %s\nCommit: %s\nDate: %s",
		version,
		commit,
		date,
	)
}

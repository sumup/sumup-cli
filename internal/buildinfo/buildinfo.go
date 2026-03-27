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

// Details describes the current CLI build metadata.
type Details struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Dirty   bool   `json:"dirty"`
}

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

// Current returns the current build metadata as a structured value.
func Current() Details {
	version, commit, date, dirty := info()
	return Details{
		Version: version,
		Commit:  commit,
		Date:    date,
		Dirty:   dirty,
	}
}

// Short returns a concise version string suitable for --version output.
func Short() string {
	current := Current()

	var parts []string
	parts = append(parts, current.Version)
	if current.Commit != "" && current.Commit != "unknown" {
		shortCommit := current.Commit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		parts = append(parts, shortCommit)
	}
	if current.Dirty {
		parts = append(parts, "dirty")
	}

	return strings.Join(parts, " ")
}

// Long returns full build details.
func Long() string {
	current := Current()
	return fmt.Sprintf(
		"Version: %s\nCommit: %s\nDate: %s",
		current.Version,
		current.Commit,
		current.Date,
	)
}

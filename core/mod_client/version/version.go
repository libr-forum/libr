package version

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// These variables are placeholders.
// They will be replaced by the Go linker's -X flag in the build script.
var (
	Version   = "v1.0.0-beta" // Represents the semantic version, e.g., "v1.0.0"
	GitCommit = "unknown"     // Represents the git commit hash
	BuildTime = "unknown"     // Represents the build timestamp
)

// Info holds all the version-related information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the full version information.
func Get() Info {
	info := Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: strings.TrimPrefix(runtime.Version(), "go"),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	// Fallback for when -ldflags aren't used (e.g., `go run` or `wails dev`)
	if info.Version == "dev" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			// Get version from Go module information
			if buildInfo.Main.Version != "(devel)" {
				info.Version = buildInfo.Main.Version
			}
		}
	}

	return info
}

// GetVersion is a convenience helper to get just the version string.
// This is what the auto-updater will use for comparison.
func GetVersion() string {
	return Get().Version
}

package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (set by build)
	Version = "dev"
	// GitCommit is the git commit hash (set by build)
	GitCommit = ""
	// BuildDate is when the binary was built (set by build)
	BuildDate = ""
)

// Info holds version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit,omitempty"`
	BuildDate string `json:"buildDate,omitempty"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	if i.GitCommit != "" && i.BuildDate != "" {
		return fmt.Sprintf("%s (commit: %s, built: %s, %s, %s)",
			i.Version, i.GitCommit[:8], i.BuildDate, i.GoVersion, i.Platform)
	}
	return fmt.Sprintf("%s (%s, %s)", i.Version, i.GoVersion, i.Platform)
}

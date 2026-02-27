// Package version holds build metadata injected at compile time via ldflags.
package version

import "runtime"

// These variables are set at build time using:
//
//	go build -ldflags "-X github.com/mstephenholl/gitops-demo/internal/version.Tag=v1.0.0
//	  -X github.com/mstephenholl/gitops-demo/internal/version.Commit=abc1234
//	  -X github.com/mstephenholl/gitops-demo/internal/version.BuildTime=2026-02-26T00:00:00Z"
var (
	Tag       = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Info contains the full build metadata returned by the /info endpoint.
type Info struct {
	Tag       string `json:"tag"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// Get returns the current build metadata.
func Get() Info {
	return Info{
		Tag:       Tag,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
	}
}

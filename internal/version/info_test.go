package version

import (
	"runtime"
	"testing"
)

func TestGet_DefaultValues(t *testing.T) {
	info := Get()

	if info.Tag != Tag {
		t.Errorf("expected Tag %q, got %q", Tag, info.Tag)
	}
	if info.Commit != Commit {
		t.Errorf("expected Commit %q, got %q", Commit, info.Commit)
	}
	if info.BuildTime != BuildTime {
		t.Errorf("expected BuildTime %q, got %q", BuildTime, info.BuildTime)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %q, got %q", runtime.Version(), info.GoVersion)
	}
}

func TestGet_OverriddenValues(t *testing.T) {
	origTag, origCommit, origBuildTime := Tag, Commit, BuildTime
	defer func() {
		Tag, Commit, BuildTime = origTag, origCommit, origBuildTime
	}()

	Tag = "v1.2.3"
	Commit = "abc1234def5678"
	BuildTime = "2026-02-26T12:00:00Z"

	info := Get()

	if info.Tag != "v1.2.3" {
		t.Errorf("expected Tag %q, got %q", "v1.2.3", info.Tag)
	}
	if info.Commit != "abc1234def5678" {
		t.Errorf("expected Commit %q, got %q", "abc1234def5678", info.Commit)
	}
	if info.BuildTime != "2026-02-26T12:00:00Z" {
		t.Errorf("expected BuildTime %q, got %q", "2026-02-26T12:00:00Z", info.BuildTime)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %q, got %q", runtime.Version(), info.GoVersion)
	}
}

func TestGet_GoVersionMatchesRuntime(t *testing.T) {
	info := Get()
	want := runtime.Version()

	if info.GoVersion != want {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, want)
	}
}

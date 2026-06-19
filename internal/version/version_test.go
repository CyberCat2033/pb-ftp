package version

import (
	"testing"
)

func TestCurrent(t *testing.T) {
	oldVersionName := VersionName
	oldVersionCode := VersionCode
	oldBuildTime := BuildTime
	t.Cleanup(func() {
		VersionName = oldVersionName
		VersionCode = oldVersionCode
		BuildTime = oldBuildTime
	})

	VersionName = "9.9.9"
	VersionCode = "99"
	BuildTime = "2026-06-19T12:00:00Z"

	info := Current()

	if info.AppName != AppName {
		t.Fatalf("AppName = %q, want %q", info.AppName, AppName)
	}
	if info.VersionName != "9.9.9" {
		t.Fatalf("VersionName = %q, want 9.9.9", info.VersionName)
	}
	if info.VersionCode != 99 {
		t.Fatalf("VersionCode = %d, want 99", info.VersionCode)
	}
}

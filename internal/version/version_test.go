package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadInstalled(t *testing.T) {
	path := filepath.Join(t.TempDir(), VersionFileName)
	data := []byte(`{"schemaVersion":1,"appName":"pb-ftp","versionName":"1.2.3","versionCode":42,"releasedAt":"2026-06-19T12:00:00Z"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	info, err := ReadInstalled(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.AppName != AppName {
		t.Fatalf("AppName = %q, want %q", info.AppName, AppName)
	}
	if info.VersionName != "1.2.3" {
		t.Fatalf("VersionName = %q, want 1.2.3", info.VersionName)
	}
	if info.VersionCode != 42 {
		t.Fatalf("VersionCode = %d, want 42", info.VersionCode)
	}
}

func TestReadInstalledFallsBackToCurrent(t *testing.T) {
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

	info, err := ReadInstalled(filepath.Join(t.TempDir(), VersionFileName))
	if err == nil {
		t.Fatal("expected missing version file error")
	}

	if info.VersionName != "9.9.9" {
		t.Fatalf("VersionName = %q, want 9.9.9", info.VersionName)
	}
	if info.VersionCode != 99 {
		t.Fatalf("VersionCode = %d, want 99", info.VersionCode)
	}
}

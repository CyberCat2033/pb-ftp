package version

import (
	"strconv"
)

const (
	AppName = "pb-ftp"
)

var (
	VersionName = "dev"
	VersionCode = "0"
	BuildTime   = ""
	BuildID     = ""
)

type Info struct {
	SchemaVersion int    `json:"schemaVersion"`
	AppName       string `json:"appName"`
	VersionName   string `json:"versionName"`
	VersionCode   int64  `json:"versionCode"`
	ReleasedAt    string `json:"releasedAt,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
}

func Current() Info {
	code, err := strconv.ParseInt(VersionCode, 10, 64)
	if err != nil {
		code = 0
	}

	return Info{
		SchemaVersion: 1,
		AppName:       AppName,
		VersionName:   VersionName,
		VersionCode:   code,
		ReleasedAt:    BuildTime,
		BuildID:       BuildID,
	}
}

func DisplayName() string {
	if VersionName == "" {
		return AppName
	}
	return AppName + " " + VersionName
}

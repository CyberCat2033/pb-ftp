package version

import (
	"encoding/json"
	"os"
	"strconv"
)

const (
	AppName                = "pb-ftp"
	VersionFileName        = "pb-ftp.version"
	DefaultVersionFilePath = "/mnt/ext1/applications/" + VersionFileName
)

var (
	VersionName = "dev"
	VersionCode = "0"
	BuildTime   = ""
)

type Info struct {
	SchemaVersion int    `json:"schemaVersion"`
	AppName       string `json:"appName"`
	VersionName   string `json:"versionName"`
	VersionCode   int64  `json:"versionCode"`
	ReleasedAt    string `json:"releasedAt,omitempty"`
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
	}
}

func ReadInstalled(path string) (Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Current(), err
	}

	var info Info
	if err := json.Unmarshal(data, &info); err != nil {
		return Current(), err
	}

	if info.SchemaVersion == 0 {
		info.SchemaVersion = 1
	}
	if info.AppName == "" {
		info.AppName = AppName
	}

	return info, nil
}

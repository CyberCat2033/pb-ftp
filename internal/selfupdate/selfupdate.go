package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pb-ftp/internal/version"
	"regexp"
	"strings"
)

const (
	LauncherPath = "/mnt/ext1/applications/pb-ftp.app"
	BackupPath   = "/mnt/ext1/applications/pb-ftp.app.previous"
	StagingDir   = "/mnt/ext1/applications/.pb-ftp-update"
)

var sha256Pattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

type Request struct {
	SourcePath  string
	VersionName string
	VersionCode int64
	ReleasedAt  string
	BuildID     string
	SHA256      string
}

func Apply(request Request) error {
	if err := validateRequest(request); err != nil {
		return err
	}

	sourceHash, err := fileSHA256(request.SourcePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(sourceHash, request.SHA256) {
		return errors.New("staged launcher checksum mismatch")
	}

	current := version.Current()
	if request.VersionCode < current.VersionCode {
		return fmt.Errorf(
			"refusing downgrade from versionCode %d to %d",
			current.VersionCode,
			request.VersionCode,
		)
	}

	currentHash, currentHashErr := fileSHA256(LauncherPath)
	if currentHashErr == nil && strings.EqualFold(currentHash, sourceHash) {
		_ = os.Remove(request.SourcePath)
		return nil
	}

	if currentHashErr == nil {
		_ = os.Remove(BackupPath)
		if err := os.Rename(LauncherPath, BackupPath); err != nil {
			return fmt.Errorf("backup current launcher: %w", err)
		}
	}

	if err := os.Chmod(request.SourcePath, 0o755); err != nil {
		rollbackBackup()
		return fmt.Errorf("chmod staged launcher: %w", err)
	}
	if err := os.Rename(request.SourcePath, LauncherPath); err != nil {
		rollbackBackup()
		return fmt.Errorf("install staged launcher: %w", err)
	}

	if err := os.Chmod(LauncherPath, 0o755); err != nil {
		return fmt.Errorf("chmod installed launcher: %w", err)
	}

	return nil
}

func validateRequest(request Request) error {
	if strings.TrimSpace(request.VersionName) == "" {
		return errors.New("versionName is required")
	}
	if request.VersionCode <= 0 {
		return errors.New("versionCode must be positive")
	}
	if !sha256Pattern.MatchString(request.SHA256) {
		return errors.New("sha256 is invalid")
	}

	sourcePath := filepath.Clean(request.SourcePath)
	if sourcePath != request.SourcePath || !filepath.IsAbs(sourcePath) {
		return errors.New("sourcePath must be a clean absolute path")
	}
	if !strings.HasPrefix(sourcePath, StagingDir+"/") {
		return errors.New("sourcePath must be inside the update staging directory")
	}
	if filepath.Ext(sourcePath) != ".app" {
		return errors.New("sourcePath must point to a .app launcher")
	}
	if sourcePath == LauncherPath || sourcePath == BackupPath {
		return errors.New("sourcePath cannot point to the active launcher")
	}

	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("sourcePath cannot be a symlink")
	}
	if !info.Mode().IsRegular() {
		return errors.New("sourcePath must point to a regular file")
	}
	return nil
}

func rollbackBackup() {
	if _, err := os.Stat(BackupPath); err == nil {
		_ = os.Remove(LauncherPath)
		_ = os.Rename(BackupPath, LauncherPath)
	}
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

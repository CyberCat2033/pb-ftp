package rescan

import (
	"os/exec"
	"syscall"
)

const (
	defaultLibraryPath = "/mnt/ext1"
	defaultScannerPath = "/ebrmain/bin/scanner.app"
)

func TriggerDefault() error {
	return Trigger(defaultScannerPath, defaultLibraryPath)
}

func Trigger(scannerPath, libraryPath string) error {
	cmd := exec.Command(scannerPath, libraryPath)
	cmd.Dir = "/"
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}

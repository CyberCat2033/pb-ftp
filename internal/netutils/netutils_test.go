package netutils

import (
	"os/exec"
	"testing"
	"time"
)

func TestFTPServerRunningReportsLiveProcess(t *testing.T) {
	server := startTestFTPServer(t, "sleep 1")
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	}()

	if !server.Running() {
		t.Fatal("Running() = false for a live process")
	}
}

func TestFTPServerRunningReportsExitedProcess(t *testing.T) {
	server := startTestFTPServer(t, "exit 0")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !server.Running() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("Running() = true after process exit")
}

func startTestFTPServer(t *testing.T, script string) *FTPServer {
	t.Helper()

	cmd := exec.Command("sh", "-c", script)
	if err := cmd.Start(); err != nil {
		t.Fatalf("cmd.Start() error = %v", err)
	}

	server := &FTPServer{
		cmd:  cmd,
		done: make(chan error, 1),
	}
	go func() {
		server.done <- cmd.Wait()
	}()

	return server
}

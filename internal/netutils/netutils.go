package netutils

import (
	"errors"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	vsftpdPath       = "/mnt/ext1/applications/vsftpd"
	vsftpdConfigPath = "/mnt/ext1/applications/vsftpd.conf"
	ftpStoragePath   = "/mnt/ext1/"
)

type FTPServer struct {
	cmd     *exec.Cmd
	done    chan error
	mu      sync.Mutex
	stopped bool
	exited  bool
	exitErr error
}

func GenerateLink(host string, port string) string {
	u := url.URL{
		Scheme: "ftp",
		User:   url.User("anonymous"),
		Host:   net.JoinHostPort(host, port),
		Path:   ftpStoragePath,
	}
	return u.String()
}

// GetLocalIP attempts to find the local IP natively using net.InterfaceAddrs.
// If it fails or returns no IP, it falls back to the shell-based method.
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ipStr := ipnet.IP.String()
					if !strings.HasPrefix(ipStr, "127.") {
						return ipStr, nil
					}
				}
			}
		}
	}
	return GetLocalIPShell()
}

func GetLocalIPShell() (string, error) {
	cmd := exec.Command("sh", "-c",
		"/sbin/ifconfig | grep 'inet addr:' | grep -v '127.0.0.1' | sed 's/.*addr:\\([0-9.]*\\).*/\\1/'")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func StartVSFTPD() (*FTPServer, error) {
	cmd := exec.Command(vsftpdPath, vsftpdConfigPath)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	server := &FTPServer{
		cmd:  cmd,
		done: make(chan error, 1),
	}
	go func() {
		server.done <- cmd.Wait()
	}()

	return server, nil
}

func (s *FTPServer) Running() bool {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return !s.stopped && !s.exitedLocked()
}

func (s *FTPServer) exitedLocked() bool {
	if s.exited {
		return true
	}

	select {
	case err := <-s.done:
		s.exited = true
		s.exitErr = err
		return true
	default:
		return false
	}
}

func (s *FTPServer) Stop() error {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}
	s.stopped = true

	if s.exitedLocked() {
		return nil
	}

	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	}

	select {
	case err := <-s.done:
		s.exited = true
		s.exitErr = err
		return nil
	case <-time.After(2 * time.Second):
		if err := s.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		s.exitErr = <-s.done
		s.exited = true
		return nil
	}
}

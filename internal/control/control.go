package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"pb-ftp/internal/rescan"
	"pb-ftp/internal/version"
	"sync"
	"time"
)

const minRescanInterval = 5 * time.Second

type Server struct {
	httpServer      *http.Server
	rescanner       *Rescanner
	versionFilePath string
}

type Rescanner struct {
	mu          sync.Mutex
	running     bool
	lastStarted time.Time
}

func Start(address string) (*Server, error) {
	return StartWithVersionFile(address, version.DefaultVersionFilePath)
}

func StartWithVersionFile(address string, versionFilePath string) (*Server, error) {
	rescanner := &Rescanner{}
	mux := http.NewServeMux()
	server := &Server{
		rescanner:       rescanner,
		versionFilePath: versionFilePath,
	}

	mux.HandleFunc("/rescan", rescanner.handleRescan)
	mux.HandleFunc("/version", server.handleVersion)

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = server.httpServer.Serve(listener)
	}()

	return server, nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (r *Rescanner) handleRescan(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status, err := r.Trigger()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintf(w, "{\"status\":\"%s\"}\n", status)
}

func (s *Server) handleVersion(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info, _ := version.ReadInstalled(s.versionFilePath)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func (r *Rescanner) Trigger() (string, error) {
	now := time.Now()

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return "already_running", nil
	}
	if !r.lastStarted.IsZero() && now.Sub(r.lastStarted) < minRescanInterval {
		r.mu.Unlock()
		return "throttled", nil
	}
	r.running = true
	r.lastStarted = now
	r.mu.Unlock()

	err := rescan.TriggerDefault()

	r.mu.Lock()
	r.running = false
	r.mu.Unlock()

	if err != nil {
		return "failed", err
	}
	return "started", nil
}

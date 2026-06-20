package control

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"pb-ftp/internal/rescan"
	"pb-ftp/internal/version"
	"sync"
	"time"
)

const minRescanInterval = 5 * time.Second
const maxUpdateRequestBytes = 64 * 1024

type Server struct {
	httpServer    *http.Server
	rescanner     *Rescanner
	updateHandler UpdateHandler
}

type Rescanner struct {
	mu          sync.Mutex
	running     bool
	lastStarted time.Time
}

type UpdateRequest struct {
	SourcePath  string `json:"sourcePath"`
	VersionName string `json:"versionName"`
	VersionCode int64  `json:"versionCode"`
	ReleasedAt  string `json:"releasedAt"`
	BuildID     string `json:"buildId,omitempty"`
	SHA256      string `json:"sha256"`
}

type UpdateHandler func(UpdateRequest) error

type Option func(*Server)

func WithUpdateHandler(handler UpdateHandler) Option {
	return func(server *Server) {
		server.updateHandler = handler
	}
}

func Start(address string, options ...Option) (*Server, error) {
	rescanner := &Rescanner{}
	mux := http.NewServeMux()
	server := &Server{
		rescanner: rescanner,
	}
	for _, option := range options {
		option(server)
	}

	mux.HandleFunc("/rescan", rescanner.handleRescan)
	mux.HandleFunc("/version", server.handleVersion)
	mux.HandleFunc("/update", server.handleUpdate)

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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(version.Current())
}

func (s *Server) handleUpdate(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.updateHandler == nil {
		http.Error(w, "update handler unavailable", http.StatusNotFound)
		return
	}

	defer request.Body.Close()
	reader := io.LimitReader(request.Body, maxUpdateRequestBytes)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var updateRequest UpdateRequest
	if err := decoder.Decode(&updateRequest); err != nil {
		http.Error(w, "invalid update request", http.StatusBadRequest)
		return
	}
	if err := s.updateHandler(updateRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintln(w, `{"status":"accepted","restartRequired":true}`)
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

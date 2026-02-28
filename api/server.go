// Package api provides the REST API server for the accessibility audit service.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/config"
)

// Server is the REST API server.
type Server struct {
	engine  AuditEngine
	logger  *slog.Logger
	jobs    map[string]*AuditJob
	jobsMu  sync.RWMutex
	port    int
}

// AuditEngine is the interface for running audits.
type AuditEngine interface {
	RunAudit(ctx context.Context, cfg *config.Config) (*audit.AuditResult, error)
}

// AuditJob represents a running or completed audit job.
type AuditJob struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"` // pending, running, completed, failed
	Config    *config.Config      `json:"config"`
	Result    *audit.AuditResult  `json:"result,omitempty"`
	Error     string              `json:"error,omitempty"`
	StartTime time.Time           `json:"startTime"`
	EndTime   time.Time           `json:"endTime,omitempty"`
	Progress  int                 `json:"progress"` // 0-100
}

// NewServer creates a new API server.
func NewServer(engine AuditEngine, logger *slog.Logger, port int) *Server {
	return &Server{
		engine: engine,
		logger: logger,
		jobs:   make(map[string]*AuditJob),
		port:   port,
	}
}

// Start starts the API server.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("POST /api/v1/audits", s.handleCreateAudit)
	mux.HandleFunc("GET /api/v1/audits", s.handleListAudits)
	mux.HandleFunc("GET /api/v1/audits/{id}", s.handleGetAudit)
	mux.HandleFunc("DELETE /api/v1/audits/{id}", s.handleCancelAudit)
	mux.HandleFunc("GET /api/v1/audits/{id}/report", s.handleGetReport)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.withMiddleware(mux),
	}

	s.logger.Info("starting API server", "port", s.port)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Logging
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

// CreateAuditRequest is the request body for creating an audit.
type CreateAuditRequest struct {
	URL     string                 `json:"url"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Journey *config.JourneyRef     `json:"journey,omitempty"`
}

// CreateAuditResponse is the response for creating an audit.
type CreateAuditResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (s *Server) handleCreateAudit(w http.ResponseWriter, r *http.Request) {
	var req CreateAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		s.writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	// Create audit config
	cfg := config.DefaultConfig()
	cfg.URL = req.URL

	if req.Journey != nil {
		cfg.Journey = req.Journey
	}

	// Generate job ID
	jobID := generateJobID()

	// Create job
	job := &AuditJob{
		ID:        jobID,
		Status:    "pending",
		Config:    cfg,
		StartTime: time.Now(),
	}

	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

	// Start audit in background
	go s.runAuditJob(context.Background(), job)

	s.writeJSON(w, http.StatusAccepted, CreateAuditResponse{
		ID:     jobID,
		Status: "pending",
	})
}

func (s *Server) runAuditJob(ctx context.Context, job *AuditJob) {
	s.jobsMu.Lock()
	job.Status = "running"
	s.jobsMu.Unlock()

	result, err := s.engine.RunAudit(ctx, job.Config)

	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	job.EndTime = time.Now()
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
	} else {
		job.Status = "completed"
		job.Result = result
		job.Progress = 100
	}
}

func (s *Server) handleListAudits(w http.ResponseWriter, r *http.Request) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs := make([]*AuditJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs": jobs,
	})
}

func (s *Server) handleGetAudit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.jobsMu.RLock()
	job, ok := s.jobs[id]
	s.jobsMu.RUnlock()

	if !ok {
		s.writeError(w, http.StatusNotFound, "audit not found")
		return
	}

	s.writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleCancelAudit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.jobsMu.Lock()
	job, ok := s.jobs[id]
	if ok && job.Status == "running" {
		job.Status = "cancelled"
	}
	s.jobsMu.Unlock()

	if !ok {
		s.writeError(w, http.StatusNotFound, "audit not found")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	s.jobsMu.RLock()
	job, ok := s.jobs[id]
	s.jobsMu.RUnlock()

	if !ok {
		s.writeError(w, http.StatusNotFound, "audit not found")
		return
	}

	if job.Status != "completed" {
		s.writeError(w, http.StatusBadRequest, "audit not completed")
		return
	}

	switch format {
	case "json":
		s.writeJSON(w, http.StatusOK, job.Result)
	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// TODO: Use report.Writer for HTML
		s.writeJSON(w, http.StatusOK, job.Result)
	default:
		s.writeError(w, http.StatusBadRequest, "unsupported format")
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

func generateJobID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}

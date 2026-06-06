package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/agent-a11y/audit"
	"github.com/plexusone/agent-a11y/config"
)

// mockEngine implements AuditEngine for testing.
type mockEngine struct {
	result *audit.AuditResult
	err    error
}

func (m *mockEngine) RunAudit(ctx context.Context, cfg *config.Config) (*audit.AuditResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := &mockEngine{
		result: &audit.AuditResult{
			ID:          "test-audit-1",
			TargetURL:   "https://example.com",
			WCAGVersion: audit.WCAG22,
			WCAGLevel:   audit.WCAGLevelAA,
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Stats: audit.AuditStats{
				TotalPages:    1,
				TotalFindings: 3,
			},
			Conformance: audit.ConformanceSummary{
				TargetLevel: audit.WCAGLevelAA,
				Version:     "2.2",
				LevelA: audit.LevelConformance{
					Status:      "supports",
					TotalIssues: 0,
				},
				LevelAA: audit.LevelConformance{
					Status:      "partially_supports",
					TotalIssues: 3,
				},
				Criteria: []audit.CriterionResult{
					{
						ID:         "1.1.1",
						Name:       "Non-text Content",
						Level:      "A",
						Status:     "supports",
						IssueCount: 0,
					},
					{
						ID:         "1.4.3",
						Name:       "Contrast (Minimum)",
						Level:      "AA",
						Status:     "partially_supports",
						IssueCount: 3,
						Remarks:    "Some elements have insufficient contrast.",
					},
				},
			},
		},
	}
	return NewServer(engine, logger, 8080)
}

func TestHandleHealth(t *testing.T) {
	server := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("status = %q, want %q", body["status"], "healthy")
	}

	if body["time"] == "" {
		t.Error("expected 'time' field in response")
	}
}

func TestHandleCreateAudit(t *testing.T) {
	server := newTestServer()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       `{"url": "https://example.com"}`,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "missing url",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/audits", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleCreateAudit(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusAccepted {
				var body CreateAuditResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if body.ID == "" {
					t.Error("expected non-empty job ID")
				}
				if body.Status != "pending" {
					t.Errorf("status = %q, want %q", body.Status, "pending")
				}
			}
		})
	}
}

func TestHandleListAudits(t *testing.T) {
	server := newTestServer()

	// Create a job first
	server.jobsMu.Lock()
	server.jobs["test-job-1"] = &AuditJob{
		ID:        "test-job-1",
		Status:    "completed",
		StartTime: time.Now(),
	}
	server.jobsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits", nil)
	w := httptest.NewRecorder()

	server.handleListAudits(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	jobs, ok := body["jobs"].([]interface{})
	if !ok {
		t.Fatal("expected 'jobs' array in response")
	}
	if len(jobs) != 1 {
		t.Errorf("jobs count = %d, want 1", len(jobs))
	}
}

func TestHandleGetAudit(t *testing.T) {
	server := newTestServer()

	// Create a job
	server.jobsMu.Lock()
	server.jobs["test-job-1"] = &AuditJob{
		ID:        "test-job-1",
		Status:    "completed",
		StartTime: time.Now(),
	}
	server.jobsMu.Unlock()

	tests := []struct {
		name       string
		jobID      string
		wantStatus int
	}{
		{
			name:       "existing job",
			jobID:      "test-job-1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing job",
			jobID:      "non-existent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/audits/"+tt.jobID, nil)
			req.SetPathValue("id", tt.jobID)
			w := httptest.NewRecorder()

			server.handleGetAudit(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHandleCancelAudit(t *testing.T) {
	server := newTestServer()

	// Create a running job
	server.jobsMu.Lock()
	server.jobs["running-job"] = &AuditJob{
		ID:        "running-job",
		Status:    "running",
		StartTime: time.Now(),
	}
	server.jobsMu.Unlock()

	tests := []struct {
		name       string
		jobID      string
		wantStatus int
	}{
		{
			name:       "cancel running job",
			jobID:      "running-job",
			wantStatus: http.StatusOK,
		},
		{
			name:       "cancel non-existing job",
			jobID:      "non-existent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/audits/"+tt.jobID, nil)
			req.SetPathValue("id", tt.jobID)
			w := httptest.NewRecorder()

			server.handleCancelAudit(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}

	// Verify the job was cancelled
	server.jobsMu.RLock()
	job := server.jobs["running-job"]
	server.jobsMu.RUnlock()

	if job.Status != "cancelled" {
		t.Errorf("job status = %q, want %q", job.Status, "cancelled")
	}
}

func TestHandleGetReport(t *testing.T) {
	server := newTestServer()

	// Create a completed job with result
	server.jobsMu.Lock()
	server.jobs["completed-job"] = &AuditJob{
		ID:        "completed-job",
		Status:    "completed",
		StartTime: time.Now(),
		Result: &audit.AuditResult{
			TargetURL: "https://example.com",
		},
	}
	server.jobs["pending-job"] = &AuditJob{
		ID:        "pending-job",
		Status:    "pending",
		StartTime: time.Now(),
	}
	server.jobsMu.Unlock()

	tests := []struct {
		name       string
		jobID      string
		format     string
		wantStatus int
	}{
		{
			name:       "json format",
			jobID:      "completed-job",
			format:     "json",
			wantStatus: http.StatusOK,
		},
		{
			name:       "default format",
			jobID:      "completed-job",
			format:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing job",
			jobID:      "non-existent",
			format:     "json",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "pending job",
			jobID:      "pending-job",
			format:     "json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported format",
			jobID:      "completed-job",
			format:     "pdf",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/audits/" + tt.jobID + "/report"
			if tt.format != "" {
				url += "?format=" + tt.format
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", tt.jobID)
			w := httptest.NewRecorder()

			server.handleGetReport(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHandleGetOpenACR(t *testing.T) {
	server := newTestServer()

	// Create a completed job with result
	completedResult := &audit.AuditResult{
		ID:          "test-audit",
		TargetURL:   "https://example.com",
		WCAGVersion: audit.WCAG22,
		WCAGLevel:   audit.WCAGLevelAA,
		StartTime:   time.Now(),
		Conformance: audit.ConformanceSummary{
			TargetLevel: audit.WCAGLevelAA,
			Version:     "2.2",
			Criteria: []audit.CriterionResult{
				{
					ID:     "1.1.1",
					Name:   "Non-text Content",
					Level:  "A",
					Status: "supports",
				},
			},
		},
	}

	server.jobsMu.Lock()
	server.jobs["completed-job"] = &AuditJob{
		ID:        "completed-job",
		Status:    "completed",
		StartTime: time.Now(),
		Result:    completedResult,
	}
	server.jobs["pending-job"] = &AuditJob{
		ID:        "pending-job",
		Status:    "pending",
		StartTime: time.Now(),
	}
	server.jobsMu.Unlock()

	tests := []struct {
		name        string
		jobID       string
		queryParams string
		wantStatus  int
		wantType    string
	}{
		{
			name:       "yaml format (default)",
			jobID:      "completed-job",
			wantStatus: http.StatusOK,
			wantType:   "application/x-yaml",
		},
		{
			name:        "json format",
			jobID:       "completed-job",
			queryParams: "format=json",
			wantStatus:  http.StatusOK,
			wantType:    "application/json",
		},
		{
			name:        "with custom metadata",
			jobID:       "completed-job",
			queryParams: "product_name=MyApp&product_version=1.0.0&author_email=test@example.com",
			wantStatus:  http.StatusOK,
			wantType:    "application/x-yaml",
		},
		{
			name:       "non-existing job",
			jobID:      "non-existent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "pending job",
			jobID:      "pending-job",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/audits/" + tt.jobID + "/openacr"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", tt.jobID)
			w := httptest.NewRecorder()

			server.handleGetOpenACR(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK && tt.wantType != "" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != tt.wantType {
					t.Errorf("Content-Type = %q, want %q", contentType, tt.wantType)
				}
			}
		})
	}
}

func TestHandleGetOpenACR_YAMLContent(t *testing.T) {
	server := newTestServer()

	// Create a completed job
	server.jobsMu.Lock()
	server.jobs["test-job"] = &AuditJob{
		ID:        "test-job",
		Status:    "completed",
		StartTime: time.Now(),
		Result: &audit.AuditResult{
			TargetURL:   "https://example.com",
			WCAGVersion: audit.WCAG22,
			StartTime:   time.Now(),
			Conformance: audit.ConformanceSummary{
				Criteria: []audit.CriterionResult{
					{ID: "1.1.1", Level: "A", Status: "supports"},
				},
			},
		},
	}
	server.jobsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits/test-job/openacr", nil)
	req.SetPathValue("id", "test-job")
	w := httptest.NewRecorder()

	server.handleGetOpenACR(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Check YAML contains expected fields
	content := string(body)
	expectedFields := []string{"title:", "product:", "catalog:", "chapters:"}
	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("YAML output missing field %q", field)
		}
	}
}

func TestHandleGetOpenACR_JSONContent(t *testing.T) {
	server := newTestServer()

	// Create a completed job
	server.jobsMu.Lock()
	server.jobs["test-job"] = &AuditJob{
		ID:        "test-job",
		Status:    "completed",
		StartTime: time.Now(),
		Result: &audit.AuditResult{
			TargetURL:   "https://example.com",
			WCAGVersion: audit.WCAG22,
			StartTime:   time.Now(),
			Conformance: audit.ConformanceSummary{
				Criteria: []audit.CriterionResult{
					{ID: "1.1.1", Level: "A", Status: "supports"},
				},
			},
		},
	}
	server.jobsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits/test-job/openacr?format=json", nil)
	req.SetPathValue("id", "test-job")
	w := httptest.NewRecorder()

	server.handleGetOpenACR(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	// Check expected fields
	if _, ok := result["title"]; !ok {
		t.Error("JSON output missing 'title' field")
	}
	if _, ok := result["product"]; !ok {
		t.Error("JSON output missing 'product' field")
	}
	if _, ok := result["catalog"]; !ok {
		t.Error("JSON output missing 'catalog' field")
	}
}

func TestHandleGetOpenACR_AcceptHeader(t *testing.T) {
	server := newTestServer()

	// Create a completed job
	server.jobsMu.Lock()
	server.jobs["test-job"] = &AuditJob{
		ID:        "test-job",
		Status:    "completed",
		StartTime: time.Now(),
		Result: &audit.AuditResult{
			TargetURL:   "https://example.com",
			WCAGVersion: audit.WCAG22,
			StartTime:   time.Now(),
			Conformance: audit.ConformanceSummary{
				Criteria: []audit.CriterionResult{},
			},
		},
	}
	server.jobsMu.Unlock()

	// Request with Accept: application/json header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audits/test-job/openacr", nil)
	req.SetPathValue("id", "test-job")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	server.handleGetOpenACR(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestMiddleware_CORS(t *testing.T) {
	server := newTestServer()

	handler := server.withMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header Access-Control-Allow-Origin: *")
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS header Access-Control-Allow-Methods")
	}
}

func TestMiddleware_Options(t *testing.T) {
	server := newTestServer()

	handler := server.withMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Should not reach here
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("OPTIONS request status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGenerateJobID(t *testing.T) {
	id1 := generateJobID()
	id2 := generateJobID()

	if id1 == "" {
		t.Error("generateJobID() returned empty string")
	}

	if !strings.HasPrefix(id1, "audit-") {
		t.Errorf("generateJobID() = %q, want prefix 'audit-'", id1)
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("generateJobID() should return unique IDs")
	}
}

func TestWriteJSON(t *testing.T) {
	server := newTestServer()

	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	server.writeJSON(w, http.StatusOK, data)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", resp.Header.Get("Content-Type"), "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want %q", body["key"], "value")
	}
}

func TestWriteError(t *testing.T) {
	server := newTestServer()

	w := httptest.NewRecorder()
	server.writeError(w, http.StatusBadRequest, "test error")

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["error"] != "test error" {
		t.Errorf("body[error] = %q, want %q", body["error"], "test error")
	}
}

func TestNewServer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := &mockEngine{}

	server := NewServer(engine, logger, 8080)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.port != 8080 {
		t.Errorf("port = %d, want 8080", server.port)
	}
	if server.jobs == nil {
		t.Error("jobs map not initialized")
	}
}

func TestRunAuditJob(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	result := &audit.AuditResult{
		ID:        "test-result",
		TargetURL: "https://example.com",
	}
	engine := &mockEngine{result: result}
	server := NewServer(engine, logger, 8080)

	job := &AuditJob{
		ID:        "test-job",
		Status:    "pending",
		Config:    config.DefaultConfig(),
		StartTime: time.Now(),
	}
	server.jobs[job.ID] = job

	// Run the audit job
	server.runAuditJob(context.Background(), job)

	// Verify job status
	if job.Status != "completed" {
		t.Errorf("job.Status = %q, want %q", job.Status, "completed")
	}
	if job.Result == nil {
		t.Error("job.Result is nil")
	}
	if job.Progress != 100 {
		t.Errorf("job.Progress = %d, want 100", job.Progress)
	}
	if job.EndTime.IsZero() {
		t.Error("job.EndTime not set")
	}
}

func TestRunAuditJob_Error(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	engine := &mockEngine{err: bytes.ErrTooLarge} // Use any error
	server := NewServer(engine, logger, 8080)

	job := &AuditJob{
		ID:        "test-job",
		Status:    "pending",
		Config:    config.DefaultConfig(),
		StartTime: time.Now(),
	}
	server.jobs[job.ID] = job

	// Run the audit job
	server.runAuditJob(context.Background(), job)

	// Verify job status
	if job.Status != "failed" {
		t.Errorf("job.Status = %q, want %q", job.Status, "failed")
	}
	if job.Error == "" {
		t.Error("job.Error is empty")
	}
}

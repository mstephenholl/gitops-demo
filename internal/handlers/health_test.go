package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mstephenholl/gitops-demo/internal/version"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io_Discard(), nil))
}

// io_Discard returns a writer that discards all data.
// We use a helper to avoid importing io in every test.
func io_Discard() *discardWriter { return &discardWriter{} }

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestHealthz_ReturnsOK(t *testing.T) {
	handler := Healthz(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", resp.Status)
	}
}

func TestHealthz_ContentType(t *testing.T) {
	handler := Healthz(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

func TestReadyz_ReturnsReady(t *testing.T) {
	handler := Readyz(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ready" {
		t.Errorf("expected status %q, got %q", "ready", resp.Status)
	}
}

func TestInfo_ReturnsBuildMetadata(t *testing.T) {
	// Save and restore originals
	origTag, origCommit, origBuildTime := version.Tag, version.Commit, version.BuildTime
	defer func() {
		version.Tag, version.Commit, version.BuildTime = origTag, origCommit, origBuildTime
	}()

	version.Tag = "v0.1.0-test"
	version.Commit = "deadbeef"
	version.BuildTime = "2026-01-01T00:00:00Z"

	handler := Info(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var info version.Info
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if info.Tag != "v0.1.0-test" {
		t.Errorf("expected Tag %q, got %q", "v0.1.0-test", info.Tag)
	}
	if info.Commit != "deadbeef" {
		t.Errorf("expected Commit %q, got %q", "deadbeef", info.Commit)
	}
	if info.BuildTime != "2026-01-01T00:00:00Z" {
		t.Errorf("expected BuildTime %q, got %q", "2026-01-01T00:00:00Z", info.BuildTime)
	}
	if info.GoVersion == "" {
		t.Error("expected GoVersion to be non-empty")
	}
}

func TestInfo_ContentType(t *testing.T) {
	handler := Info(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

// errorWriter is an http.ResponseWriter whose Write method always fails,
// forcing json.Encoder.Encode to return an error inside writeJSON.
type errorWriter struct {
	header     http.Header
	statusCode int
}

func (e *errorWriter) Header() http.Header       { return e.header }
func (e *errorWriter) WriteHeader(code int)      { e.statusCode = code }
func (e *errorWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("forced write error") }

func TestWriteJSON_EncodeError(t *testing.T) {
	ew := &errorWriter{header: http.Header{}}

	// A broken writer forces json.Encoder.Encode to fail, exercising the
	// error branch inside writeJSON. After the error, http.Error overwrites
	// Content-Type to "text/plain; charset=utf-8".
	writeJSON(ew, http.StatusOK, map[string]string{"key": "value"})

	ct := ew.header.Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type %q after error, got %q", "text/plain; charset=utf-8", ct)
	}
}

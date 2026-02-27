package handlers

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	middleware := RequestLogger(logger)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()

	if !strings.Contains(logOutput, "request completed") {
		t.Error("expected log to contain 'request completed'")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("expected log to contain HTTP method")
	}
	if !strings.Contains(logOutput, "/test-path") {
		t.Error("expected log to contain request path")
	}
}

func TestRequestLogger_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	middleware := RequestLogger(logger)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "404") {
		t.Error("expected log to contain status code 404")
	}
}

func TestRequestLogger_DefaultStatusOK(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	middleware := RequestLogger(logger)
	// Handler that writes body but never calls WriteHeader explicitly.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	handler := middleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/implicit-ok", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "200") {
		t.Errorf("expected log to contain status 200, got: %s", logOutput)
	}
}

func TestResponseRecorder_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rr := &responseRecorder{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
	}

	rr.WriteHeader(http.StatusCreated)

	if rr.statusCode != http.StatusCreated {
		t.Errorf("expected statusCode %d, got %d", http.StatusCreated, rr.statusCode)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected underlying recorder code %d, got %d", http.StatusCreated, rec.Code)
	}
}

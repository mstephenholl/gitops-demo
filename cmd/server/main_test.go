package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mstephenholl/gitops-demo/internal/version"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&discardWriter{}, nil))
}

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestNewRouter_HealthzRoute(t *testing.T) {
	r := newRouter(testLogger())
	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestNewRouter_ReadyzRoute(t *testing.T) {
	r := newRouter(testLogger())
	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestNewRouter_InfoRoute(t *testing.T) {
	r := newRouter(testLogger())
	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/info")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var info version.Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if info.Tag == "" {
		t.Error("expected Tag to be non-empty")
	}
}

func TestNewRouter_NotFound(t *testing.T) {
	r := newRouter(testLogger())
	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestEnvOrDefault_UsesDefault(t *testing.T) {
	val := envOrDefault("GITOPS_DEMO_NONEXISTENT_KEY_12345", "fallback")
	if val != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", val)
	}
}

func TestEnvOrDefault_UsesEnv(t *testing.T) {
	t.Setenv("GITOPS_DEMO_TEST_KEY", "from_env")

	val := envOrDefault("GITOPS_DEMO_TEST_KEY", "fallback")
	if val != "from_env" {
		t.Errorf("expected %q, got %q", "from_env", val)
	}
}

func TestNewServer_Configuration(t *testing.T) {
	handler := http.NewServeMux()
	srv := newServer("9090", handler)

	if srv.Addr != ":9090" {
		t.Errorf("expected Addr %q, got %q", ":9090", srv.Addr)
	}
	if srv.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("expected ReadHeaderTimeout 10s, got %v", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout 30s, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 30*time.Second {
		t.Errorf("expected WriteTimeout 30s, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("expected IdleTimeout 120s, got %v", srv.IdleTimeout)
	}
	if srv.Handler == nil {
		t.Error("expected Handler to be non-nil")
	}
}

func TestNewLogger_ReturnsNonNil(t *testing.T) {
	logger := newLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLogStartup_DoesNotPanic(t *testing.T) {
	logger := testLogger()
	logStartup(logger, "8080")
}

func TestRun_GracefulShutdown(t *testing.T) {
	logger := testLogger()
	srv := newServer("0", newRouter(logger)) // port 0 = random available port

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, srv, logger)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error on graceful shutdown, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run() did not return within timeout")
	}
}

func TestRun_InvalidPort(t *testing.T) {
	logger := testLogger()
	// Bind to a known-used port to force an error.
	// First start a listener so the port is taken.
	blocker := &http.Server{Addr: ":0", Handler: http.NewServeMux(), ReadHeaderTimeout: time.Second}
	go func() { _ = blocker.ListenAndServe() }()
	time.Sleep(20 * time.Millisecond)
	defer func() { _ = blocker.Close() }()

	// Use a port that's definitely invalid
	srv := newServer("99999", newRouter(logger))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := run(ctx, srv, logger)
	if err == nil {
		t.Error("expected an error for invalid port, got nil")
	}
}

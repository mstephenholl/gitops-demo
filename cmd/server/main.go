// Package main starts the gitops-demo HTTP server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/mstephenholl/gitops-demo/internal/handlers"
	"github.com/mstephenholl/gitops-demo/internal/version"
)

func main() {
	if err := start(); err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func start() error {
	// Load .env file if present; existing env vars take precedence.
	_ = godotenv.Load()

	logger := newLogger()
	port := envOrDefault("PORT", "8080")

	logStartup(logger, port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := newServer(port, newRouter(logger))

	return run(ctx, srv, logger)
}

// newLogger creates the default JSON logger for the application.
func newLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}

// logStartup logs the server configuration at startup.
func logStartup(logger *slog.Logger, port string) {
	info := version.Get()
	logger.Info("starting server",
		slog.String("port", port),
		slog.String("tag", info.Tag),
		slog.String("commit", info.Commit),
		slog.String("build_time", info.BuildTime),
	)
}

// newServer creates a configured *http.Server.
func newServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// newRouter builds and returns the Chi router with all routes and middleware.
func newRouter(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(handlers.RequestLogger(logger))

	r.Get("/healthz", handlers.Healthz(logger))
	r.Get("/readyz", handlers.Readyz(logger))
	r.Get("/info", handlers.Info(logger))

	return r
}

// run starts the HTTP server and performs graceful shutdown when ctx is cancelled.
// It returns nil on clean shutdown, or an error if shutdown fails.
func run(ctx context.Context, srv *http.Server, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err, ok := <-errCh:
		if ok && err != nil {
			return fmt.Errorf("server listen: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

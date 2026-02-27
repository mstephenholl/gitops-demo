// Package handlers provides HTTP handlers for the demo service.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mstephenholl/gitops-demo/internal/version"
)

// HealthResponse is the JSON body returned by the health and readiness probes.
type HealthResponse struct {
	Status string `json:"status"`
}

// Healthz returns an HTTP 200 with status "ok". Used as a liveness probe.
func Healthz(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("liveness probe hit")
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	}
}

// Readyz returns an HTTP 200 with status "ready". Used as a readiness probe.
func Readyz(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("readiness probe hit")
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ready"})
	}
}

// Info returns build metadata injected at compile time.
func Info(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := version.Get()
		logger.Info("info endpoint hit",
			slog.String("tag", info.Tag),
			slog.String("commit", info.Commit),
		)
		writeJSON(w, http.StatusOK, info)
	}
}

// writeJSON marshals v to JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}

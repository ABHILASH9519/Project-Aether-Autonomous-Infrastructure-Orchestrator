// Package httpapi defines the Phase 1 control API endpoints.
// All responses are informational; no remediation operations are exposed.
package httpapi

import (
    "encoding/json"
    "net/http"
)

type response struct {
    Status  string `json:"status"`
    Service string `json:"service"`
    Version string `json:"version,omitempty"`
}

func jsonResponse(w http.ResponseWriter, status int, data response) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Cache-Control", "no-store")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(data)
}

// NewHandler creates an HTTP handler without requiring cloud services.
func NewHandler() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        jsonResponse(w, http.StatusOK, response{Status: "ok", Service: "control-api"})
    })
    mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
        // Readiness covers only current Phase 1 dependencies (none).
        jsonResponse(w, http.StatusOK, response{Status: "ready", Service: "control-api"})
    })
    mux.HandleFunc("GET /api/v1/info", func(w http.ResponseWriter, r *http.Request) {
        jsonResponse(w, http.StatusOK, response{Status: "experimental", Service: "control-api", Version: "0.1.0"})
    })
    return mux
}

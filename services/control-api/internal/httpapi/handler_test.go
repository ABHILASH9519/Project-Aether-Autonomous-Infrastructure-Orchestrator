package httpapi

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestEndpoints(t *testing.T) {
    for _, tc := range []struct{ path, want string }{
        {"/healthz", "ok"},
        {"/readyz", "ready"},
        {"/api/v1/info", "experimental"},
    } {
        t.Run(tc.path, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, tc.path, nil)
            rec := httptest.NewRecorder()
            NewHandler().ServeHTTP(rec, req)
            if rec.Code != http.StatusOK {
                t.Fatalf("status = %d; want 200", rec.Code)
            }
            if got := rec.Header().Get("Content-Type"); got != "application/json" {
                t.Fatalf("content type = %q", got)
            }
            var res response
            if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
                t.Fatalf("decode: %v", err)
            }
            if res.Status != tc.want {
                t.Fatalf("status = %q; want %q", res.Status, tc.want)
            }
        })
    }
}

func TestOtherMethodIsRejected(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
    rec := httptest.NewRecorder()
    NewHandler().ServeHTTP(rec, req)
    if rec.Code != http.StatusMethodNotAllowed {
        t.Fatalf("status = %d; want 405", rec.Code)
    }
}

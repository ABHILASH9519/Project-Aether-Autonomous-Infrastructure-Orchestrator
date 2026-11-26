package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics(t *testing.T) {
	m := &HTTP{}
	h := m.Wrap(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(503)}))
	h.ServeHTTP(httptest.NewRecorder(),httptest.NewRequest("GET","/private/id-123",nil))
	w := httptest.NewRecorder()
	h.ServeHTTP(w,httptest.NewRequest("GET","/metrics",nil))
	body:=w.Body.String()
	if !strings.Contains(body,`aether_http_requests_total{status_class="5xx"} 1`) || !strings.Contains(body,"aether_http_inflight_requests 0") || strings.Contains(body,"id-123") {
		t.Fatal(body)
	}
}

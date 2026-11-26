// Package metrics exposes bounded-cardinality HTTP metrics in Prometheus text format.
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type counter struct {
	requests atomic.Uint64
	micros atomic.Uint64
}

type HTTP struct {
	classes [6]counter
	inflight atomic.Int64
}

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	if r.status != 0 { return }
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
func (r *recorder) Write(b []byte) (int,error) {
	if r.status == 0 { r.WriteHeader(http.StatusOK) }
	return r.ResponseWriter.Write(b)
}

// Wrap instruments non-streaming handlers. Paths and user data never become labels.
func (m *HTTP) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if r.URL.Path == "/metrics" && r.Method == http.MethodGet {
			m.serve(w)
			return
		}
		start := time.Now()
		m.inflight.Add(1)
		defer m.inflight.Add(-1)
		rec := &recorder{ResponseWriter:w}
		next.ServeHTTP(rec,r)
		if rec.status == 0 { rec.status = http.StatusOK }
		class := rec.status/100
		if class < 1 || class > 5 { class = 0 }
		m.classes[class].requests.Add(1)
		m.classes[class].micros.Add(uint64(time.Since(start).Microseconds()))
	})
}

func (m *HTTP) serve(w http.ResponseWriter) {
	w.Header().Set("Content-Type","text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control","no-store")
	fmt.Fprintln(w,"# HELP aether_http_requests_total Completed HTTP requests.")
	fmt.Fprintln(w,"# TYPE aether_http_requests_total counter")
	for i:=1;i<=5;i++ {
		fmt.Fprintf(w,"aether_http_requests_total{status_class=\"%dxx\"} %d\n",i,m.classes[i].requests.Load())
	}
	fmt.Fprintln(w,"# HELP aether_http_request_duration_seconds_total Cumulative request duration in seconds.")
	fmt.Fprintln(w,"# TYPE aether_http_request_duration_seconds_total counter")
	for i:=1;i<=5;i++ {
		fmt.Fprintf(w,"aether_http_request_duration_seconds_total{status_class=\"%dxx\"} %.6f\n",i,float64(m.classes[i].micros.Load())/1e6)
	}
	fmt.Fprintln(w,"# HELP aether_http_inflight_requests Active HTTP requests.")
	fmt.Fprintln(w,"# TYPE aether_http_inflight_requests gauge")
	fmt.Fprintf(w,"aether_http_inflight_requests %d\n",m.inflight.Load())
}

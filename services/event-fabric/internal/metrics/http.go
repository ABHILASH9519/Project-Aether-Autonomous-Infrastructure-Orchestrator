// Package metrics exposes bounded-cardinality HTTP metrics in Prometheus text format.
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"sync"
	"time"
)

var latencyBounds = [...]uint64{1000,5000,10000,25000,50000,100000,250000,500000,1000000,2500000,5000000,10000000}

type counter struct {
	mu sync.Mutex
	requests uint64
	micros uint64
	buckets [len(latencyBounds)]uint64
}
func (c *counter) observe(us uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests++
	c.micros+=us
	for i,b:=range latencyBounds{if us<=b{c.buckets[i]++}}
}
func (c *counter) snapshot()(uint64,uint64,[len(latencyBounds)]uint64){
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.requests,c.micros,c.buckets
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
		m.classes[class].observe(uint64(time.Since(start).Microseconds()))
	})
}

func (m *HTTP) serve(w http.ResponseWriter) {
	w.Header().Set("Content-Type","text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control","no-store")
	var counts [6]uint64
	var sums [6]uint64
	var buckets [6][len(latencyBounds)]uint64
	for i:=1;i<=5;i++{counts[i],sums[i],buckets[i]=m.classes[i].snapshot()}
	fmt.Fprintln(w,"# HELP aether_http_requests_total Completed HTTP requests.")
	fmt.Fprintln(w,"# TYPE aether_http_requests_total counter")
	for i:=1;i<=5;i++ {
		fmt.Fprintf(w,"aether_http_requests_total{status_class=\"%dxx\"} %d\n",i,counts[i])
	}
	fmt.Fprintln(w,"# HELP aether_http_request_duration_seconds_total Cumulative request duration in seconds.")
	fmt.Fprintln(w,"# TYPE aether_http_request_duration_seconds_total counter")
	for i:=1;i<=5;i++ {
		fmt.Fprintf(w,"aether_http_request_duration_seconds_total{status_class=\"%dxx\"} %.6f\n",i,float64(sums[i])/1e6)
	}
	fmt.Fprintln(w,"# HELP aether_http_request_duration_seconds HTTP request duration histogram.")
	fmt.Fprintln(w,"# TYPE aether_http_request_duration_seconds histogram")
	for i:=1;i<=5;i++{
		for j,b:=range latencyBounds{
			fmt.Fprintf(w,"aether_http_request_duration_seconds_bucket{status_class=\"%dxx\",le=\"%g\"} %d\n",i,float64(b)/1e6,buckets[i][j])
		}
		fmt.Fprintf(w,"aether_http_request_duration_seconds_bucket{status_class=\"%dxx\",le=\"+Inf\"} %d\n",i,counts[i])
		fmt.Fprintf(w,"aether_http_request_duration_seconds_sum{status_class=\"%dxx\"} %.6f\n",i,float64(sums[i])/1e6)
		fmt.Fprintf(w,"aether_http_request_duration_seconds_count{status_class=\"%dxx\"} %d\n",i,counts[i])
	}
	fmt.Fprintln(w,"# HELP aether_http_inflight_requests Active HTTP requests.")
	fmt.Fprintln(w,"# TYPE aether_http_inflight_requests gauge")
	fmt.Fprintf(w,"aether_http_inflight_requests %d\n",m.inflight.Load())
}

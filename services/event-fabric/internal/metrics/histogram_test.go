package metrics

import (
 "net/http/httptest"
 "strings"
 "sync"
 "testing"
)

func TestHistogramConcurrentSnapshot(t *testing.T){
 m:=&HTTP{}
 var wg sync.WaitGroup
 for i:=0;i<100;i++{
  wg.Add(1)
  go func(){defer wg.Done();m.classes[2].observe(5000)}()
 }
 wg.Wait()
 count,sum,buckets:=m.classes[2].snapshot()
 if count!=100 || sum!=500000 || buckets[0]!=0 || buckets[1]!=100 || buckets[len(buckets)-1]!=100{t.Fatalf("count=%d sum=%d buckets=%v",count,sum,buckets)}
 w:=httptest.NewRecorder()
 m.serve(w)
 for _,want:=range []string{
  `aether_http_request_duration_seconds_bucket{status_class="2xx",le="+Inf"} 100`,
  `aether_http_request_duration_seconds_count{status_class="2xx"} 100`,
  `aether_http_request_duration_seconds_sum{status_class="2xx"} 0.500000`,
 }{if !strings.Contains(w.Body.String(),want){t.Fatal(want)}}
}

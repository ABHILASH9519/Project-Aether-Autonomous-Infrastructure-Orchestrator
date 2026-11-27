// Command loadtest exercises only the local AETHER demo APIs.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type sample struct { status int; latency time.Duration; err error }
type report struct {
	Requests int `json:"requests"`
	Accepted int `json:"accepted"`
	Other int `json:"other"`
	Errors int `json:"transport_errors"`
	Projected int `json:"projected_events"`
	P50MS float64 `json:"p50_ms"`
	P95MS float64 `json:"p95_ms"`
	P99MS float64 `json:"p99_ms"`
	WallMS float64 `json:"wall_ms"`
}

// percentile uses nearest rank on a sorted, nonempty sample.
func percentile(sorted []time.Duration,percent int)float64{
	rank:=(len(sorted)*percent+99)/100
	if rank<1{rank=1}
	return float64(sorted[rank-1])/float64(time.Millisecond)
}

func run(ctx context.Context,n,workers int,token string)(report,error){
	var out report
	if n<1 || n>100 || workers<1 || workers>16 || workers>n || token==""{return out,errors.New("requests 1..100, concurrency 1..16 and operator token required")}
	var nonce [12]byte
	if _,err:=rand.Read(nonce[:]);err!=nil{return out,err}
	id:="load-"+hex.EncodeToString(nonce[:])
	client:=&http.Client{Timeout:8*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
	jobs:=make(chan int,n)
	results:=make(chan sample,n)
	for i:=0;i<n;i++{jobs<-i}
	close(jobs)
	start:=time.Now()
	var wg sync.WaitGroup
	for i:=0;i<workers;i++{
		wg.Add(1)
		go func(){
			defer wg.Done()
			for seq:=range jobs{
				payload,_:=json.Marshal(map[string]any{"event_id":fmt.Sprintf("%s-%d",id,seq),"event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":time.Now().UTC(),"source":"loadtest","correlation_id":id,"payload":map[string]any{}})
				req,err:=http.NewRequestWithContext(ctx,http.MethodPost,"http://127.0.0.1:8081/api/v1/events",bytes.NewReader(payload))
				if err!=nil{results<-sample{err:err};continue}
				req.Header.Set("Content-Type","application/json")
				t0:=time.Now()
				resp,err:=client.Do(req)
				if err!=nil{results<-sample{err:err};continue}
				_,readErr:=io.Copy(io.Discard,io.LimitReader(resp.Body,4096))
				resp.Body.Close()
				results<-sample{status:resp.StatusCode,latency:time.Since(t0),err:readErr}
			}
		}()
	}
	wg.Wait()
	close(results)
	out.Requests=n
	out.WallMS=float64(time.Since(start))/float64(time.Millisecond)
	var times []time.Duration
	for s:=range results{
		if s.err!=nil{out.Errors++;continue}
		if s.status==http.StatusAccepted{out.Accepted++;times=append(times,s.latency)}else{out.Other++}
	}
	if len(times)>0{
		sort.Slice(times,func(i,j int)bool{return times[i]<times[j]})
		out.P50MS=percentile(times,50)
		out.P95MS=percentile(times,95)
		out.P99MS=percentile(times,99)
	}
	if out.Accepted!=n{return out,errors.New("not all requests accepted; check local readiness and Redis budget")}
	for {
		req,err:=http.NewRequestWithContext(ctx,http.MethodGet,"http://127.0.0.1:8082/api/v1/incidents/"+id,nil)
		if err!=nil{return out,err}
		req.Header.Set("Authorization","Bearer "+token)
		resp,err:=client.Do(req)
		if err==nil{
			var incident struct{EventCount int `json:"event_count"`}
			if resp.StatusCode==200 && json.NewDecoder(io.LimitReader(resp.Body,4096)).Decode(&incident)==nil{out.Projected=incident.EventCount}
			resp.Body.Close()
			if out.Projected==n{return out,nil}
			if out.Projected>n{return out,errors.New("duplicate projection effects")}
		}
		select{case <-ctx.Done():return out,ctx.Err();case <-time.After(500*time.Millisecond):}
	}
}

func main(){
	n:=flag.Int("requests",40,"number of local synthetic events (max 100)")
	workers:=flag.Int("concurrency",4,"parallel requests (max 16)")
	flag.Parse()
	ctx,cancel:=context.WithTimeout(context.Background(),90*time.Second)
	defer cancel()
	out,err:=run(ctx,*n,*workers,os.Getenv("AETHER_OPERATOR_TOKEN"))
	_ = json.NewEncoder(os.Stdout).Encode(out)
	if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
}

// Package admission bounds telemetry ingestion across API replicas.
package admission

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type Gate interface {
	Allow(context.Context) (bool,error)
}

type RedisGate struct {
	Client *redis.Client
	Limit int64
}

// A single Lua operation prevents a counter without a corresponding expiry.
// A shared, fixed-window budget is intentional; this is not per-tenant fairness.
var script = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return n
`)

func (g RedisGate) Allow(ctx context.Context) (bool,error) {
	if g.Limit < 1 { return false,errors.New("invalid admission budget") }
	n,err:=script.Run(ctx,g.Client,[]string{"aether:ingest:budget:v1"},60000).Int64()
	return n<=g.Limit,err
}

func Wrap(next http.Handler,g Gate) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.Method!="POST" || r.URL.Path!="/api/v1/events" {
			next.ServeHTTP(w,r)
			return
		}
		ctx,cancel:=context.WithTimeout(r.Context(),time.Second)
		defer cancel()
		ok,err:=g.Allow(ctx)
		if err!=nil{
			http.Error(w,"admission unavailable",http.StatusServiceUnavailable)
			return
		}
		if !ok{
			w.Header().Set("Retry-After","60")
			http.Error(w,"ingestion budget exhausted",http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w,r)
	})
}

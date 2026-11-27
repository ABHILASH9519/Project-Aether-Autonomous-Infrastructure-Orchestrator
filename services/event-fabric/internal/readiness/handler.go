// Package readiness separates process liveness from dependency readiness.
package readiness

import (
	"context"
	"net/http"
	"time"
)

type Check func(context.Context) error

// Wrap exposes a bounded dependency check, without disclosing backend errors.
func Wrap(next http.Handler, checks ...Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.Method!="GET" || r.URL.Path!="/readyz"{next.ServeHTTP(w,r);return}
		w.Header().Set("Cache-Control","no-store")
		ctx,cancel:=context.WithTimeout(r.Context(),2500*time.Millisecond)
		defer cancel()
		for _,check:=range checks{
			if err:=check(ctx);err!=nil{http.Error(w,"dependencies not ready",503);return}
		}
		if ctx.Err()!=nil{http.Error(w,"dependencies not ready",503);return}
		w.WriteHeader(200)
	})
}

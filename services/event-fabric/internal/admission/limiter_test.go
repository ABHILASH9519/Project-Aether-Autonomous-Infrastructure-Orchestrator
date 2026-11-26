package admission

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeGate struct { allowed bool; err error }
func (f fakeGate) Allow(context.Context)(bool,error){return f.allowed,f.err}

func TestAdmission(t *testing.T) {
	for _,tc:=range []struct{ path,method string; gate fakeGate; want int }{
		{"/api/v1/events","POST",fakeGate{allowed:true},202},
		{"/api/v1/events","POST",fakeGate{},429},
		{"/api/v1/events","POST",fakeGate{err:errors.New("redis unavailable")},503},
		{"/healthz","GET",fakeGate{err:errors.New("redis unavailable")},202},
	}{
		h:=Wrap(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(202)}),tc.gate)
		w:=httptest.NewRecorder()
		h.ServeHTTP(w,httptest.NewRequest(tc.method,tc.path,nil))
		if w.Code!=tc.want {t.Fatalf("%s: got %d want %d",tc.path,w.Code,tc.want)}
		if tc.want==429 && w.Header().Get("Retry-After")!="60" {t.Fatal("missing retry hint")}
	}
}

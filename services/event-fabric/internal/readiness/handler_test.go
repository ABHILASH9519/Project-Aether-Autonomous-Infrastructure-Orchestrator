package readiness

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessDoesNotChangeLiveness(t *testing.T){
	h:=Wrap(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(200)}),func(context.Context)error{return errors.New("private backend address")})
	for _,tc:=range []struct{path string; code int}{{"/readyz",503},{"/healthz",200}}{
		w:=httptest.NewRecorder()
		h.ServeHTTP(w,httptest.NewRequest("GET",tc.path,nil))
		if w.Code!=tc.code || strings.Contains(w.Body.String(),"private"){t.Fatalf("%s status=%d body=%s",tc.path,w.Code,w.Body.String())}
	}
}

func TestReadinessChecksDependencies(t *testing.T){
	called:=0
	check:=func(ctx context.Context)error{
		called++
		if _,ok:=ctx.Deadline();!ok{t.Error("check missing deadline")}
		return nil
	}
	h:=Wrap(http.NotFoundHandler(),check,check)
	w:=httptest.NewRecorder()
	h.ServeHTTP(w,httptest.NewRequest("GET","/readyz",nil))
	if w.Code!=200 || called!=2{t.Fatalf("status=%d called=%d",w.Code,called)}
}

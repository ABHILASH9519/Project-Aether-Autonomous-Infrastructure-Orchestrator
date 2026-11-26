package ingest

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"aether/event-fabric/internal/event"
)

type fake struct { calls int; err error }
func (f *fake) Publish(_ context.Context,_ event.Envelope) error {f.calls++;return f.err}

const valid = `{"event_id":"e1","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"test","correlation_id":"i1","payload":{}}`

func TestHandler(t *testing.T){
	for _,tc:=range []struct{body string; err error; code,calls int}{
		{valid,nil,202,1},
		{valid,errors.New("offline"),503,1},
		{"{}",nil,400,0},
		{valid+" {}",nil,400,0},
		{strings.Replace(valid,`"payload":{}`,`"payload":null`,1),nil,400,0},
	}{
		f:=&fake{err:tc.err}
		w:=httptest.NewRecorder()
		Handler(f).ServeHTTP(w,httptest.NewRequest("POST","/api/v1/events",strings.NewReader(tc.body)))
		if w.Code!=tc.code || f.calls!=tc.calls {t.Fatalf("got %d/%d want %d/%d",w.Code,f.calls,tc.code,tc.calls)}
	}
}

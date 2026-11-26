package ingest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"aether/event-fabric/internal/event"
)

type Publisher interface {
	Publish(context.Context,event.Envelope) error
}

func Handler(p Publisher) http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusOK)})
	mux.HandleFunc("POST /api/v1/events",func(w http.ResponseWriter,r *http.Request){
		r.Body=http.MaxBytesReader(w,r.Body,1<<20)
		d:=json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		var e event.Envelope
		if d.Decode(&e)!=nil || e.Validate()!=nil || d.Decode(new(any))!=io.EOF {
			http.Error(w,"invalid event",http.StatusBadRequest)
			return
		}
		ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second)
		defer cancel()
		if p.Publish(ctx,e)!=nil {
			http.Error(w,"event publication failed; retry with same event_id",http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	return mux
}

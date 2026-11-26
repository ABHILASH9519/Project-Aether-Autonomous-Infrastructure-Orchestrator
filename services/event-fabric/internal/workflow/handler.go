package workflow

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"aether/event-fabric/internal/lifecycle"
	"aether/event-fabric/internal/store"
)

type Repository interface {
	Incident(context.Context,string)(store.Incident,error)
	Transition(context.Context,string,int64,string)(store.Incident,error)
}

func Handler(repo Repository, token string) (http.Handler,error) {
	if len(token)<24 { return nil,errors.New("operator token must contain at least 24 characters") }
	digest:=sha256.Sum256([]byte(token))
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusOK)})
	mux.HandleFunc("GET /api/v1/incidents/{id}",func(w http.ResponseWriter,r *http.Request){
		i,err:=repo.Incident(r.Context(),r.PathValue("id"))
		respond(w,i,err)
	})
	mux.HandleFunc("POST /api/v1/incidents/{id}/transitions",func(w http.ResponseWriter,r *http.Request){
		r.Body=http.MaxBytesReader(w,r.Body,4096)
		d:=json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		var cmd struct{
			Version *int64 `json:"expected_version"`
			Action string `json:"action"`
		}
		if d.Decode(&cmd)!=nil || d.Decode(new(any))!=io.EOF || cmd.Version==nil || *cmd.Version<0 || cmd.Action=="" {
			http.Error(w,"invalid transition request",http.StatusBadRequest)
			return
		}
		i,err:=repo.Transition(r.Context(),r.PathValue("id"),*cmd.Version,cmd.Action)
		respond(w,i,err)
	})
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		w.Header().Set("Cache-Control","no-store")
		if r.URL.Path!="/healthz" {
			presented,ok:=strings.CutPrefix(r.Header.Get("Authorization"),"Bearer ")
			got:=sha256.Sum256([]byte(presented))
			if !ok || subtle.ConstantTimeCompare(got[:],digest[:])!=1 {
				w.Header().Set("WWW-Authenticate","Bearer")
				http.Error(w,"unauthorized",http.StatusUnauthorized)
				return
			}
		}
		ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second)
		defer cancel()
		mux.ServeHTTP(w,r.WithContext(ctx))
	}),nil
}

func respond(w http.ResponseWriter,i store.Incident,err error){
	if err!=nil{
		switch{
		case errors.Is(err,sql.ErrNoRows):http.Error(w,"incident not found",404)
		case errors.Is(err,store.ErrVersion),errors.Is(err,lifecycle.ErrTransition):http.Error(w,"transition conflict",409)
		default:http.Error(w,"storage unavailable",503)
		}
		return
	}
	w.Header().Set("Content-Type","application/json")
	_ = json.NewEncoder(w).Encode(i)
}

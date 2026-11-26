// Package lab owns a deliberately isolated fixture filesystem.
package lab

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const desired = "aether-lab:healthy\n"

// Handler exposes one fixed, idempotent repair operation. IDs are hashed
// before use as filenames; no caller-supplied paths or contents are accepted.
func Handler(dir string) http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,r *http.Request){w.WriteHeader(200)})
	path:=func(r *http.Request)string{
		sum:=sha256.Sum256([]byte(r.PathValue("id")))
		return filepath.Join(dir,fmt.Sprintf("%x.fixture",sum))
	}
	mux.HandleFunc("PUT /v1/fixtures/{id}",func(w http.ResponseWriter,r *http.Request){
		f,err:=os.CreateTemp(dir,".repair-*")
		if err!=nil{http.Error(w,"fixture unavailable",503);return}
		defer os.Remove(f.Name())
		_,err=f.WriteString(desired)
		closeErr:=f.Close()
		if err!=nil || closeErr!=nil || os.Rename(f.Name(),path(r))!=nil{http.Error(w,"repair failed",503);return}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/fixtures/{id}",func(w http.ResponseWriter,r *http.Request){
		b,err:=os.ReadFile(path(r))
		if err!=nil || string(b)!=desired{http.Error(w,"fixture degraded",503);return}
		w.Header().Set("Content-Type","text/plain")
		_,_=w.Write(b)
	})
	return mux
}

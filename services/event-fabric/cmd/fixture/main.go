package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"aether/event-fabric/internal/lab"
)

func main(){
	dir:=os.Getenv("FIXTURE_DIR")
	if dir==""{log.Fatal("FIXTURE_DIR required")}
	if err:=os.MkdirAll(dir,0700);err!=nil{log.Fatal(err)}
	handler:=lab.Handler(dir)
	timeout:=5*time.Second
	if os.Getenv("AETHER_CONFIG_RECONCILE")=="true" {
		handler=lab.HandlerWithConfiguration(dir,lab.ReconcileConfiguration,lab.VerifyConfiguration)
		timeout=50*time.Second
	}
	s:=&http.Server{Addr:":8083",Handler:handler,ReadHeaderTimeout:3*time.Second,ReadTimeout:5*time.Second,WriteTimeout:timeout,IdleTimeout:30*time.Second}
	log.Fatal(s.ListenAndServe())
}

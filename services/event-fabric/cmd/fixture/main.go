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
	s:=&http.Server{Addr:":8083",Handler:lab.Handler(dir),ReadHeaderTimeout:3*time.Second,ReadTimeout:5*time.Second,WriteTimeout:5*time.Second,IdleTimeout:30*time.Second}
	log.Fatal(s.ListenAndServe())
}

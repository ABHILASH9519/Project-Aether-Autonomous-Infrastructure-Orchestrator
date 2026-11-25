package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/services/control-api/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("AETHER control API listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

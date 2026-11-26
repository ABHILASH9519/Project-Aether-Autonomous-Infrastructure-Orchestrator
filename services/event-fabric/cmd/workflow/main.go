package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aether/event-fabric/internal/store"
	"aether/event-fabric/internal/policy"
	"aether/event-fabric/internal/metrics"
	"aether/event-fabric/internal/workflow"
	_ "github.com/lib/pq"
)

func main(){
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	db,err:=sql.Open("postgres",os.Getenv("DATABASE_URL"))
	if err!=nil{log.Fatal(err)}
	defer db.Close()
	db.SetMaxOpenConns(8)
	st:=store.Store{DB:db}
	gate,err:=policy.New(os.Getenv("OPA_URL"))
	if err!=nil{log.Fatal(err)}
	handler,err:=workflow.Handler(st,os.Getenv("AETHER_OPERATOR_TOKEN"),os.Getenv("AETHER_APPROVER_TOKEN"),gate)
	if err!=nil{log.Fatal(err)}
	initCtx,cancel:=context.WithTimeout(ctx,15*time.Second)
	err=st.Migrate(initCtx)
	cancel()
	if err!=nil{log.Fatal(err)}
	srv:=&http.Server{Addr:":8082",Handler:(&metrics.HTTP{}).Wrap(handler),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:60*time.Second}
	go func(){
		<-ctx.Done()
		shutdown,done:=context.WithTimeout(context.Background(),10*time.Second)
		defer done()
		_ = srv.Shutdown(shutdown)
	}()
	if err=srv.ListenAndServe();err!=nil && !errors.Is(err,http.ErrServerClosed){log.Fatal(err)}
}

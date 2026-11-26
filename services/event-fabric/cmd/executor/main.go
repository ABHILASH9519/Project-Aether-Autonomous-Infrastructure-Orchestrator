package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aether/event-fabric/internal/lab"
	"aether/event-fabric/internal/store"
	_ "github.com/lib/pq"
)

func main(){
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	dsn:=os.Getenv("DATABASE_URL")
	if dsn==""{log.Fatal("DATABASE_URL required")}
	db,err:=sql.Open("postgres",dsn)
	if err!=nil{log.Fatal(err)}
	defer db.Close()
	db.SetMaxOpenConns(4)
	st:=store.Store{DB:db}
	a,err:=lab.NewAdapter(os.Getenv("LAB_FIXTURE_URL"))
	if err!=nil{log.Fatal(err)}
	initCtx,cancel:=context.WithTimeout(ctx,15*time.Second)
	err=st.Migrate(initCtx)
	cancel()
	if err!=nil{log.Fatal(err)}
	for ctx.Err()==nil{
		attempt,done:=context.WithTimeout(ctx,10*time.Second)
		ok,err:=st.RunJob(attempt,a.Repair)
		done()
		if err!=nil{log.Printf("job processing failed: %v",err)}
		if ok && err==nil{continue}
		select{case <-ctx.Done():case <-time.After(time.Second):}
	}
}

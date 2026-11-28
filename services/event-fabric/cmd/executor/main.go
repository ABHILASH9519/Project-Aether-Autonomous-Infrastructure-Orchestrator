package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aether/event-fabric/internal/lab"
	"aether/event-fabric/internal/store"
	"aether/event-fabric/internal/telemetry"
	_ "github.com/lib/pq"
)

func main(){
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	shutdown,err:=telemetry.Init(ctx,"aether-executor")
	if err!=nil{log.Fatal(err)}
	defer func(){c,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel();_ = shutdown(c)}()
	dsn:=os.Getenv("DATABASE_URL")
	if dsn==""{log.Fatal("DATABASE_URL required")}
	db,err:=sql.Open("postgres",dsn)
	if err!=nil{log.Fatal(err)}
	defer db.Close()
	db.SetMaxOpenConns(4)
	st:=store.Store{DB:db}
	repairTimeout:=5*time.Second
	clientTimeout:=2*time.Second
	if raw:=os.Getenv("LAB_REPAIR_TIMEOUT");raw!="" {
		repairTimeout,err=time.ParseDuration(raw)
		if err!=nil || repairTimeout<=0 || repairTimeout>45*time.Second{log.Fatal("invalid LAB_REPAIR_TIMEOUT")}
		clientTimeout=repairTimeout
	}
	var config *tls.Config
 if f:=lab.TLSFromEnv();f.Enabled(){config,err=f.Client();if err!=nil{log.Fatal(err)}}
 a,err:=lab.NewAdapterWithTLS(os.Getenv("LAB_FIXTURE_URL"),clientTimeout,config)
	if err!=nil{log.Fatal(err)}
	initCtx,cancel:=context.WithTimeout(ctx,15*time.Second)
	err=st.Migrate(initCtx)
	cancel()
	if err!=nil{log.Fatal(err)}
	for ctx.Err()==nil{
		attempt,done:=context.WithTimeout(ctx,repairTimeout+10*time.Second)
		ok,err:=st.RunJobWithTimeout(attempt,a.Repair,repairTimeout)
		done()
		if err!=nil{log.Printf("job processing failed: %v",err)}
		if ok && err==nil{continue}
		select{case <-ctx.Done():case <-time.After(time.Second):}
	}
}

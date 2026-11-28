package main

import (
 "context"
 "database/sql"
 "log"
 "os"
 "os/signal"
 "syscall"
 "time"

 "aether/event-fabric/internal/cloud"
 "aether/event-fabric/internal/store"
 _ "github.com/lib/pq"
)
func main(){
 ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer stop()
 if os.Getenv("DATABASE_URL")==""{log.Fatal("DATABASE_URL required")}
 db,err:=sql.Open("postgres",os.Getenv("DATABASE_URL"))
 if err!=nil{log.Fatal(err)}
 defer db.Close()
 db.SetMaxOpenConns(4)
 s:=store.Store{DB:db}
 init,cancel:=context.WithTimeout(ctx,15*time.Second)
 err=s.Migrate(init);cancel()
 if err!=nil{log.Fatal(err)}
 adapter:=cloud.NewLocalStack()
 for ctx.Err()==nil{
  attempt,done:=context.WithTimeout(ctx,18*time.Second)
  worked,err:=s.ArchiveOne(attempt,adapter.Publish);done()
  if err!=nil&&ctx.Err()==nil{log.Printf("evidence delivery deferred: %v",err)}
  if worked&&err==nil{continue}
  select{case <-ctx.Done():case <-time.After(time.Second):}
 }
}

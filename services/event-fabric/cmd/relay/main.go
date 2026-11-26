package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"aether/event-fabric/internal/store"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	dsn:=os.Getenv("DATABASE_URL")
	if dsn=="" { log.Fatal("DATABASE_URL required") }
	db,err:=sql.Open("postgres",dsn)
	if err!=nil { log.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(4)
	st:=store.Store{DB:db}
	initCtx,cancel:=context.WithTimeout(ctx,15*time.Second)
	err=st.Migrate(initCtx)
	cancel()
	if err!=nil { log.Fatal(err) }
	brokers:=os.Getenv("KAFKA_BROKERS")
	if brokers=="" { brokers="localhost:9092" }
	w:=&kafka.Writer{Addr:kafka.TCP(strings.Split(brokers,",")...),Topic:"aether.incidents.observed.v1",Balancer:&kafka.Hash{},RequiredAcks:kafka.RequireAll,WriteTimeout:5*time.Second,BatchTimeout:10*time.Millisecond}
	defer w.Close()
	for ctx.Err()==nil {
		attempt,done:=context.WithTimeout(ctx,10*time.Second)
		ok,err:=st.RelayOne(attempt,func(c context.Context,r store.OutboxRecord) error {
			return w.WriteMessages(c,kafka.Message{Key:[]byte(r.CorrelationID),Value:r.Payload,Headers:[]kafka.Header{{Key:"event_id",Value:[]byte(r.EventID)}}})
		})
		done()
		if err!=nil { log.Printf("outbox relay: %v",err) }
		if ok && err==nil { continue }
		delay:=250*time.Millisecond
		if err!=nil { delay=2*time.Second }
		select { case <-ctx.Done(): case <-time.After(delay): }
	}
}

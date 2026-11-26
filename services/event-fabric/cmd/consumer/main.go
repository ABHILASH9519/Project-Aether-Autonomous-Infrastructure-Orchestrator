package main

import (
	"bytes"
	"context"
	"database/sql"
	"aether/event-fabric/internal/store"
	_ "github.com/lib/pq"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"aether/event-fabric/internal/event"
	"aether/event-fabric/internal/telemetry"
	"github.com/segmentio/kafka-go"
)

func main(){
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	shutdown,err:=telemetry.Init(ctx,"aether-consumer")
	if err!=nil{log.Fatal(err)}
	defer func(){c,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel();_ = shutdown(c)}()
	b:=os.Getenv("KAFKA_BROKERS")
	if b==""{b="localhost:9092"}
	brokers:=strings.Split(b,",")
	dsn:=os.Getenv("DATABASE_URL")
	if dsn==""{log.Fatal("DATABASE_URL required")}
	db,err:=sql.Open("postgres",dsn)
	if err!=nil{log.Fatal(err)}
	defer db.Close()
	st:=store.Store{DB:db}
	if err=st.Migrate(ctx);err!=nil{log.Fatal(err)}
	r:=kafka.NewReader(kafka.ReaderConfig{Brokers:brokers,GroupID:"aether-incident-persistent-v1",Topic:"aether.events.v1",StartOffset:kafka.FirstOffset})
	defer r.Close()
	dlq:=&kafka.Writer{Addr:kafka.TCP(brokers...),Topic:"aether.events.dlq.v1",RequiredAcks:kafka.RequireAll,WriteTimeout:5*time.Second}
	defer dlq.Close()
	for ctx.Err()==nil{
		m,err:=r.FetchMessage(ctx)
		if err!=nil{if ctx.Err()==nil{log.Printf("fetch: %v",err);select{case <-ctx.Done():case <-time.After(time.Second):}};continue}
		msgCtx,span:=telemetry.Consumer(ctx,m)
		var e event.Envelope
		d:=json.NewDecoder(bytes.NewReader(m.Value))
		d.DisallowUnknownFields()
		if d.Decode(&e)!=nil || e.Validate()!=nil || d.Decode(new(any))!=io.EOF{
			// Stop on failed DLQ publication; never commit past a failed record.
			if err=dlq.WriteMessages(msgCtx,kafka.Message{Key:m.Key,Value:m.Value});err!=nil{log.Fatalf("DLQ failed: %v",err)}
		}else{
			applied,err:=st.Apply(msgCtx,e)
			if err!=nil{log.Fatalf("database apply failed: %v",err)}
			log.Printf("incident event_id=%q correlation_id=%q type=%q applied=%t",e.ID,e.CorrelationID,e.Type,applied)
		}
		span.End()
		if err=r.CommitMessages(ctx,m);err!=nil{log.Fatalf("commit failed: %v",err)}
	}
}

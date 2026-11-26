package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"aether/event-fabric/internal/event"
	"github.com/segmentio/kafka-go"
)

func main(){
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
	defer stop()
	b:=os.Getenv("KAFKA_BROKERS")
	if b==""{b="localhost:9092"}
	brokers:=strings.Split(b,",")
	r:=kafka.NewReader(kafka.ReaderConfig{Brokers:brokers,GroupID:"aether-incident-preview-v1",Topic:"aether.events.v1",StartOffset:kafka.FirstOffset})
	defer r.Close()
	dlq:=&kafka.Writer{Addr:kafka.TCP(brokers...),Topic:"aether.events.dlq.v1",RequiredAcks:kafka.RequireAll,WriteTimeout:5*time.Second}
	defer dlq.Close()
	for ctx.Err()==nil{
		m,err:=r.FetchMessage(ctx)
		if err!=nil{if ctx.Err()==nil{log.Printf("fetch: %v",err);select{case <-ctx.Done():case <-time.After(time.Second):}};continue}
		var e event.Envelope
		d:=json.NewDecoder(bytes.NewReader(m.Value))
		d.DisallowUnknownFields()
		if d.Decode(&e)!=nil || e.Validate()!=nil || d.Decode(new(any))!=io.EOF{
			// Stop on failed DLQ publication; never commit past a failed record.
			if err=dlq.WriteMessages(ctx,kafka.Message{Key:m.Key,Value:m.Value});err!=nil{log.Printf("DLQ failed: %v",err);return}
		}else{
			log.Printf("incident preview event_id=%q correlation_id=%q type=%q",e.ID,e.CorrelationID,e.Type)
		}
		if err=r.CommitMessages(ctx,m);err!=nil{log.Printf("commit failed: %v",err);return}
	}
}

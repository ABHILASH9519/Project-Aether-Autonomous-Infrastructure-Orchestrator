package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"aether/event-fabric/internal/event"
	"aether/event-fabric/internal/admission"
	"github.com/redis/go-redis/v9"
	"aether/event-fabric/internal/metrics"
	"aether/event-fabric/internal/ingest"
	"github.com/segmentio/kafka-go"
)

type publisher struct{ w *kafka.Writer }
func (p publisher) Publish(ctx context.Context,e event.Envelope) error {
	b,err:=json.Marshal(e)
	if err!=nil{return err}
	return p.w.WriteMessages(ctx,kafka.Message{Key:[]byte(e.CorrelationID),Value:b})
}
func main(){
	brokers:=os.Getenv("KAFKA_BROKERS")
	if brokers==""{brokers="localhost:9092"}
	w:=&kafka.Writer{Addr:kafka.TCP(strings.Split(brokers,",")...),Topic:"aether.events.v1",Balancer:&kafka.Hash{},RequiredAcks:kafka.RequireAll,WriteTimeout:4*time.Second,BatchTimeout:10*time.Millisecond}
	defer w.Close()
	handler:=ingest.Handler(publisher{w})
	if addr:=os.Getenv("REDIS_ADDR"); addr!="" {
		client:=redis.NewClient(&redis.Options{Addr:addr,MaxRetries:-1,DialTimeout:time.Second,ReadTimeout:time.Second,WriteTimeout:time.Second,ContextTimeoutEnabled:true})
		defer client.Close()
		handler=admission.Wrap(handler,admission.RedisGate{Client:client,Limit:120})
	}
	s:=&http.Server{Addr:":8081",Handler:(&metrics.HTTP{}).Wrap(handler),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second}
	log.Fatal(s.ListenAndServe())
}

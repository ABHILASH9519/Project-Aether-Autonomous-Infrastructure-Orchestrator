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
	s:=&http.Server{Addr:":8081",Handler:ingest.Handler(publisher{w}),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second}
	log.Fatal(s.ListenAndServe())
}

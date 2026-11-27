package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"aether/event-fabric/internal/event"
	_ "github.com/lib/pq"
)

func TestConcurrentInboxAndProjection(t *testing.T) {
	dsn := os.Getenv("AETHER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(8)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s := Store{DB: db}
	if err = s.Migrate(ctx); err != nil { t.Fatal(err) }
	id := fmt.Sprintf("concurrent-%d", time.Now().UnixNano())
	const unique = 8
	const copies = 4
	type result struct { applied bool; err error }
	results := make(chan result, unique*copies)
	var wg sync.WaitGroup
	for i := 0; i < unique; i++ {
		e := event.Envelope{ID:fmt.Sprintf("%s-%d",id,i),Type:"Infrastructure.Alert",Version:1,OccurredAt:time.Now(),Source:"concurrency-test",CorrelationID:id,Payload:json.RawMessage("{}")}
		for j := 0; j < copies; j++ {
			wg.Add(1)
			go func(e event.Envelope){
				defer wg.Done()
				ok,err:=s.Apply(ctx,e)
				results<-result{ok,err}
			}(e)
		}
	}
	wg.Wait()
	close(results)
	applied:=0
	for r:=range results{
		if r.err!=nil{t.Fatal(r.err)}
		if r.applied{applied++}
	}
	if applied!=unique{t.Fatalf("applied=%d want %d",applied,unique)}
	i,err:=s.Incident(ctx,id)
	if err!=nil || i.EventCount!=unique{t.Fatalf("incident=%+v err=%v",i,err)}
	var outbox int
	if err=db.QueryRowContext(ctx,"SELECT count(*) FROM incident_outbox WHERE correlation_id=$1",id).Scan(&outbox);err!=nil || outbox!=unique{t.Fatalf("outbox=%d err=%v",outbox,err)}
	// Existing integration tests use a shared disposable database. Leave no
	// pending records for the subsequent relay tests.
	if _,err=db.ExecContext(ctx,"UPDATE incident_outbox SET published_at=now() WHERE correlation_id=$1",id);err!=nil{t.Fatal(err)}
}

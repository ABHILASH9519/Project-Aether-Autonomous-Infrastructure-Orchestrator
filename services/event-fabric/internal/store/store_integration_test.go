package store

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "os"
 "testing"
 "time"

 "aether/event-fabric/internal/event"
 _ "github.com/lib/pq"
)

func TestPersistenceAndRelay(t *testing.T) {
 dsn:=os.Getenv("AETHER_TEST_DATABASE_URL")
 if dsn=="" { t.Skip("requires dedicated disposable PostgreSQL database") }
 db,err:=sql.Open("postgres",dsn)
 if err!=nil { t.Fatal(err) }
 defer db.Close()
 ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)
 defer cancel()
 s:=Store{DB:db}
 if err=s.Migrate(ctx);err!=nil {t.Fatal(err)}
 // Test database must be dedicated and disposable.
 if _,err=db.ExecContext(ctx,"TRUNCATE incident_audit, incident_outbox, incidents, incident_inbox");err!=nil {t.Fatal(err)}
 e:=event.Envelope{ID:"test-1",Type:"Infrastructure.Alert",Version:1,OccurredAt:time.Now(),Source:"test",CorrelationID:"incident-1",Payload:json.RawMessage("{}")}
 for _,want:=range []bool{true,false} {
  got,err:=s.Apply(ctx,e)
  if err!=nil || got!=want {t.Fatalf("apply=%v err=%v want=%v",got,err,want)}
 }
 var count int
 if err=db.QueryRowContext(ctx,"SELECT event_count FROM incidents WHERE correlation_id=$1",e.CorrelationID).Scan(&count);err!=nil || count!=1 {t.Fatalf("count=%d err=%v",count,err)}
 failure:=errors.New("broker unavailable")
 if _,err=s.RelayOne(ctx,func(context.Context,OutboxRecord)error{return failure});!errors.Is(err,failure){t.Fatalf("expected rollback: %v",err)}
 got,err:=s.RelayOne(ctx,func(_ context.Context,r OutboxRecord)error{
  if r.EventID!=e.ID {return errors.New("wrong event")}
  return nil
 })
 if err!=nil || !got {t.Fatalf("relay=%v err=%v",got,err)}
 // State changes use a separate monotonic version.
 for _,step:=range []struct{action string; version int64; state string}{
  {"acknowledge",0,"acknowledged"},
  {"request_approval",1,"awaiting_approval"},
  {"approve",2,"approved"},
  {"simulate",3,"resolved"},
 }{
  role:="operator"
  if step.action=="approve"{role="approver"}
  i,err:=s.Transition(ctx,e.CorrelationID,step.version,step.action,role)
  if err!=nil || i.State!=step.state{t.Fatalf("transition %s: %v %v",step.action,i,err)}
 }
 if _,err=s.Transition(ctx,e.CorrelationID,0,"acknowledge","operator");!errors.Is(err,ErrVersion){t.Fatalf("stale version: %v",err)}
 var audit int
 if err=db.QueryRowContext(ctx,"SELECT count(*) FROM incident_audit WHERE correlation_id=$1",e.CorrelationID).Scan(&audit);err!=nil || audit!=6{t.Fatalf("audit=%d err=%v",audit,err)}
 var approvals int
 if err=db.QueryRowContext(ctx,"SELECT count(*) FROM incident_audit WHERE correlation_id=$1 AND actor_role='approver' AND action='approve'",e.CorrelationID).Scan(&approvals);err!=nil || approvals!=1{t.Fatalf("approvals=%d err=%v",approvals,err)}
 got,err=s.RelayOne(ctx,func(context.Context,OutboxRecord)error{return errors.New("unexpected second publish")})
 if err!=nil || got {t.Fatalf("empty relay=%v err=%v",got,err)}
}

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

func TestArchiveRetryAndStableEvidence(t *testing.T){
 dsn:=os.Getenv("AETHER_TEST_DATABASE_URL")
 if dsn==""{t.Skip("requires disposable PostgreSQL")}
 db,err:=sql.Open("postgres",dsn)
 if err!=nil{t.Fatal(err)}
 defer db.Close()
 ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)
 defer cancel()
 s:=Store{DB:db}
 if err=s.Migrate(ctx);err!=nil{t.Fatal(err)}
 if _,err=db.ExecContext(ctx,"TRUNCATE remediation_jobs, incident_audit, incident_outbox, incidents, incident_inbox");err!=nil{t.Fatal(err)}
 id:="archive-test"
 e:=event.Envelope{ID:id,Type:"Infrastructure.Alert",Version:1,OccurredAt:time.Now(),Source:"test",CorrelationID:id,Payload:json.RawMessage("{}")}
 if _,err=s.Apply(ctx,e);err!=nil{t.Fatal(err)}
 for v,action:=range []string{"acknowledge","request_approval","approve","execute_lab"}{
  role:="operator";if action=="approve"{role="approver"}
  if _,err=s.Transition(ctx,id,int64(v),action,role);err!=nil{t.Fatal(err)}
 }
 never:=func(context.Context,Evidence)error{t.Error("unexpected publication");return nil}
 if ok,err:=s.ArchiveOne(ctx,never);ok||err!=nil{t.Fatalf("pending job eligible: %v %v",ok,err)}
 if ok,err:=s.RunJob(ctx,func(context.Context,string)error{return nil});!ok||err!=nil{t.Fatalf("job: %v %v",ok,err)}
 var first Evidence
 unavailable:=errors.New("archive unavailable")
 if ok,err:=s.ArchiveOne(ctx,func(_ context.Context,e Evidence)error{first=e;return unavailable});!ok||!errors.Is(err,unavailable){t.Fatalf("failure: %v %v",ok,err)}
 if ok,err:=s.ArchiveOne(ctx,never);ok||err!=nil{t.Fatalf("backoff: %v %v",ok,err)}
 if _,err=db.ExecContext(ctx,"UPDATE remediation_jobs SET archive_available_at=now() WHERE correlation_id=$1",id);err!=nil{t.Fatal(err)}
 if ok,err:=s.ArchiveOne(ctx,func(_ context.Context,e Evidence)error{
  if e.Key!=first.Key||e.SHA256!=first.SHA256||string(e.Body)!=string(first.Body){t.Error("retry changed evidence")}
  j,err:=s.Job(ctx,id);if err!=nil{return err}
  want,err:=json.Marshal(j);if err!=nil{return err}
  if string(e.Body)!=string(want)+"\n"{t.Error("archive does not match job API")}
  return nil
 });!ok||err!=nil{t.Fatalf("retry: %v %v",ok,err)}
 var attempts int;var digest string;var archived bool
 if err=db.QueryRowContext(ctx,"SELECT archive_attempts,archive_sha256,archived_at IS NOT NULL FROM remediation_jobs WHERE correlation_id=$1",id).Scan(&attempts,&digest,&archived);err!=nil||attempts!=2||digest!=first.SHA256||!archived{t.Fatalf("archive record: %d %s %v %v",attempts,digest,archived,err)}
 if ok,err:=s.ArchiveOne(ctx,never);ok||err!=nil{t.Fatalf("completed job redelivered: %v %v",ok,err)}
}

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

func TestJobLifecycle(t *testing.T) {
	dsn:=os.Getenv("AETHER_TEST_DATABASE_URL")
	if dsn=="" {t.Skip("requires disposable PostgreSQL")}
	db,err:=sql.Open("postgres",dsn)
	if err!=nil{t.Fatal(err)}
	defer db.Close()
	ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)
	defer cancel()
	s:=Store{DB:db}
	if err=s.Migrate(ctx);err!=nil{t.Fatal(err)}
	if _,err=db.ExecContext(ctx,"TRUNCATE remediation_jobs, incident_audit, incident_outbox, incidents, incident_inbox");err!=nil{t.Fatal(err)}
	for _,fail:=range []bool{false,true}{
		id:="job-success"
		if fail{id="job-failure"}
		e:=event.Envelope{ID:id,Type:"Infrastructure.Alert",Version:1,OccurredAt:time.Now(),Source:"test",CorrelationID:id,Payload:json.RawMessage("{}")}
		if _,err=s.Apply(ctx,e);err!=nil{t.Fatal(err)}
		if _,err=s.Transition(ctx,id,0,"execute_lab","operator");err==nil{t.Fatal("approval bypass")}
		for version,action:=range []string{"acknowledge","request_approval","approve","execute_lab"}{
			role:="operator"
			if action=="approve"{role="approver"}
			if _,err=s.Transition(ctx,id,int64(version),action,role);err!=nil{t.Fatal(err)}
		}
		n:=1
		want,status:="resolved","succeeded"
		if fail{n=3;want,status="failed","failed"}
		for attempt:=0;attempt<n;attempt++{
			// Advance the retry clock in the isolated test database.
			if _,err=db.ExecContext(ctx,"UPDATE remediation_jobs SET available_at=now() WHERE correlation_id=$1",id);err!=nil{t.Fatal(err)}
			ok,err:=s.RunJob(ctx,func(_ context.Context,got string)error{
				if got!=id{t.Errorf("unexpected job %q",got)}
				if fail{return errors.New("adapter unavailable")}
				return nil
			})
			if err!=nil || !ok{t.Fatalf("run=%v err=%v",ok,err)}
		}
		i,err:=s.Incident(ctx,id)
		if err!=nil || i.State!=want{t.Fatalf("incident=%+v err=%v",i,err)}
		var actual string
		var attempts int
		if err=db.QueryRowContext(ctx,"SELECT status,attempts FROM remediation_jobs WHERE correlation_id=$1",id).Scan(&actual,&attempts);err!=nil || actual!=status || attempts!=n{t.Fatalf("status=%s attempts=%d err=%v",actual,attempts,err)}
	}
}

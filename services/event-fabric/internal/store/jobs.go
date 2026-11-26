package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Repair func(context.Context,string) error

// RunJob claims a durable job. The adapter must be idempotent because a crash
// after an external effect and before commit causes a replay.
// The transaction holds a row lock during bounded adapter I/O.
func (s Store) RunJob(ctx context.Context, repair Repair) (bool,error) {
	tx,err:=s.DB.BeginTx(ctx,nil)
	if err!=nil{return false,err}
	defer tx.Rollback()
	var id string
	var attempts int
	err=tx.QueryRowContext(ctx,`SELECT correlation_id, attempts FROM remediation_jobs
WHERE status='pending' AND available_at<=now() ORDER BY available_at,correlation_id
LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&id,&attempts)
	if errors.Is(err,sql.ErrNoRows){return false,nil}
	if err!=nil{return false,err}
	var state string
	var version int64
	if err=tx.QueryRowContext(ctx,"SELECT state,version FROM incidents WHERE correlation_id=$1 FOR UPDATE",id).Scan(&state,&version);err!=nil{return false,err}
	if state!="executing"{return false,errors.New("job incident is not executing")}
	attempts++
	opctx,cancel:=context.WithTimeout(ctx,5*time.Second)
	repairErr:=repair(opctx,id)
	cancel()
	if ctx.Err()!=nil{return false,ctx.Err()}
	status,evidence:="succeeded","fixture content verified"
	if repairErr!=nil {
		status,evidence="pending","adapter failed"
		if attempts>=3{status="failed"}
	}
	if _,err=tx.ExecContext(ctx,`UPDATE remediation_jobs SET status=$2,attempts=$3,evidence=$4,
available_at=now()+($3 * interval '2 seconds'),updated_at=now() WHERE correlation_id=$1`,id,status,attempts,evidence);err!=nil{return false,err}
	if status!="pending" {
		next:=[]string{"verifying","resolved"}
		actions:=[]string{"lab_verified","lab_complete"}
		if status=="failed"{next=[]string{"failed"};actions=[]string{"lab_retry_exhausted"}}
		for n,target:=range next{
			version++
			if _,err=tx.ExecContext(ctx,`INSERT INTO incident_audit
(correlation_id,version,action,previous_state,next_state,actor_role) VALUES($1,$2,$3,$4,$5,'worker')`,id,version,actions[n],state,target);err!=nil{return false,err}
			state=target
		}
		if _,err=tx.ExecContext(ctx,"UPDATE incidents SET state=$2,version=$3,updated_at=now() WHERE correlation_id=$1",id,state,version);err!=nil{return false,err}
	}
	if err=tx.Commit();err!=nil{return false,err}
	return true,nil
}

type Job struct {
	CorrelationID string `json:"correlation_id"`
	Status string `json:"status"`
	Attempts int `json:"attempts"`
	Evidence string `json:"evidence"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s Store) Job(ctx context.Context,id string)(Job,error){
	var j Job
	err:=s.DB.QueryRowContext(ctx,"SELECT correlation_id,status,attempts,evidence,updated_at FROM remediation_jobs WHERE correlation_id=$1",id).Scan(&j.CorrelationID,&j.Status,&j.Attempts,&j.Evidence,&j.UpdatedAt)
	return j,err
}

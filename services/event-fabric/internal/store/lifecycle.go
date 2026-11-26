package store

import (
	"context"
	"database/sql"
	"errors"

	"aether/event-fabric/internal/lifecycle"
)

var ErrVersion = errors.New("incident version conflict")

type Incident struct {
	ID string `json:"correlation_id"`
	State string `json:"state"`
	Version int64 `json:"version"`
	EventCount int64 `json:"event_count"`
}

func (s Store) Incident(ctx context.Context,id string) (Incident,error) {
	var i Incident
	err:=s.DB.QueryRowContext(ctx,"SELECT correlation_id,state,version,event_count FROM incidents WHERE correlation_id=$1",id).Scan(&i.ID,&i.State,&i.Version,&i.EventCount)
	return i,err
}

// Transition serializes commands and rejects stale operator versions.
// All simulation steps and audit entries commit together.
func (s Store) Transition(ctx context.Context,id string,expected int64,action string) (Incident,error) {
	tx,err:=s.DB.BeginTx(ctx,nil)
	if err!=nil{return Incident{},err}
	defer tx.Rollback()
	var i Incident
	err=tx.QueryRowContext(ctx,"SELECT correlation_id,state,version,event_count FROM incidents WHERE correlation_id=$1 FOR UPDATE",id).Scan(&i.ID,&i.State,&i.Version,&i.EventCount)
	if err!=nil{return Incident{},err}
	if i.Version!=expected{return Incident{},ErrVersion}
	steps,err:=lifecycle.Plan(i.State,action)
	if err!=nil{return Incident{},err}
	for _,step:=range steps{
		i.Version++
		_,err=tx.ExecContext(ctx,"INSERT INTO incident_audit(correlation_id,version,action,previous_state,next_state) VALUES($1,$2,$3,$4,$5)",id,i.Version,step.Action,i.State,step.State)
		if err!=nil{return Incident{},err}
		i.State=step.State
	}
	_,err=tx.ExecContext(ctx,"UPDATE incidents SET state=$2,version=$3,updated_at=now() WHERE correlation_id=$1",id,i.State,i.Version)
	if err!=nil{return Incident{},err}
	if err=tx.Commit();err!=nil{return Incident{},err}
	return i,nil
}

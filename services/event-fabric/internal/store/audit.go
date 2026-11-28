package store

import (
 "context"
 "time"
)

type AuditEntry struct {
 Version int64 `json:"version"`
 Action string `json:"action"`
 PreviousState string `json:"previous_state"`
 NextState string `json:"next_state"`
 ActorRole string `json:"actor_role"`
 CreatedAt time.Time `json:"created_at"`
}

type AuditPage struct {
 Entries []AuditEntry `json:"entries"`
 NextAfter *int64 `json:"next_after,omitempty"`
}

// Audit returns a bounded, version-ordered history for an existing incident.
// The version is a stable cursor, independent of clock ordering.
func (s Store) Audit(ctx context.Context,id string,after int64)(AuditPage,error){
 page:=AuditPage{Entries:make([]AuditEntry,0)}
 if _,err:=s.Incident(ctx,id);err!=nil{return page,err}
 rows,err:=s.DB.QueryContext(ctx,`SELECT version,action,previous_state,next_state,actor_role,created_at
 FROM incident_audit WHERE correlation_id=$1 AND version>$2 ORDER BY version LIMIT 101`,id,after)
 if err!=nil{return page,err}
 defer rows.Close()
 for rows.Next(){
  var e AuditEntry
  if err=rows.Scan(&e.Version,&e.Action,&e.PreviousState,&e.NextState,&e.ActorRole,&e.CreatedAt);err!=nil{return page,err}
  if len(page.Entries)==100{v:=page.Entries[99].Version;page.NextAfter=&v;break}
  page.Entries=append(page.Entries,e)
 }
 return page,rows.Err()
}

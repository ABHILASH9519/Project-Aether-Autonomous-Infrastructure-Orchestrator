package store

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "time"
)

type Evidence struct {
 Key string
 Body []byte
 SHA256 string
}

// ArchiveOne delivers one terminal job's immutable evidence. The adapter
// must verify the uploaded bytes before returning nil. The transaction
// retains its row lock during bounded I/O, and retries use the same key
// and bytes if publication succeeded but the database commit did not.
func (s Store) ArchiveOne(ctx context.Context, publish func(context.Context,Evidence)error)(bool,error){
 tx,err:=s.DB.BeginTx(ctx,nil)
 if err!=nil{return false,err}
 defer tx.Rollback()
 var j Job
 var attempts int64
 err=tx.QueryRowContext(ctx,`SELECT correlation_id,status,attempts,evidence,updated_at,archive_attempts
 FROM remediation_jobs WHERE status IN ('succeeded','failed') AND archived_at IS NULL AND archive_available_at<=now()
 ORDER BY archive_available_at,correlation_id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&j.CorrelationID,&j.Status,&j.Attempts,&j.Evidence,&j.UpdatedAt,&attempts)
 if errors.Is(err,sql.ErrNoRows){return false,nil}
 if err!=nil{return false,err}
 // Match the JSON job API, including the trailing newline.
 body,err:=json.Marshal(j)
 if err!=nil{return false,err}
 body=append(body,'\n')
 digest:=sha256.Sum256(body)
 idHash:=sha256.Sum256([]byte(j.CorrelationID))
 e:=Evidence{Key:"jobs/"+hex.EncodeToString(idHash[:])+".json",Body:body,SHA256:hex.EncodeToString(digest[:])}
 deliveryErr:=publish(ctx,e)
 if ctx.Err()!=nil{return false,ctx.Err()}
 if deliveryErr!=nil{
  delay:=int64(60)
  if attempts<29{delay=2*(attempts+1)}
  _,err=tx.ExecContext(ctx,`UPDATE remediation_jobs SET archive_attempts=archive_attempts+1,
   archive_available_at=now()+($2::integer * interval '1 second') WHERE correlation_id=$1`,j.CorrelationID,delay)
 }else{
  _,err=tx.ExecContext(ctx,`UPDATE remediation_jobs SET archive_attempts=archive_attempts+1,
   archived_at=now(),archive_sha256=$2 WHERE correlation_id=$1`,j.CorrelationID,e.SHA256)
 }
 if err!=nil{return false,err}
 if err=tx.Commit();err!=nil{return false,err}
 return true,deliveryErr
}

type ArchiveStatus struct {
 State string `json:"state"`
 Attempts int64 `json:"attempts"`
 SHA256 string `json:"sha256,omitempty"`
 ArchivedAt *time.Time `json:"archived_at,omitempty"`
}
func(s Store) ArchiveStatus(ctx context.Context,id string)(ArchiveStatus,error){
 var a ArchiveStatus
 var status string
 var at sql.NullTime
 err:=s.DB.QueryRowContext(ctx,"SELECT status,archive_attempts,archive_sha256,archived_at FROM remediation_jobs WHERE correlation_id=$1",id).Scan(&status,&a.Attempts,&a.SHA256,&at)
 if err!=nil{return a,err}
 switch{
 case at.Valid:a.State="archived";a.ArchivedAt=&at.Time
 case status=="succeeded"||status=="failed":a.State="pending"
 default:a.State="not_eligible"
 }
 return a,nil
}

package store

import (
	"context"
	"database/sql"
	"errors"
)

type OutboxRecord struct {
	ID int64
	EventID string
	CorrelationID string
	Payload []byte
}

// RelayOne locks one pending record until publication and database commit.
// Publication can repeat if the process dies after broker acknowledgement.
func (s Store) RelayOne(ctx context.Context, publish func(context.Context, OutboxRecord) error) (bool, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil { return false, err }
	defer tx.Rollback()
	var r OutboxRecord
	err = tx.QueryRowContext(ctx, `SELECT id,event_id,correlation_id,payload
FROM incident_outbox WHERE published_at IS NULL
ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&r.ID,&r.EventID,&r.CorrelationID,&r.Payload)
	if errors.Is(err,sql.ErrNoRows) { return false,nil }
	if err != nil { return false,err }
	if err = publish(ctx,r); err != nil { return false,err }
	if _,err = tx.ExecContext(ctx,"UPDATE incident_outbox SET published_at=now() WHERE id=$1",r.ID); err != nil { return false,err }
	return true,tx.Commit()
}

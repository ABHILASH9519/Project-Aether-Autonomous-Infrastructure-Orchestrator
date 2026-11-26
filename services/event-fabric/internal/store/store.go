package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"

	"aether/event-fabric/internal/event"
)

//go:embed schema.sql
var schema string

type Store struct { DB *sql.DB }

func (s Store) Migrate(ctx context.Context) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(62391501)"); err != nil { return err }
	if _, err = tx.ExecContext(ctx, schema); err != nil { return err }
	return tx.Commit()
}

// Apply atomically records a unique event, updates the incident and queues an outbox record.
func (s Store) Apply(ctx context.Context, e event.Envelope) (bool, error) {
	if err := e.Validate(); err != nil { return false, err }
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil { return false, err }
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "INSERT INTO incident_inbox(event_id) VALUES ($1) ON CONFLICT DO NOTHING", e.ID)
	if err != nil { return false, err }
	n, err := res.RowsAffected()
	if err != nil { return false, err }
	if n == 0 { return false, tx.Commit() }
	_, err = tx.ExecContext(ctx, `INSERT INTO incidents(correlation_id,last_event_type) VALUES ($1,$2)
ON CONFLICT(correlation_id) DO UPDATE SET event_count=incidents.event_count+1,last_event_type=EXCLUDED.last_event_type,updated_at=now()`, e.CorrelationID,e.Type)
	if err != nil { return false, err }
	payload, err := json.Marshal(e)
	if err != nil { return false, err }
	_, err = tx.ExecContext(ctx, "INSERT INTO incident_outbox(event_id,correlation_id,payload) VALUES ($1,$2,$3::jsonb)",e.ID,e.CorrelationID,string(payload))
	if err != nil { return false, err }
	if err := tx.Commit(); err != nil { return false, err }
	return true,nil
}

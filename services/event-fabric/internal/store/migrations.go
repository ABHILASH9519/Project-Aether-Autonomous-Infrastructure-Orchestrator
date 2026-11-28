package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
)

// schema.sql is immutable after this migration is published. Add new files
// and ordered entries for future changes rather than editing version 1.
//go:embed schema.sql
var baselineSchema string

type migration struct {
	version int
	sql string
}

//go:embed 002_trace_context.sql
var traceSchema string

//go:embed 003_evidence_archive.sql
var archiveSchema string

var migrations = []migration{{1,baselineSchema},{2,traceSchema},{3,archiveSchema}}

// Migrate serializes schema changes and commits each migration ledger entry
// with its DDL. Existing installations without a ledger adopt the idempotent
// baseline once. Restarts no longer rebuild constraints on the hot path.
func (s Store) Migrate(ctx context.Context) error {
	tx,err:=s.DB.BeginTx(ctx,nil)
	if err!=nil{return err}
	defer tx.Rollback()
	if _,err=tx.ExecContext(ctx,"SELECT pg_advisory_xact_lock(62391501)");err!=nil{return err}
	if _,err=tx.ExecContext(ctx,`CREATE TABLE IF NOT EXISTS aether_schema_migrations (
version integer PRIMARY KEY CHECK(version>0), checksum text NOT NULL,
applied_at timestamptz NOT NULL DEFAULT now())`);err!=nil{return err}
	var latest int
	if err=tx.QueryRowContext(ctx,"SELECT COALESCE(max(version),0) FROM aether_schema_migrations").Scan(&latest);err!=nil{return err}
	if latest>migrations[len(migrations)-1].version{return errors.New("database schema is newer than this binary")}
	for _,m:=range migrations{
		checksum:=fmt.Sprintf("%x",sha256.Sum256([]byte(m.sql)))
		var existing string
		err=tx.QueryRowContext(ctx,"SELECT checksum FROM aether_schema_migrations WHERE version=$1",m.version).Scan(&existing)
		switch {
		case err==nil:
			if existing!=checksum{return fmt.Errorf("migration %d checksum mismatch",m.version)}
		case errors.Is(err,sql.ErrNoRows):
			if m.version<=latest{return fmt.Errorf("migration ledger gap at version %d",m.version)}
			if _,err=tx.ExecContext(ctx,m.sql);err!=nil{return fmt.Errorf("migration %d: %w",m.version,err)}
			if _,err=tx.ExecContext(ctx,"INSERT INTO aether_schema_migrations(version,checksum) VALUES($1,$2)",m.version,checksum);err!=nil{return err}
		default:return err
		}
	}
	return tx.Commit()
}

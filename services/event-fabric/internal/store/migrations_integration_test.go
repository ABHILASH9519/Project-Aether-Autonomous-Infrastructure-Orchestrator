package store

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestMigrationsConcurrentAndStable(t *testing.T){
	dsn:=os.Getenv("AETHER_TEST_DATABASE_URL")
	if dsn==""{t.Skip("requires disposable PostgreSQL")}
	db,err:=sql.Open("postgres",dsn)
	if err!=nil{t.Fatal(err)}
	defer db.Close()
	ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second)
	defer cancel()
	s:=Store{DB:db}
	if err=s.Migrate(ctx);err!=nil{t.Fatal(err)}
	const query="SELECT oid::bigint FROM pg_constraint WHERE conrelid='incidents'::regclass AND conname='incidents_state_check'"
	var before,after int64
	if err=db.QueryRowContext(ctx,query).Scan(&before);err!=nil{t.Fatal(err)}
	var wg sync.WaitGroup
	errs:=make(chan error,4)
	for i:=0;i<4;i++{
		wg.Add(1)
		go func(){defer wg.Done();errs<-s.Migrate(ctx)}()
	}
	wg.Wait()
	close(errs)
	for e:=range errs{if e!=nil{t.Fatal(e)}}
	if err=db.QueryRowContext(ctx,query).Scan(&after);err!=nil{t.Fatal(err)}
	if before!=after{t.Fatalf("constraint recreated: %d -> %d",before,after)}
	var count int
	if err=db.QueryRowContext(ctx,"SELECT count(*) FROM aether_schema_migrations WHERE version=1 AND length(checksum)=64").Scan(&count);err!=nil || count!=1{t.Fatalf("ledger count=%d err=%v",count,err)}
}

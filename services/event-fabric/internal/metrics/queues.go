package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Queues struct {
	OutboxPending int64
	OutboxOldestSeconds float64
	JobsPending int64
	JobsFailed int64
}

// QueryQueues collects a consistent snapshot in a single bounded SQL statement.
func QueryQueues(ctx context.Context, db *sql.DB) (Queues,error) {
	var q Queues
	err:=db.QueryRowContext(ctx,`SELECT
	(SELECT count(*) FROM incident_outbox WHERE published_at IS NULL),
	(SELECT COALESCE(GREATEST(EXTRACT(epoch FROM now()-min(created_at)),0),0)::double precision FROM incident_outbox WHERE published_at IS NULL),
	(SELECT count(*) FROM remediation_jobs WHERE status='pending'),
	(SELECT count(*) FROM remediation_jobs WHERE status='failed')`).Scan(&q.OutboxPending,&q.OutboxOldestSeconds,&q.JobsPending,&q.JobsFailed)
	return q,err
}

type QueueCollector func(context.Context)(Queues,error)

// WithQueues augments an existing Prometheus endpoint. A failed database
// query exposes an explicit failure gauge, never misleading zero queue depths.
func WithQueues(next http.Handler, collect QueueCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.Method!="GET" || r.URL.Path!="/metrics"{next.ServeHTTP(w,r);return}
		// The existing non-streaming HTTP metrics endpoint writes a 200 response.
		next.ServeHTTP(w,r)
		ctx,cancel:=context.WithTimeout(r.Context(),time.Second)
		defer cancel()
		q,err:=collect(ctx)
		writeQueues(w,q,err)
	})
}

func writeQueues(w io.Writer,q Queues,err error) {
	fmt.Fprintln(w,"# HELP aether_queue_collection_success Whether the database queue snapshot succeeded.")
	fmt.Fprintln(w,"# TYPE aether_queue_collection_success gauge")
	if err!=nil{fmt.Fprintln(w,"aether_queue_collection_success 0");return}
	fmt.Fprintln(w,"aether_queue_collection_success 1")
	fmt.Fprintf(w,"# TYPE aether_outbox_pending gauge\naether_outbox_pending %d\n",q.OutboxPending)
	fmt.Fprintf(w,"# TYPE aether_outbox_oldest_seconds gauge\naether_outbox_oldest_seconds %f\n",q.OutboxOldestSeconds)
	fmt.Fprintf(w,"# TYPE aether_remediation_pending gauge\naether_remediation_pending %d\n",q.JobsPending)
	fmt.Fprintf(w,"# TYPE aether_remediation_failed gauge\naether_remediation_failed %d\n",q.JobsFailed)
}

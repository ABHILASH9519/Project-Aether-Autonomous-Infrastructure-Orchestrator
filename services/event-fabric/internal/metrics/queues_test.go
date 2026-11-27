package metrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueueCollection(t *testing.T){
	for _,fail:=range []bool{false,true}{
		h:=WithQueues((&HTTP{}).Wrap(http.NotFoundHandler()),func(ctx context.Context)(Queues,error){
			if _,ok:=ctx.Deadline();!ok{t.Error("missing deadline")}
			if fail{return Queues{},errors.New("private SQL failure")}
			return Queues{OutboxPending:3,OutboxOldestSeconds:4.5,JobsPending:2,JobsFailed:1},nil
		})
		w:=httptest.NewRecorder()
		h.ServeHTTP(w,httptest.NewRequest("GET","/metrics",nil))
		body:=w.Body.String()
		if !strings.Contains(body,"aether_http_requests_total"){t.Fatal(body)}
		if fail{
			if !strings.Contains(body,"aether_queue_collection_success 0") || strings.Contains(body,"aether_outbox_pending") || strings.Contains(body,"private"){t.Fatal(body)}
		}else if !strings.Contains(body,"aether_queue_collection_success 1") || !strings.Contains(body,"aether_outbox_pending 3") || !strings.Contains(body,"aether_remediation_failed 1"){t.Fatal(body)}
	}
}

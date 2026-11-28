#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?required}"
: "${AETHER_APPROVER_TOKEN:?required}"
ns=aether
# Isolated CI cluster only: scale app processes, leave singleton fixture/data alone.
restore(){
 kubectl -n "$ns" scale deployment/consumer deployment/relay deployment/executor --replicas=1 >/dev/null 2>&1 || true
}
trap restore EXIT
kubectl -n "$ns" scale deployment/consumer deployment/relay deployment/executor --replicas=2
for service in consumer relay executor; do
 kubectl -n "$ns" rollout status deployment/"$service" --timeout=180s
 test "$(kubectl -n "$ns" get deployment "$service" -o jsonpath='{.status.readyReplicas}')" = 2
done
kubectl -n "$ns" port-forward svc/ingest 8081:8081 >/tmp/aether-scale-ingest.log 2>&1 &
p1=$!
kubectl -n "$ns" port-forward svc/workflow 8082:8082 >/tmp/aether-scale-workflow.log 2>&1 &
p2=$!
trap 'kill "$p1" "$p2" 2>/dev/null || true; restore' EXIT
for i in {1..30}; do
 if curl -fsS localhost:8081/readyz >/dev/null && curl -fsS localhost:8082/readyz >/dev/null; then break; fi
 sleep 2
done
export AETHER_EXECUTION_ID="worker-failover-$(date +%s)-$$"
# Terminate one pod in each pair, then immediately exercise the real pipeline.
for service in consumer relay executor; do
 pod=$(kubectl -n "$ns" get pods -l app="$service" -o jsonpath='{.items[0].metadata.name}')
 test -n "$pod"
 kubectl -n "$ns" delete pod "$pod" --wait=false
done
bash tests/execution-smoke.sh
for service in consumer relay executor; do
 kubectl -n "$ns" rollout status deployment/"$service" --timeout=180s
 test "$(kubectl -n "$ns" get deployment "$service" -o jsonpath='{.status.readyReplicas}')" = 2
done
curl -fsS -H "Authorization: Bearer $AETHER_OPERATOR_TOKEN" "http://localhost:8082/api/v1/incidents/$AETHER_EXECUTION_ID" | jq -e '.state=="resolved" and .version==6 and .event_count==1'
echo "Application workers recovered from pod deletion with one durable incident effect"

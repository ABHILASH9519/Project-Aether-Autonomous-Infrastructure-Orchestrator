#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?required}"
: "${AETHER_APPROVER_TOKEN:?required}"
ns=aether
cluster=aether-postgres
kubectl -n "$ns" wait --for=condition=Ready cluster/"$cluster" --timeout=600s
kubectl -n "$ns" wait --for=condition=complete job/kafka-topics --timeout=300s
kubectl -n "$ns" rollout status deployment/ingest --timeout=300s
kubectl -n "$ns" rollout status deployment/workflow --timeout=300s
kubectl -n "$ns" port-forward svc/ingest 8081:8081 >/tmp/aether-pgha-ingest.log 2>&1 &
p1=$!
kubectl -n "$ns" port-forward svc/workflow 8082:8082 >/tmp/aether-pgha-workflow.log 2>&1 &
p2=$!
tmp=$(mktemp -d)
trap 'kill "$p1" "$p2" 2>/dev/null || true; rm -rf "$tmp"' EXIT
for i in {1..60}; do
 if curl --max-time 5 -fsS localhost:8081/readyz >/dev/null && curl --max-time 5 -fsS localhost:8082/readyz >/dev/null; then break; fi
 sleep 2
done
primary(){ kubectl -n "$ns" get cluster "$cluster" -o jsonpath='{.status.currentPrimary}'; }
old=$(primary)
test -n "$old"
# Require two streaming standby connections and at least one synchronous/quorum peer.
for i in {1..60}; do
 if kubectl -n "$ns" exec "$old" -- psql -X -qAt -d postgres -c "SELECT count(*)=2 AND count(*) FILTER (WHERE sync_state IN ('sync','quorum'))>=1 FROM pg_stat_replication WHERE state='streaming'" | grep -Fx t; then break; fi
 sleep 2
done
kubectl -n "$ns" exec "$old" -- psql -X -qAt -d postgres -c "SELECT count(*)=2 AND count(*) FILTER (WHERE sync_state IN ('sync','quorum'))>=1 FROM pg_stat_replication WHERE state='streaming'" | grep -Fx t
export AETHER_EXECUTION_ID="pg-ha-before-$(date +%s)-$$"
bash tests/execution-smoke.sh
url="http://localhost:8082/api/v1/incidents/$AETHER_EXECUTION_ID"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
for suffix in incident job audit; do
 path="$url"; [[ "$suffix" == incident ]] || path="$url/$suffix"
 curl -fsS -H "$auth" "$path" | jq -S . > "$tmp/$suffix-before"
done
kubectl -n "$ns" delete pod "$old" --wait=false
new=""
for i in {1..120}; do
 candidate=$(primary)
 if [[ -n "$candidate" && "$candidate" != "$old" ]] && kubectl -n "$ns" get cluster "$cluster" -o json | jq -e '.status.readyInstances>=2' >/dev/null && curl --max-time 5 -fsS localhost:8082/readyz >/dev/null; then new="$candidate"; break; fi
 sleep 2
done
test -n "$new"
echo "Primary promoted: $old -> $new"
for suffix in incident job audit; do
 path="$url"; [[ "$suffix" == incident ]] || path="$url/$suffix"
 curl -fsS -H "$auth" "$path" | jq -S . > "$tmp/$suffix-after"
 diff -u "$tmp/$suffix-before" "$tmp/$suffix-after"
done
export AETHER_EXECUTION_ID="pg-ha-after-$(date +%s)-$$"
AETHER_INGEST_ATTEMPTS=10 bash tests/execution-smoke.sh
kubectl -n "$ns" wait --for=condition=Ready cluster/"$cluster" --timeout=600s
kubectl -n "$ns" get cluster "$cluster" -o json | jq -e '.status.readyInstances==3'
echo "PostgreSQL primary replacement preserved incident/job/audit and completed a new workflow"

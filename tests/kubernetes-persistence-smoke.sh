#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_APPROVER_TOKEN:?set approver token}"
# Destructive to pods, not claims. Use only the disposable aether namespace.
ns=aether
kubectl -n "$ns" wait --for=jsonpath='{.status.phase}'=Bound pvc/postgres-data pvc/kafka-data --timeout=180s
export AETHER_EXECUTION_ID="persistence-$(date +%s)-$$"
kubectl -n "$ns" port-forward svc/ingest 8081:8081 >/tmp/aether-persist-ingest.log 2>&1 &
p1=$!
kubectl -n "$ns" port-forward svc/workflow 8082:8082 >/tmp/aether-persist-workflow.log 2>&1 &
p2=$!
tmp=$(mktemp -d)
trap 'kill "$p1" "$p2" 2>/dev/null || true; rm -rf "$tmp"' EXIT
for i in {1..30}; do
  if curl -fsS localhost:8081/healthz >/dev/null && curl -fsS localhost:8082/healthz >/dev/null; then break; fi
  sleep 2
done
bash tests/execution-smoke.sh
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
url="http://localhost:8082/api/v1/incidents/$AETHER_EXECUTION_ID"
curl -fsS -H "$auth" "$url" | jq -S . > "$tmp/incident-before.json"
curl -fsS -H "$auth" "$url/job" | jq -S . > "$tmp/job-before.json"
offsets() {
  kubectl -n "$ns" exec deployment/kafka -- /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server localhost:29092 --topic aether.events.v1 | grep '^aether.events.v1:' | sort
}
offsets > "$tmp/offsets-before"
test "$(wc -l < "$tmp/offsets-before")" -eq 3
for service in postgres kafka; do
  old=$(kubectl -n "$ns" get pod -l app="$service" -o jsonpath='{.items[0].metadata.uid}')
  test -n "$old"
  kubectl -n "$ns" rollout restart deployment/"$service"
  kubectl -n "$ns" rollout status deployment/"$service" --timeout=300s
  new=$(kubectl -n "$ns" get pod -l app="$service" -o jsonpath='{.items[0].metadata.uid}')
  test -n "$new"
  test "$old" != "$new"
done
for i in {1..30}; do
  if offsets > "$tmp/offsets-after" 2>/dev/null && test "$(wc -l < "$tmp/offsets-after")" -eq 3; then break; fi
  sleep 2
done
diff -u "$tmp/offsets-before" "$tmp/offsets-after"
for i in {1..30}; do
  if curl -fsS -H "$auth" "$url" > "$tmp/incident-raw.json" 2>/dev/null; then break; fi
  sleep 2
done
jq -S . "$tmp/incident-raw.json" > "$tmp/incident-after.json"
curl -fsS -H "$auth" "$url/job" | jq -S . > "$tmp/job-after.json"
diff -u "$tmp/incident-before.json" "$tmp/incident-after.json"
diff -u "$tmp/job-before.json" "$tmp/job-after.json"
echo "PostgreSQL incident/job state and Kafka partition end offsets survived pod replacement"

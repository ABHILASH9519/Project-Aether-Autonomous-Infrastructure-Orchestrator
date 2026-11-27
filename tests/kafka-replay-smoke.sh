#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
id="replay-$(date +%s)-$$"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
url="http://localhost:8082/api/v1/incidents/$id"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"replay-smoke","correlation_id":"%s","payload":{}}' "$id" "$id")
send() { curl --max-time 10 -fsS -H 'Content-Type: application/json' -d "$body" http://localhost:8081/api/v1/events; }
send
for i in {1..60}; do
  if curl -fsS -H "$auth" "$url" 2>/dev/null | jq -e '.event_count==1' >/dev/null; then break; fi
  sleep 1
done
curl -fsS -H "$auth" "$url" | jq -e '.event_count==1'
trap '"${compose[@]}" start consumer >/dev/null || true' EXIT
"${compose[@]}" stop consumer
# Append the exact same envelope while the consumer is offline.
send
"${compose[@]}" start consumer
for i in {1..60}; do
  if "${compose[@]}" logs consumer | grep -F "event_id=\"$id\"" | grep -F 'applied=false' >/dev/null; then break; fi
  sleep 1
done
"${compose[@]}" logs consumer | grep -F "event_id=\"$id\"" | grep -F 'applied=false'
curl -fsS -H "$auth" "$url" | jq -e '.event_count==1'
counts=$("${compose[@]}" exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U aether -d aether -v "event_id=$id" <<'SQL'
SELECT (SELECT count(*) FROM incident_inbox WHERE event_id=:'event_id') || ':' ||
       (SELECT count(*) FROM incident_outbox WHERE event_id=:'event_id');
SQL
)
test "$counts" = "1:1"
trap - EXIT
echo "Consumer restart and duplicate Kafka record preserved one inbox/outbox effect"

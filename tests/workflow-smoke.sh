#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set a local operator token}"
id="workflow-$(date +%s)-$$"
url="http://127.0.0.1:8082/api/v1/incidents/$id"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"workflow-smoke","correlation_id":"%s","payload":{}}' "$id" "$id")
curl -fsS -H 'Content-Type: application/json' -d "$body" http://127.0.0.1:8081/api/v1/events
for i in {1..60}; do
  if curl -fsS -H "$auth" "$url" >/dev/null 2>&1; then break; fi
  sleep 2
done
test "$(curl -s -o /dev/null -w '%{http_code}' "$url")" = 401
transition() {
  local version="$1" action="$2" state="$3"
  curl -fsS -H "$auth" -H 'Content-Type: application/json' \
    -d "{\"expected_version\":$version,\"action\":\"$action\"}" "$url/transitions" |
    jq -e --arg state "$state" '.state == $state' >/dev/null
}
test "$(curl -s -o /dev/null -w '%{http_code}' -H "$auth" -d '{"expected_version":0,"action":"simulate"}' "$url/transitions")" = 409
transition 0 acknowledge acknowledged
transition 1 request_approval awaiting_approval
transition 2 approve approved
transition 3 simulate resolved
curl -fsS -H "$auth" "$url" | jq -e '.version == 6 and .state == "resolved"'

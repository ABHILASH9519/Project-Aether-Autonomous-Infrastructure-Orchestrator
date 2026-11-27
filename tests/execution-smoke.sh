#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_APPROVER_TOKEN:?set approver token}"
id="${AETHER_EXECUTION_ID:-execution-$(date +%s)-$}"
url="http://127.0.0.1:8082/api/v1/incidents/$id"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"execution-smoke","correlation_id":"%s","payload":{}}' "$id" "$id")
curl -fsS -H 'Content-Type: application/json' -d "$body" http://127.0.0.1:8081/api/v1/events
for i in {1..60}; do
  if curl -fsS -H "$auth" "$url" >/dev/null 2>&1; then break; fi
  sleep 2
done
test "$(curl -s -o /dev/null -w '%{http_code}' -H "$auth" -d '{"expected_version":0,"action":"execute_lab"}' "$url/transitions")" = 409
transition() {
  local version="$1" action="$2" state="$3" credential="$auth"
  if [[ "$action" == approve ]]; then credential="Authorization: Bearer $AETHER_APPROVER_TOKEN"; fi
  local trace_header=()
  if [[ -n "${AETHER_EXECUTION_TRACEPARENT:-}" && "$action" == execute_lab ]]; then trace_header=(-H "traceparent: $AETHER_EXECUTION_TRACEPARENT"); fi
  curl -fsS "${trace_header[@]}" -H "$credential" -H 'Content-Type: application/json' -d "{\"expected_version\":$version,\"action\":\"$action\"}" "$url/transitions" |
    jq -e --arg state "$state" '.state == $state' >/dev/null
}
transition 0 acknowledge acknowledged
transition 1 request_approval awaiting_approval
transition 2 approve approved
transition 3 execute_lab executing
for i in {1..90}; do
  if curl -fsS -H "$auth" "$url" | jq -e '.state=="resolved" and .version==6' >/dev/null; then
    curl -fsS -H "$auth" "$url/job" | jq -e '.status=="succeeded" and .attempts>=1 and .attempts<=3 and .evidence=="fixture content verified"'
    exit 0
  fi
  sleep 2
done
echo "lab execution did not resolve incident" >&2
exit 1

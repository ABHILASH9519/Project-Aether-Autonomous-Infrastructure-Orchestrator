#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set a local operator token}"
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability)
id="policy-outage-$(date +%s)-$$"
url="http://127.0.0.1:8082/api/v1/incidents/$id"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
cmd='{"expected_version":0,"action":"acknowledge"}'
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"policy-outage-smoke","correlation_id":"%s","payload":{}}' "$id" "$id")
curl -fsS -H 'Content-Type: application/json' -d "$body" http://127.0.0.1:8081/api/v1/events
for i in {1..30}; do
  if curl -fsS -H "$auth" "$url" 2>/dev/null | jq -e '.state=="open" and .version==0' >/dev/null; then break; fi
  sleep 2
done
curl -fsS -H "$auth" "$url" | jq -e '.state=="open" and .version==0' >/dev/null
trap '"${compose[@]}" start opa >/dev/null || true' EXIT
"${compose[@]}" stop opa
test "$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -H "$auth" -H 'Content-Type: application/json' -d "$cmd" "$url/transitions")" = 503
curl -fsS -H "$auth" "$url" | jq -e '.state=="open" and .version==0' >/dev/null
"${compose[@]}" start opa
for i in {1..30}; do
  code=$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -H "$auth" -H 'Content-Type: application/json' -d "$cmd" "$url/transitions")
  if [[ "$code" == 200 ]]; then
    curl -fsS -H "$auth" "$url" | jq -e '.state=="acknowledged" and .version==1'
    trap - EXIT
    exit 0
  fi
  [[ "$code" == 503 ]] || { echo "unexpected recovery status: $code" >&2; exit 1; }
  sleep 2
done
echo "policy service did not recover" >&2
exit 1

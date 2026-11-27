#!/usr/bin/env bash
set -euo pipefail
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
id="redis-outage-$(date +%s)-$$"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"redis-outage","correlation_id":"%s","payload":{}}' "$id" "$id")
trap '"${compose[@]}" start redis >/dev/null || true' EXIT
"${compose[@]}" stop redis
test "$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -d "$body" http://localhost:8081/api/v1/events)" = 503
curl -fsS http://localhost:8081/healthz
test "$(curl --max-time 5 -sS -o /dev/null -w '%{http_code}' http://localhost:8081/readyz)" = 503
"${compose[@]}" start redis
for i in {1..30}; do
 code=$(curl --max-time 10 -sS -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -d "$body" http://localhost:8081/api/v1/events)
 if [[ "$code" == 202 ]]; then curl -fsS http://localhost:8081/readyz; trap - EXIT; exit 0; fi
 [[ "$code" == 503 ]] || { echo "unexpected status $code" >&2; exit 1; }
 sleep 2
done
exit 1

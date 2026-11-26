#!/usr/bin/env bash
set -euo pipefail
compose=(docker compose -f deploy/compose/events.yaml)
for i in {1..60}; do
  if curl -fsS http://localhost:8081/healthz; then break; fi
  sleep 2
done
id="smoke-$(date +%s)-$$"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"smoke","correlation_id":"incident-1","payload":{"severity":"warning"}}' "$id")
test "$(curl -sS -o /dev/null -w '%{http_code}' -H 'Content-Type: application/json' -d "$body" http://localhost:8081/api/v1/events)" = 202
test "$(curl -sS -o /dev/null -w '%{http_code}' -d '{}' http://localhost:8081/api/v1/events)" = 400
for i in {1..30}; do
  if "${compose[@]}" logs consumer | grep -F "event_id=\"$id\""; then exit 0; fi
  sleep 2
done
echo "consumer did not observe event" >&2
exit 1

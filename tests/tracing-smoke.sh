#!/usr/bin/env bash
set -euo pipefail
trace=$(openssl rand -hex 16)
span=$(openssl rand -hex 8)
id="trace-$(date +%s)-$$"
body=$(printf '{"event_id":"%s","event_type":"Infrastructure.Alert","schema_version":1,"occurred_at":"2026-01-01T00:00:00Z","source":"trace-smoke","correlation_id":"%s","payload":{}}' "$id" "$id")
curl -fsS -H "traceparent: 00-$trace-$span-01" -H 'Content-Type: application/json' -d "$body" http://localhost:8081/api/v1/events
for i in {1..45}; do
 if curl -fsS "http://localhost:16686/api/traces/$trace" 2>/dev/null | jq -e '[.data[].spans[].operationName] | index("ingest HTTP") != null and index("incident consume") != null and index("outbox publish") != null and index("outbox publish") != null' >/dev/null; then
  echo "HTTP, Kafka consumer and outbox spans share trace $trace"
  exit 0
 fi
 sleep 2
done
echo "trace propagation was not observed" >&2
exit 1

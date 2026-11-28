#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_APPROVER_TOKEN:?set approver token}"
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
# Pause rather than restart the ephemeral emulator: preserve existing objects.
"${compose[@]}" pause localstack
trap '"${compose[@]}" unpause localstack >/dev/null 2>&1 || true' EXIT
export AETHER_EXECUTION_ID="cloud-outage-$(date +%s)-$$"
bash tests/execution-smoke.sh
url="http://127.0.0.1:8082/api/v1/incidents/$AETHER_EXECUTION_ID"
auth="Authorization: Bearer $AETHER_OPERATOR_TOKEN"
observed=false
for i in {1..45}; do
  if curl --max-time 5 -fsS -H "$auth" "$url/archive" | jq -e '.state=="pending" and .attempts>=1' >/dev/null; then observed=true; break; fi
  sleep 2
done
test "$observed" = true
curl -fsS -H "$auth" "$url" | jq -e '.state=="resolved" and .version==6'
curl -fsS -H "$auth" "$url/job" | jq -e '.status=="succeeded"'
"${compose[@]}" unpause localstack
trap - EXIT
bash tests/cloud-evidence-smoke.sh
echo "LocalStack outage preserved completed repair and recovered durable evidence delivery"

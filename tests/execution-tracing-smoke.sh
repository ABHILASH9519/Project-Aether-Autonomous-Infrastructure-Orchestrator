#!/usr/bin/env bash
set -euo pipefail
trace=$(openssl rand -hex 16)
span=$(openssl rand -hex 8)
export AETHER_EXECUTION_TRACEPARENT="00-$trace-$span-01"
export AETHER_EXECUTION_ID="execution-trace-$(date +%s)-$$"
bash tests/execution-smoke.sh
for i in {1..45}; do
 if curl -fsS "http://localhost:16686/api/traces/$trace" 2>/dev/null | jq -e '[.data[].spans[].operationName] | index("workflow transition") != null and index("remediation attempt") != null' >/dev/null; then
  echo "Workflow and remediation worker share trace $trace"
  exit 0
 fi
 sleep 2
done
echo "Durable execution trace not observed" >&2
exit 1

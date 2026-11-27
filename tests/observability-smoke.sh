#!/usr/bin/env bash
set -euo pipefail
for i in {1..45}; do
  if curl -fsS http://localhost:9090/-/ready >/dev/null 2>&1 &&
     curl -fsS http://localhost:3000/api/health >/dev/null 2>&1 &&
     curl -fsS http://localhost:9090/api/v1/targets | jq -e '[.data.activeTargets[] | select(.health == "up" and (.labels.job == "ingest" or .labels.job == "workflow"))] | length == 2' >/dev/null; then
    curl -fsS http://localhost:8081/metrics | grep -F 'aether_http_requests_total'
    curl -fsS http://localhost:8082/metrics | grep -F 'aether_http_inflight_requests'
    curl -fsS http://localhost:3000/api/dashboards/uid/aether-http | jq -e '.dashboard.uid == "aether-http"'
    curl -fsS http://localhost:8082/metrics | grep -Fx 'aether_queue_collection_success 1'
    curl -fsS http://localhost:9090/api/v1/rules | jq -e '[.data.groups[].rules[].name] | index("AetherOutboxStalled") != null' >/dev/null
    exit 0
  fi
  sleep 2
done
echo "observability stack failed readiness or scrape checks" >&2
exit 1

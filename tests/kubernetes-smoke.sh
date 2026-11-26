#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_APPROVER_TOKEN:?set approver token}"
kubectl -n aether wait --for=condition=complete job/kafka-topics --timeout=300s
kubectl -n aether rollout status deployment/ingest --timeout=300s
kubectl -n aether rollout status deployment/workflow --timeout=300s
kubectl -n aether port-forward svc/ingest 8081:8081 >/tmp/aether-ingest-pf.log 2>&1 &
p1=$!
kubectl -n aether port-forward svc/workflow 8082:8082 >/tmp/aether-workflow-pf.log 2>&1 &
p2=$!
trap 'kill "$p1" "$p2" 2>/dev/null || true' EXIT
for i in {1..30}; do
  if curl -fsS localhost:8081/healthz && curl -fsS localhost:8082/healthz; then break; fi
  sleep 2
done
bash tests/workflow-smoke.sh
bash tests/execution-smoke.sh

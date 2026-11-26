#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_GIT_REVISION:?expected immutable Git commit SHA}"
[[ "$AETHER_GIT_REVISION" =~ ^[a-f0-9]{40}$ ]]
wait_for_sync() {
  for i in {1..120}; do
    if kubectl -n argocd get application aether-local -o json | jq -e --arg revision "$AETHER_GIT_REVISION" '
      .status.sync.status == "Synced" and
      .status.sync.revision == $revision and
      .status.health.status == "Healthy" and
      .status.operationState.phase == "Succeeded"
    ' >/dev/null; then return 0; fi
    sleep 5
  done
  kubectl -n argocd get application aether-local -o yaml
  return 1
}
wait_for_sync
bash tests/kubernetes-smoke.sh
# Inject drift only in the dedicated disposable namespace. The controller
# must restore replicas to the Git-declared count without a Helm upgrade.
kubectl -n aether scale deployment fixture --replicas=0
for i in {1..60}; do
  if [[ "$(kubectl -n aether get deployment fixture -o jsonpath='{.spec.replicas}')" == 1 ]]; then
    kubectl -n aether rollout status deployment/fixture --timeout=120s
    wait_for_sync
    echo "Argo CD restored fixture replicas to the Git desired state"
    exit 0
  fi
  sleep 5
done
echo "Argo CD did not repair replica drift" >&2
exit 1

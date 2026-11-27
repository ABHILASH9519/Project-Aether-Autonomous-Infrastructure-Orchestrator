#!/usr/bin/env bash
set -euo pipefail
ns=aether
probe="network-probe-$$"
trap 'kubectl -n "$ns" delete pod "$probe" --ignore-not-found --wait=false >/dev/null 2>&1 || true' EXIT
kubectl -n "$ns" run "$probe" --image=busybox:1.37.0 --restart=Never --command -- sleep 300
kubectl -n "$ns" wait --for=condition=Ready pod/"$probe" --timeout=120s
kubectl -n "$ns" get networkpolicy default-deny-ingress
# Require both a reachable API and denied backend connections. Resolve
# backend addresses from Kubernetes so DNS failure cannot fake a denial.
kubectl -n "$ns" exec "$probe" -- wget -q -T 5 -O /dev/null http://ingest:8081/healthz
kubectl -n "$ns" exec "$probe" -- wget -q -T 5 -O /dev/null http://workflow:8082/healthz
for item in postgres:5432 redis:6379 kafka:29092 opa:8181 fixture:8083; do
  service=${item%:*}
  port=${item#*:}
  kubectl -n "$ns" rollout status deployment/"$service" --timeout=120s
  ip=$(kubectl -n "$ns" get pod -l app="$service" -o jsonpath='{.items[0].status.podIP}')
  test -n "$ip"
  if kubectl -n "$ns" exec "$probe" -- nc -w 3 "$ip" "$port" </dev/null; then
    echo "Unexpected backend access to $service" >&2
    exit 1
  fi
  echo "Denied untrusted pod access to $service"
done
echo "Network policy enforcement observed; API endpoints remain reachable"

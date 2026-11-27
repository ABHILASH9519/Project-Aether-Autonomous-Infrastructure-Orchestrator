#!/usr/bin/env bash
# Run against an existing disposable kind cluster only.
set -euo pipefail
cluster="${1:-aether}"
registry=aether-lab-registry
port=5001
mapfile -t nodes < <(kind get nodes --name "$cluster")
test "${#nodes[@]}" -gt 0
# Fail on a conflicting name rather than deleting someone else's container.
docker run -d --name "$registry" -p "127.0.0.1:$port:5000" registry:2.8.3 >/dev/null
docker network connect kind "$registry"
for node in "${nodes[@]}"; do
  dir="/etc/containerd/certs.d/localhost:$port"
  docker exec "$node" mkdir -p "$dir"
  printf '[host."http://%s:5000"]\n' "$registry" | docker exec -i "$node" sh -c "cat > '$dir/hosts.toml'"
done
for i in {1..30}; do
  if curl -fsS "http://127.0.0.1:$port/v2/" >/dev/null; then exit 0; fi
  sleep 1
done
echo "local registry not ready" >&2
exit 1

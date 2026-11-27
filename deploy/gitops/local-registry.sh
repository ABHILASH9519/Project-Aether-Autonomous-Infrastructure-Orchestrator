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
registry_ip=$(docker inspect -f '{{.NetworkSettings.Networks.kind.IPAddress}}' "$registry")
test -n "$registry_ip"
for node in "${nodes[@]}"; do
  dir="/etc/containerd/certs.d/localhost:$port"
  docker exec "$node" mkdir -p "$dir"
  printf 'server = "http://%s:5000"\n[host."http://%s:5000"]\n  capabilities = ["pull", "resolve"]\n' "$registry_ip" "$registry_ip" | docker exec -i "$node" sh -c "cat > '$dir/hosts.toml'"
  docker exec "$node" containerd config dump | grep -F 'config_path' >&2
  docker exec "$node" cat "$dir/hosts.toml" >&2
done
for i in {1..30}; do
  if curl -fsS "http://127.0.0.1:$port/v2/" >/dev/null; then exit 0; fi
  sleep 1
done
echo "local registry not ready" >&2
exit 1

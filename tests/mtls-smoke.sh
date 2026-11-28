#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_TLS_DIR:?required}"
compose=(docker compose -f deploy/compose/events.yaml -f deploy/compose/mtls.yaml --profile workflow)
cid=$("${compose[@]}" ps -q fixture)
network=$(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' "$cid")
test -n "$network"
image=curlimages/curl:8.12.1
docker pull "$image"
probe(){
 local identity="$1"
 shift
 local mounts=()
 if [[ "$identity" != none ]]; then
  mounts=(-v "$AETHER_TLS_DIR/$identity.crt:/client.crt:ro" -v "$AETHER_TLS_DIR/$identity.key:/client.key:ro")
  set -- --cert /client.crt --key /client.key "$@"
 fi
 docker run --rm --pull never --network "$network" --user 0:0 -v "$AETHER_TLS_DIR/ca.crt:/ca.crt:ro" "${mounts[@]}" "$image" --silent --show-error --fail --max-time 5 --cacert /ca.crt "$@" https://fixture:8083/healthz
}
# Prove the exact probe path works before accepting negative results.
bash tests/execution-smoke.sh
probe executor
for identity in none untrusted; do
 rc=0
 probe "$identity" || rc=$?
 case "$rc" in
  35|55|56) echo "Rejected TLS identity: $identity (curl $rc)" ;;
  *) echo "Unexpected probe result for $identity: $rc" >&2; exit 1 ;;
 esac
done
echo "mTLS verified repair; missing and wrong client identities rejected"

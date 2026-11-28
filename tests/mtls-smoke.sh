#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_TLS_DIR:?required}"
compose=(docker compose -f deploy/compose/events.yaml -f deploy/compose/mtls.yaml --profile workflow)
cid=$("${compose[@]}" ps -q fixture)
network=$(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' "$cid")
test -n "$network"
probe(){
 docker run --rm --network "$network" --user 0:0 -v "$AETHER_TLS_DIR/ca.crt:/ca.crt:ro" -v "$AETHER_TLS_DIR/untrusted.crt:/client.crt:ro" -v "$AETHER_TLS_DIR/untrusted.key:/client.key:ro" curlimages/curl:8.12.1 --max-time 5 --cacert /ca.crt "$@" https://fixture:8083/healthz
}
# First establish successful authenticated repair so failed probes are
# not falsely attributed to a service that never started.
bash tests/execution-smoke.sh
if probe; then echo "unexpected unauthenticated TLS access" >&2; exit 1; fi
if probe --cert /client.crt --key /client.key; then echo "unexpected wrong client identity access" >&2; exit 1; fi
echo "mTLS verified repair; missing and wrong client identities rejected"

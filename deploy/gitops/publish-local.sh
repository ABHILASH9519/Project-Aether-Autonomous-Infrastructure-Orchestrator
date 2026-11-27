#!/usr/bin/env bash
set -euo pipefail
revision="${1:?source commit SHA required}"
[[ "$revision" =~ ^[a-f0-9]{40}$ ]]
test "$(git rev-parse HEAD)" = "$revision"
test -z "$(git status --porcelain --untracked-files=normal)"
repository=localhost:5001/aether/event-fabric
tag="$repository:$revision"
docker build --label "org.opencontainers.image.revision=$revision" -t "$tag" services/event-fabric >&2
docker push "$tag" >&2
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
accept='application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json'
curl -fsS --retry 5 --retry-all-errors -H "Accept: $accept" -D "$tmp/headers" -o "$tmp/manifest" "http://127.0.0.1:5001/v2/aether/event-fabric/manifests/$revision"
digest=$(tr -d '\r' < "$tmp/headers" | awk 'tolower($1)=="docker-content-digest:" {value=$2} END {print value}')
test "$digest" = "sha256:$(sha256sum "$tmp/manifest" | cut -d' ' -f1)"
[[ "$digest" =~ ^sha256:[a-f0-9]{64}$ ]]
printf '%s\n' "$digest"

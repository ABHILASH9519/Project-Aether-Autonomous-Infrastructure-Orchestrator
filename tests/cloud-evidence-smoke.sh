#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_EXECUTION_ID:?set ID used by execution-smoke.sh}"
[[ "$AETHER_EXECUTION_ID" =~ ^[a-zA-Z0-9_-]{1,128}$ ]]
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
aws_local() { "${compose[@]}" exec -T localstack awslocal "$@"; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsS -H "Authorization: Bearer $AETHER_OPERATOR_TOKEN" "http://127.0.0.1:8082/api/v1/incidents/$AETHER_EXECUTION_ID/job" > "$tmp/evidence.json"
jq -e '.status=="succeeded" and .evidence=="fixture content verified"' "$tmp/evidence.json" >/dev/null
key="jobs/$(printf '%s' "$AETHER_EXECUTION_ID" | sha256sum | cut -d' ' -f1).json"
digest=$(sha256sum "$tmp/evidence.json" | cut -d' ' -f1)
for i in {1..90}; do
  if aws_local s3 cp "s3://aether-lab-evidence/$key" - > "$tmp/download.json" 2>/dev/null && cmp -s "$tmp/evidence.json" "$tmp/download.json"; then
    stored=$("${compose[@]}" exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U aether -d aether -v "id=$AETHER_EXECUTION_ID" <<'SQL'
SELECT archive_sha256 FROM remediation_jobs WHERE correlation_id=:'id' AND archived_at IS NOT NULL;
SQL
)
    if [[ "$stored" == "$digest" ]]; then echo "Automatic verified S3 archive: $key"; exit 0; fi
  fi
  sleep 2
done
echo "Automatic archive not observed" >&2
exit 1

#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_OPERATOR_TOKEN:?set operator token}"
: "${AETHER_EXECUTION_ID:?set ID used by execution-smoke.sh}"
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
aws_local() { "${compose[@]}" exec -T localstack awslocal "$@"; }
ready=false
for i in {1..60}; do
  if aws_local s3api list-buckets >/dev/null 2>&1; then ready=true; break; fi
  sleep 2
done
test "$ready" = true
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsS -H "Authorization: Bearer $AETHER_OPERATOR_TOKEN" "http://127.0.0.1:8082/api/v1/incidents/$AETHER_EXECUTION_ID/job" > "$tmp/evidence.json"
jq -e '.status=="succeeded" and .evidence=="fixture content verified"' "$tmp/evidence.json" >/dev/null
bucket=aether-lab-evidence
if ! aws_local s3api head-bucket --bucket "$bucket" >/dev/null 2>&1; then
  aws_local s3api create-bucket --bucket "$bucket" >/dev/null
fi
key="jobs/$(printf '%s' "$AETHER_EXECUTION_ID" | sha256sum | cut -d' ' -f1).json"
aws_local s3 cp - "s3://$bucket/$key" --content-type application/json < "$tmp/evidence.json"
aws_local s3 cp "s3://$bucket/$key" - > "$tmp/download.json"
cmp "$tmp/evidence.json" "$tmp/download.json"
echo "Verified local S3 evidence round trip: $key"

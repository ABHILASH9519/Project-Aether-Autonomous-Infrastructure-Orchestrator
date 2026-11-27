#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_EXECUTION_ID:?use a completed synthetic execution ID}"
[[ "$AETHER_EXECUTION_ID" =~ ^[a-zA-Z0-9_-]{1,128}$ ]]
compose=(docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud)
pg() { "${compose[@]}" exec -T postgres "$@"; }
tmp=$(mktemp -d)
chmod 700 "$tmp"
restore_db="aether_restore_$(date +%s)_$$"
created=false
cleanup() {
  if [[ "$created" == true ]]; then pg dropdb -U aether --if-exists "$restore_db" || true; fi
  rm -rf "$tmp"
}
trap cleanup EXIT
snapshot() {
  pg psql -X -qAt -v ON_ERROR_STOP=1 -U aether -d "$1" -v "incident_id=$AETHER_EXECUTION_ID" <<'SQL'
SELECT jsonb_build_object(
  'incident', (SELECT to_jsonb(i) FROM incidents i WHERE correlation_id=:'incident_id'),
  'job', (SELECT to_jsonb(j) FROM remediation_jobs j WHERE correlation_id=:'incident_id'),
  'audit', (SELECT jsonb_agg(to_jsonb(a) ORDER BY version) FROM incident_audit a WHERE correlation_id=:'incident_id'),
  'events', (SELECT jsonb_agg(to_jsonb(b) ORDER BY b.event_id) FROM incident_inbox b JOIN incident_outbox o USING(event_id) WHERE o.correlation_id=:'incident_id'),
  'outbox', (SELECT jsonb_agg(to_jsonb(o) - 'published_at' ORDER BY id) FROM incident_outbox o WHERE correlation_id=:'incident_id')
);
SQL
}
snapshot aether > "$tmp/before.json"
jq -e '.incident.state=="resolved" and .job.status=="succeeded" and (.audit|length)==6 and (.events|length)>=1 and (.outbox|length)>=1' "$tmp/before.json" >/dev/null
pg pg_dump -U aether -d aether --format=custom --no-owner --no-acl > "$tmp/backup.dump"
test -s "$tmp/backup.dump"
pg createdb -U aether "$restore_db"
created=true
pg pg_restore -U aether -d "$restore_db" --exit-on-error --no-owner --no-acl < "$tmp/backup.dump"
snapshot "$restore_db" > "$tmp/after.json"
diff -u <(jq -S . "$tmp/before.json") <(jq -S . "$tmp/after.json")
echo "Logical PostgreSQL backup restored; incident, job, audit, inbox and outbox records match"

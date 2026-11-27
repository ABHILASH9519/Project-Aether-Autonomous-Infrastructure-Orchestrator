# PostgreSQL logical restore drill

This drill runs against a disposable local Compose stack containing synthetic events. It creates a custom-format pg_dump backup, restores it to a fresh database in the same PostgreSQL container, and compares a completed incident, remediation job, audit history, inbox and outbox records. The mutable outbox publication timestamp is excluded from the comparison.

```bash
export AETHER_EXECUTION_ID="restore-demo-$(date +%s)"
bash tests/execution-smoke.sh
bash tests/postgres-restore-smoke.sh
```

The source database is never overwritten. The temporary restore database and local dump are removed on normal exit or failure, if the shell EXIT trap runs. The test requires the existing local role tokens for execution-smoke.sh, Docker Compose, jq and Bash. Use only against the repository's disposable local environment; the local database role has creation privileges for this lab.

This verifies a logical backup and restore of selected records, not an off-site backup, measured recovery objectives, encryption, point-in-time recovery or a Kafka-consistent disaster recovery snapshot. Do not use real customer data. Verified in [run 46064464894](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46064464894).

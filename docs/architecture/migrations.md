# Schema migration ledger

PostgreSQL startup migrations are serialized with a transaction-scoped advisory lock. The migration ledger stores ordered version numbers and SHA-256 checksums. The idempotent historical schema is adopted as version 1 for existing databases without a ledger. DDL and the corresponding ledger insertion commit atomically; a failed migration rolls back.

Do not modify the embedded baseline schema.sql after publication. Add an ordered migration and its own embedded SQL file. A changed checksum, a missing earlier ledger entry or a newer schema fails startup instead of silently rewriting history.

Concurrent startup tests check that multiple replicas can migrate and that existing constraints are not recreated. This is an expand-only baseline for the current release, not a zero-downtime migration framework. No automatic downgrade or legacy data migration from emptyDir to PVC is provided. Back up existing data before schema upgrades.

Concurrent startup and stable constraint identity passed in [run 46064513027](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46064513027).

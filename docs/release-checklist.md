# Local release acceptance checklist

Before marking a versioned local release ready, manually run the Release candidate validation workflow on main. It dispatches the nine required checks at the exact current SHA and rejects missing, skipped, cancelled, failed or stale-revision runs. The workflow publishes a JSON report and run links. It never creates a release tag or publishes a container to an external registry.

Required checks: Control API; event fabric (Go race/vet/build, replay, approval, CLI, archive recovery, backup/restore, telemetry and measured load); configuration management; Kubernetes (storage, isolation and worker replacement); Argo CD (digest pull, sync and drift correction); supply chain (zero reported HIGH/CRITICAL application-image findings); fixture mTLS; Kafka broker failover; PostgreSQL primary failover.

The workflow requires Actions write permission to dispatch the checks with its repository token. Run it only on an isolated candidate on main and avoid commits until it finishes; otherwise the branch-head guard fails and the candidate must be revalidated. A green report is a necessary release criterion, not proof of production readiness. Review README, SECURITY.md and the scope/limitations before creating a release. No release has been tagged by this workflow.

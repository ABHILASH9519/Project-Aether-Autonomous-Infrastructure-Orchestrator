# Roadmap and acceptance criteria

Status key: ✅ implemented, ⏳ in progress, ⬜ planned.

## Phase 0 — Foundation
- ✅ Repository README and target architecture
- ✅ Architecture decision record and public API integration policy
- ⏳ Minimal Go control API, unit tests, container build and CI
- ⬜ CI pipeline green on GitHub

## Phase 1 — Event backbone
- ⬜ Kafka Compose profile and documented topics
- ⬜ Versioned event envelope with validation
- ⬜ Producer/consumer tests for duplicate and malformed messages
- ⬜ Bounded retries and DLQ

## Phase 2 — Durable incident state
- ⬜ PostgreSQL migrations
- ⬜ Incident state machine with optimistic concurrency
- ⬜ Transactional outbox/inbox and audit projections
- ⬜ End-to-end replay and recovery tests

## Phase 3 — Safety and execution
- ⬜ OPA rules and approval workflow
- ⬜ Dry-run remediation executor and idempotency
- ⬜ Kubernetes and Ansible adapters guarded by least privilege
- ⬜ Kill switch and audit evidence

## Phase 4 — Observability and delivery
- ⬜ Prometheus metrics and OpenTelemetry traces
- ⬜ Structured logs and Grafana dashboards
- ⬜ Helm chart and k3d/Kubernetes deployment
- ⬜ Security scan and reproducible chaos-test runbook

## Definition of done

Each milestone needs documentation, tests, local reproducibility, and a clear distinction between simulated and live capabilities.

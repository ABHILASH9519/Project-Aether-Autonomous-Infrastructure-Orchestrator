# AETHER
### Autonomous Event-driven Telemetry, Execution & Resilience

> **Status: Phase 3 — durable event processing (verification pending).** AETHER is an experimental, local-first DevOps platform. Automated remediation is intentionally restricted to explicitly approved, allowlisted actions; this repository is not production-ready.

**AETHER** is an open-source engineering project exploring how observable infrastructure events can become auditable, policy-gated operational workflows. It is built around a service-oriented control plane rather than a single dashboard.

## Why this exists

Modern infrastructure produces alerts, logs and events, but humans often still have to correlate them, select a runbook, execute it and verify the result. AETHER aims to make this process reproducible: **observe → normalize → correlate → decide → approve → execute → verify → audit**.

### Design principles
- **Safety before autonomy**: start in dry-run mode; no destructive or production actions by default.
- **Events are not commands**: Kafka delivers facts, while a policy-controlled orchestrator authorizes commands.
- **State is explicit**: PostgreSQL stores workflow/incident state, Kafka distributes events, Redis is optional for ephemeral coordination.
- **Adapters over vendor lock-in**: execution backends can target Kubernetes, Ansible, Docker or cloud-compatible APIs.
- **Local-first**: Docker Compose and k3d/kind are development targets; Floci/LocalStack are optional integration-test dependencies.
- **Evidence over claims**: all features are labelled implemented, planned, or experimental.

## System overview

![AETHER architecture](docs/architecture/aether-system.svg)

The visual below is the planned event-to-action pipeline; it does **not** imply the entire platform is implemented.


```mermaid
flowchart LR
    Operator[Operator / CLI] --> API[Control API - Go]
    API --> DB[(PostgreSQL)]
    API --> Bus[(Kafka)]
    Telemetry[OpenTelemetry / Prometheus] --> Detector[Detector]
    Detector --> Bus
    Bus --> Correlator[Incident Correlator]
    Correlator --> Policy[OPA Policy Gate]
    Policy -->|approved action| Engine[Workflow Engine]
    Engine --> Exec[Execution Adapters]
    Exec --> K8s[Kubernetes]
    Exec --> Ansible[Ansible]
    Exec --> Simulator[Safe Local Simulator]
    Exec --> Bus
    Bus --> Audit[Audit Projection]
    Audit --> DB
```

> Diagram describes **target architecture**, not a claim that all components are implemented.

## Initial repository layout

| Path | Purpose |
| --- | --- |
| `services/control-api/` | Go HTTP API and health endpoints |
| `docs/architecture/` | Architecture diagrams and decisions |
| `docs/roadmap.md` | Milestones and acceptance criteria |
| `contracts/events/` | Versioned event schemas |
| `deploy/compose/` | Local development infrastructure |
| `.github/workflows/` | CI automation |

## Planned implementation sequence

1. **Foundation** — documentation, architecture decisions, Go control API, tests and CI.
2. **Event fabric** — Kafka with structured event contracts, consumer idempotency and retry/DLQ semantics.
3. **Incident lifecycle** — PostgreSQL persistence, state machine, audit projection and correlation.
4. **Policy & execution** — authorization, dry-run engine, allowlisted adapters and approvals.
5. **Observability** — OpenTelemetry, Prometheus, Grafana, logging and SLOs.
6. **Delivery & resilience** — Kubernetes/Helm, GitOps, failure injection and measured recovery.

See [roadmap](docs/roadmap.md) and [architecture](docs/architecture/overview.md).

## Languages & technologies

Core services use **Go** (with room for Java/C++ adapters). Infrastructure uses Docker, Kubernetes, Kafka, PostgreSQL, OpenTofu/Terraform and Ansible as milestones are implemented. Optional local cloud emulators are **not prerequisites** for development.

## Safety and scope

Do not expose this prototype directly to the public internet. Credentials, access controls, persistence, policy enforcement and execution isolation require hardening. AETHER does not run privileged remediation automatically without explicit policy and operator approval.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Architecture decisions should be documented before changing cross-service contracts.

## License

Apache License 2.0 — see [LICENSE](LICENSE).

## Event fabric preview (Phase 2)

The Go ingestion and consumer processes, Kafka Compose environment and smoke test are available on main. Runtime verification is pending. See [event fabric architecture and local run instructions](docs/architecture/event-fabric.md).

### Event ingestion and processing flow

```mermaid
flowchart TD
    A[Infrastructure / Simulated Alert] -->|HTTP POST| B[Telemetry Ingestion Service - Go]
    B --> C{Event Validation}
    C -->|Invalid| D[HTTP Rejection]
    C -->|Valid| E[Kafka Event Topic]
    E --> F[Incident Consumer - Go]
    F --> G[Structured Incident Output]
    F -->|Malformed Event| H[Dead Letter Topic]
    G -. Future phase .-> I[(PostgreSQL)]
```

### Kafka architecture

![Kafka architecture — telemetry ingestion, three partitions, incident consumer and dead-letter queue](docs/architecture/kafka-event-fabric.svg)

*Implemented event-processing design. Runtime verification pending.*

Events are routed using their correlation IDs as partition keys. Malformed broker messages are sent to a dead-letter topic before their consumer offsets are committed.

### Overall AETHER development roadmap

| Phase | Milestone | Status |
| --- | --- | --- |
| 01 | Repository, architecture and control API | Committed |
| 02 | Kafka event fabric | Committed · testing pending |
| 03 | PostgreSQL incident management | Inbox, projection and relay committed; testing pending |
| 04 | Redis caching and coordination | Planned |
| 05 | Workflow orchestration and policy engine | Planned |
| 06 | Ansible and Puppet automation | Planned |
| 07 | Floci / LocalStack cloud adapters | Planned |
| 08 | Kubernetes, Helm and GitOps | Planned |
| 09 | Prometheus, Grafana and OpenTelemetry | Planned |
| 10 | Chaos engineering and security | Planned |

## Durable event processing

```mermaid
sequenceDiagram
 participant C as Incident consumer
 participant P as PostgreSQL
 participant R as Outbox relay
 participant K as Kafka
 C->>P: Atomic inbox + incident + outbox
 P-->>C: Commit
 R->>P: Claim pending row (SKIP LOCKED)
 R->>K: Publish observed event
 K-->>R: Acknowledge
 R->>P: Mark published and commit
```

**Guarantee:** at-least-once delivery with transactional inbox deduplication. No end-to-end exactly-once claim. See [failure model](docs/architecture/outbox-relay.md). Runtime verification is pending.

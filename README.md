# AETHER
### Autonomous Event-driven Telemetry, Execution & Resilience

> **Status: Phase 4/5 prototype — durable events, admission control and approval simulation.** AETHER is an experimental, local-first DevOps platform. Automated remediation is intentionally restricted to explicitly approved, allowlisted actions; this repository is not production-ready.

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
    G --> I[(PostgreSQL)]
```

### Kafka architecture

![Kafka architecture — telemetry ingestion, three partitions, incident consumer and dead-letter queue](docs/architecture/kafka-event-fabric.svg)

*Implemented event-processing design. Runtime verification pending.*

Events are routed using their correlation IDs as partition keys. Malformed broker messages are sent to a dead-letter topic before their consumer offsets are committed.

### Overall AETHER development roadmap

| Phase | Milestone | Status |
| --- | --- | --- |
| 01 | Repository, architecture and control API | Committed |
| 02 | Kafka event fabric | Prior end-to-end CI passed |
| 03 | PostgreSQL incident management | Prior PostgreSQL and workflow integration CI passed |
| 04 | Redis caching and coordination | Shared admission budget; prior integration CI passed |
| 05 | Workflow orchestration and policy engine | OPA verified; durable local remediation worker committed; new CI pending |
| 06 | Ansible and Puppet automation | Isolated convergence/drift lab committed; workflow adapter pending |
| 07 | Floci / LocalStack cloud adapters | Planned |
| 08 | Kubernetes, Helm and GitOps | Local Helm chart, kind CI and Argo CD manifest committed; verification pending |
| 09 | Prometheus, Grafana and OpenTelemetry | HTTP metrics and dashboard verified; HTTP-to-Kafka OTLP tracing committed |
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

## Approval-gated workflow (local simulation)

```mermaid
stateDiagram-v2
 [*] --> open
 open --> acknowledged: acknowledge
 acknowledged --> awaiting_approval: request_approval
 awaiting_approval --> approved: approve
 awaiting_approval --> rejected: reject
 approved --> executing: simulate
 executing --> verifying: simulated check
 verifying --> resolved: simulated completion
```

The opt-in workflow API uses distinct operator/approver bearer tokens, OPA authorization, version checks and atomic audit inserts. Simulation does not perform any infrastructure changes or prove real recovery. See [workflow architecture and limitations](docs/architecture/workflow.md).

```bash
export AETHER_OPERATOR_TOKEN="$(openssl rand -hex 32)"
export AETHER_APPROVER_TOKEN="$(openssl rand -hex 32)"
docker compose -f deploy/compose/events.yaml --profile workflow up -d --build
bash tests/event-fabric-smoke.sh
bash tests/workflow-smoke.sh
```

**Verification:** [GitHub Actions run 45892838506](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/45892838506) passed the prior Kafka/PostgreSQL/Docker workflow baseline. [Run 45893095275](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/45893095275) also passed Redis and monitoring checks. OPA authorization and recovery also passed in run 45894817663. No production deployment or real remediation is supported yet.

## Admission control and observability

```mermaid
flowchart LR
 Client --> Ingest[Go ingestion API]
 Ingest --> Redis[(Redis shared request budget)]
 Redis -->|Allowed| Kafka[(Kafka)]
 Redis -->|Exhausted| Reject[429]
 Redis -->|Unavailable| Unavailable[503]
 Kafka --> Consumer --> PostgreSQL[(PostgreSQL)]
 Ingest -. HTTP metrics .-> Prometheus
 Workflow -. HTTP metrics .-> Prometheus
 Prometheus --> Grafana
```

```bash
export AETHER_OPERATOR_TOKEN="$(openssl rand -hex 32)"
export AETHER_APPROVER_TOKEN="$(openssl rand -hex 32)"
docker compose -f deploy/compose/events.yaml --profile workflow --profile observability up -d --build
```

- Redis admission: 120 requests per fixed 60-second window shared by ingestion replicas; fail closed on Redis failure.
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (local anonymous viewer, provisioned AETHER HTTP operations dashboard)
- Metrics: http://localhost:8081/metrics and http://localhost:8082/metrics

See [admission design](docs/architecture/admission.md) and [observability design](docs/architecture/observability.md). Metrics do not contain request IDs or payloads. The optional monitoring profile is local only; prior monitoring smoke tests passed. HTTP-to-Kafka OpenTelemetry tracing is now optional; worker metrics and alerts remain planned.

## Policy-based authorization

```mermaid
flowchart LR
 Operator --> Workflow[Go Workflow API]
 Approver --> Workflow
 Workflow --> OPA[Open Policy Agent]
 OPA -->|Allow| DB[(Version-checked transaction)]
 OPA -->|Deny / error| Rejected[403 / 503]
```

Operators request approval and simulate; approvers approve or reject. OPA denies unknown actions and live execution. See [policy architecture](docs/architecture/policy.md). Two distinct tokens are required for the workflow profile. [Policy and outage integration CI passed](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/45894817663). Audit rows record the server-derived role. See the [local OPA outage drill](docs/runbooks/policy-outage.md) for fail-closed behavior and recovery.

## Durable lab execution

```mermaid
flowchart LR
 Approved -->|execute_lab| Transaction[Atomic incident + job insert]
 Transaction --> Jobs[(PostgreSQL job queue)]
 Jobs --> Worker[Go executor]
 Worker -->|PUT fixed health marker| Fixture[Isolated Go fixture]
 Worker -->|GET verify contents| Fixture
 Worker -->|verified / retries exhausted| Audit[(Incident + audit)]
```

The `execute_lab` action queues an approved job. A separate worker applies an idempotent repair to an isolated synthetic fixture and independently verifies the contents. Three recorded attempts are allowed. Fixture state is ephemeral; this is not real production infrastructure remediation. See [execution architecture](docs/architecture/execution.md). Inspect job evidence via authenticated `GET /api/v1/incidents/{id}/job`. Run `bash tests/execution-smoke.sh` with the workflow profile. New integration verification pending.

## Configuration management lab

Ansible and Puppet enforce a version-controlled desired configuration, demonstrate idempotence, and correct injected drift inside a disposable container. See [architecture and run instructions](docs/architecture/configuration-management.md). This standalone lab is not yet connected to the approved execution worker.

## Kubernetes deployment

A local Helm chart deploys the six event-fabric processes and their dependencies; a kind CI job checks the approved workflow and lab repair. An optional manual-sync Argo CD Application is included. See [Kubernetes architecture and setup](docs/architecture/kubernetes.md). This chart uses disposable storage and locally loaded images, not production persistence or image promotion.

## Distributed tracing

Set `AETHER_OTLP_ENDPOINT=http://jaeger:4318/v1/traces` before starting the observability profile. Jaeger listens on http://localhost:16686. W3C trace headers connect HTTP ingestion to Kafka consumer spans. See [trace architecture](docs/architecture/tracing.md). CI verification pending; outbox and execution spans are not yet connected.

# AETHER project showcase

## What to demonstrate

AETHER converts synthetic infrastructure events into durable, policy-controlled remediation workflows. The demo uses an explicit approval and two distinct local role credentials, verifies a synthetic repair, and reads its ordered audit history. The system does not make autonomous production changes.

1. Show the [current system architecture](architecture/visuals/current-system.svg).
2. Start the disposable local lab with the [operator demo guide](runbooks/operator-demo.md).
3. Show the actual terminal transcript of the blocked premature execution, approval, resolved job and audit history.
4. Open the provisioned Grafana dashboard and Jaeger trace view; avoid displaying credentials.
5. Show [GitOps self-healing](architecture/visuals/gitops-self-healing.svg), the [measured CI latency chart](architecture/visuals/measured-ci-latency.svg), and the linked passing builds.

## Verified engineering evidence

- [Event fabric and CI safeguards](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067786876)
- [Argo CD digest-pinned deployment and self-healing](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067787051)
- [Kubernetes persistent storage and network policy](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067477316)
- [Go 1.26.11 supply-chain report](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067342671)
- [Measured 40-event smoke workload](benchmarks/ci-46066048451.json): all 40 accepted and projected; p95 13.81 ms on one CI runner.

These are historical successful runs at specific revisions, not a claim that every later commit is verified. The [CLI, audit endpoint and automatic S3 outage recovery passed](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46131801803). Remaining vulnerabilities, production identities, TLS, high availability, and coordinated disaster recovery are outside the verified scope.

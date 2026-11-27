# CI failure analysis and recovery

Do not rerun a failed historical commit to validate a source fix: GitHub Actions reruns the original source revision. Inspect the newest main revision and its workflow results instead.

| Failure | Evidence | Corrective action |
| --- | --- | --- |
| Duplicate telemetry imports and declarations | [Failed event fabric 46067227955](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067227955) | Commit 42fc290 restores the original complete tracing tree without rewriting history. [Corrected event fabric passed](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067281907). |
| OCI registry manifest negotiation | [Failed GitOps 46065730153](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46065730153) | Accept indexes and manifests and verify SHA-256 bytes in d3319a0. |
| Node image pull resolves registry to loopback | [Failed GitOps 46066048455](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46066048455) | Configure containerd CRI registry hosts and preflight digest pulls in 0bfd916. [Preflight succeeded; full workflow status must be checked](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46067477513). |

The event fabric integration job now depends on its Go compile, race, vet and dependency verification job. This avoids redundant Docker builds when source is already known to fail. It can also be manually dispatched. GitOps failure diagnostics collect controller and repository-server logs, pod descriptions, registry logs and node registry mapping. Tests are not bypassed; diagnostics are best effort so the original failure stays visible.

When a connector returns an ambiguous error during a write, read the remote main HEAD and inspect the changed tree before retrying. Use an expected-parent compare-and-swap and never force-push. Record the new commit and exact CI run rather than declaring old red checks green.

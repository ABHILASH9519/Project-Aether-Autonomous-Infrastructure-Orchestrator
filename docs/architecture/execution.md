# Durable lab remediation

The operator may enqueue the single built-in lab repair only after approval. The incident transition to executing and job insertion commit together. A worker claims pending jobs using row locks and SKIP LOCKED, invokes a bounded idempotent adapter, and records verified success or terminal failure. The job has at most three recorded attempts; crashes before database commit may replay an external effect without incrementing the counter. No arbitrary commands or user-defined endpoints are accepted.

The executor and fixture are separate Go processes. The fixture writes a fixed health marker using atomic rename in an isolated tmpfs and independently serves its contents for verification. No host mount, shell, Docker socket, or arbitrary endpoint is provided. The fixture is deliberately ephemeral; its restart loses the marker and historical success does not guarantee ongoing health. The worker records only a bounded outcome, not response contents. No production execution is supported.

## Validation

Run `bash tests/execution-smoke.sh` against the local workflow profile. The smoke test requires both distinct role tokens, curl and jq. It verifies approval cannot be bypassed and polls for a verified resolved incident. Database integration tests cover success and retry exhaustion. The worker holds database locks across bounded network operations; it does not claim exactly-once effects. The fixture endpoint is unauthenticated on the private Compose network and is not host-published. CI results are pending.

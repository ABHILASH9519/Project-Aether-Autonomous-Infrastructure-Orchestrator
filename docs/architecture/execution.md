# Durable lab remediation

The operator may enqueue the single built-in lab repair only after approval. The incident transition to executing and job insertion commit together. A worker claims pending jobs using row locks and SKIP LOCKED, invokes a bounded idempotent adapter, and records verified success or terminal failure. The job has at most three recorded attempts; crashes before database commit may replay an external effect without incrementing the counter. No arbitrary commands or user-defined endpoints are accepted.

The storage implementation is an intermediate milestone; the worker and isolated fixture target are added separately. No production execution is supported.

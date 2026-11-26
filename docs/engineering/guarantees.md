# Engineering guarantees and non-guarantees

| Property | Mechanism | Verification |
| --- | --- | --- |
| Atomic incident update | PostgreSQL transaction | Integration test authored |
| Duplicate event suppression | Unique inbox key | Integration test authored |
| Failed publish leaves pending row | Transaction rollback | Integration test authored |
| Concurrent row claims | FOR UPDATE SKIP LOCKED | Concurrency test pending |
| End-to-end exactly once | Not provided | Downstream dedup required |
| Per-incident relay ordering | Not guaranteed with multiple workers | Design work pending |
| Authentication | Not implemented | Local-only deployment |

## Operational constraints

Use synthetic events. Never deploy the unauthenticated API publicly. The local Compose stack uses a single Kafka broker and development database credentials. No real remediation is exposed. Test results must come from actual CI or local execution, not from this document.

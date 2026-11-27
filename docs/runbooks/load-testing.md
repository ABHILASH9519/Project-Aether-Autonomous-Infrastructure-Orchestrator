# Local load and projection verification

Use a fresh disposable Compose stack with the workflow profile and both role tokens. Wait for readiness before running the Go load generator:

```sh
cd services/event-fabric
go run ./cmd/loadtest -requests 40 -concurrency 4
```

The generator targets loopback only, caps concurrency at 16 and total requests at 100, creates one unique correlation ID, and waits for every accepted event to appear in the incident projection. It emits JSON with actual accepted response latency (nearest-rank p50/p95/p99), wall time, transport errors, other statuses and projected count. The command fails unless every request is accepted and projected once.

The local Redis cap is 120 requests per fixed 60-second window, so allow the window to expire before repeating a run. The test is a bounded smoke benchmark, not saturation testing, a capacity estimate, a production SLO, or a cross-hardware comparison. CI runs a small sample after the normal smoke suite and stores the report as a workflow artifact. No invented performance data is included. CI verification pending.

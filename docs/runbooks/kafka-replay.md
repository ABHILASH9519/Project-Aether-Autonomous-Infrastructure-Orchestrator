# Kafka replay and concurrency verification

The local replay drill submits a synthetic envelope, waits for the incident projection, stops the consumer, publishes the identical event ID again, restarts the consumer and waits for explicit duplicate suppression. It verifies that the incident event count and inbox/outbox row counts remain one.

Run only in a disposable local Compose stack with the workflow profile and exported operator token:

```bash
bash tests/kafka-replay-smoke.sh
```

This is a duplicate-record/restart test, not a forced offset reset or a crash precisely between database commit and offset commit. It does not assert exactly-once broker delivery. The separate PostgreSQL concurrency test races 32 writes representing eight distinct event IDs and checks the projection and outbox cardinality. New CI verification pending.

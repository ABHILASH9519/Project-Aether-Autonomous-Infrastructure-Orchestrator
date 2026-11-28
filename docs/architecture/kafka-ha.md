# Replicated Kafka local deployment

The optional Compose overlay creates three combined KRaft controllers/brokers, separate data volumes, replication factor 3, minimum in-sync replicas 2 and clean leader election only. Applications use three bootstrap addresses and the existing producer requires all in-sync replicas. A single broker can be unavailable while two remain in sync.

```bash
# Export both distinct local role tokens first. Stop any default lab using ports 8081/8082.
docker compose -p aether-ha -f deploy/compose/events.yaml -f deploy/compose/kafka-ha.yaml --profile workflow up -d --build
bash tests/kafka-ha-smoke.sh
docker compose -p aether-ha -f deploy/compose/events.yaml -f deploy/compose/kafka-ha.yaml --profile workflow down -v
```

Use a new isolated Compose project: never attach a previous single-broker volume to the new KRaft quorum. The drill verifies three in-sync replicas, completes a real synthetic incident, stops the actual leader of partition zero, waits for two in-sync replicas, completes a second approved repair, checks the first incident remains unchanged, then restores all three replicas. The test allows bounded publication retries using a stable event ID.

This is broker-process fault tolerance on one physical host, not node/zone disaster recovery. PostgreSQL, Redis, OPA and the fixture remain single instances in this profile. Plaintext Kafka is private to the local network. [Broker-leader failover CI passed](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46153821637).

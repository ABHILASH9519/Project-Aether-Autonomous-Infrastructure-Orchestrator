# Application replica scaling

Helm exposes 1–5 replicas for the ingestion, consumer, relay, workflow and executor Deployments. The fixture remains a singleton because it owns ephemeral local files. Kafka group coordination assigns partitions among consumers; PostgreSQL row locks and SKIP LOCKED protect concurrent relay and executor claims. External effects remain at-least-once across crashes.

The Kubernetes recovery test scales consumer, relay and executor to two replicas, deletes one pod in each pair, runs the approved execution smoke and verifies the incident reaches resolved version 6 with one event. It then confirms replacement replicas become ready and restores the original replica counts. It is a pod-replacement drill, not a proof of zero downtime or guaranteed failover at the exact commit boundary.

Kafka and PostgreSQL remain single-instance stateful dependencies; full infrastructure high availability and multi-node disaster recovery are not implemented. CI pending.

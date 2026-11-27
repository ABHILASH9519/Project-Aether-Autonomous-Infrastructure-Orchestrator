# Kubernetes persistence recovery drill

Use only a disposable local kind cluster running the AETHER Helm chart with default PVCs. The test deliberately replaces the PostgreSQL and Kafka pods while retaining their claims. It first creates an approved synthetic repair, captures its incident and job evidence plus the three Kafka input partition end offsets, restarts both deployments, checks that pod UIDs changed, then compares the saved state and offsets.

```sh
bash tests/kubernetes-persistence-smoke.sh
```

Both role tokens and kubectl, curl, jq, Bash are required. Run after the namespace and Helm chart have been initialized. Do not run during other ingestion traffic; the offset comparison assumes no other input producers. This tests single-node pod recovery, not disk/node loss or high availability. It does not prove every Kafka message byte was retained. Fixture state is ephemeral independently of incident/job state. CI verification pending.

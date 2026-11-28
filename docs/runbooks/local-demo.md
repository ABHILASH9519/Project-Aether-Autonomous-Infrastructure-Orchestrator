# Local demo runbook

Prerequisites: Go (CI uses 1.26.11) for the operator CLI and Docker with Compose v2, Bash, curl, jq, OpenSSL, Make and a Linux-compatible shell. First build downloads public images and dependencies. No paid cloud account is needed; local hardware and network access are required.

```bash
export AETHER_OPERATOR_TOKEN="$(openssl rand -hex 32)"
export AETHER_APPROVER_TOKEN="$(openssl rand -hex 32)"
make up
make ready
make demo
make smoke
make automation
make down
```

The smoke suite verifies the event ingestion and durable workflow, the separate lab execution worker, S3 evidence export, monitoring scrapes and the HTTP-to-Kafka trace. Run the optional OPA and Redis outage drills separately. The automation target uses a standalone container. Kubernetes requires kind, kubectl and Helm; see the [Kubernetes guide](../architecture/kubernetes.md).

Do not use `docker compose down -v` unless you intend to destroy the local database and Kafka volume. The local cloud and fixture services are ephemeral. Never publish these development ports to the internet.

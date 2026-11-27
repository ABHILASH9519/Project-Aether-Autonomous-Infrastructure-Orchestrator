# Local Kubernetes and GitOps

The Helm chart deploys the six event-fabric processes, Kafka, Redis, PostgreSQL, OPA and a topic initialization job into a dedicated namespace. Application containers drop capabilities and do not mount service account tokens; HTTP services have liveness/readiness probes. No host-published services or cluster-admin permissions are needed by the applications.

```mermaid
flowchart TB
  Git[Git main] --> Argo[Argo CD Application]
  Git --> Helm[Helm chart]
  Argo --> K[Local Kubernetes namespace]
  Helm --> K
  K --> Apps[Six Go deployments]
  K --> Infra[Kafka / Redis / PostgreSQL / OPA]
```

```sh
kind create cluster --name aether
docker build -t aether-event-fabric:local services/event-fabric
kind load docker-image aether-event-fabric:local --name aether
kubectl create namespace aether
export AETHER_OPERATOR_TOKEN="$(openssl rand -hex 32)"
export AETHER_APPROVER_TOKEN="$(openssl rand -hex 32)"
DB_PASSWORD="$(openssl rand -hex 16)"
kubectl -n aether create secret generic aether-local \
  --from-literal=postgres-password="$DB_PASSWORD" \
  --from-literal=database-url="postgres://aether:$DB_PASSWORD@postgres:5432/aether?sslmode=disable" \
  --from-literal=operator-token="$AETHER_OPERATOR_TOKEN" \
  --from-literal=approver-token="$AETHER_APPROVER_TOKEN"
helm upgrade --install aether deploy/helm/aether -n aether
bash tests/kubernetes-smoke.sh
```

The chart is a local lab with PostgreSQL and Kafka using separate ReadWriteOnce PVCs by default. A dynamic default StorageClass is required (kind provides one). PVCs are marked for retention during Helm uninstall and Argo CD pruning/deletion; namespace deletion can still remove claims and backing volumes. Use persistence.enabled=false only for a fresh disposable installation. Changing an existing installation from emptyDir to PVC does not migrate data. The broker remains single-node; PVC retention is not backup, disaster recovery or HA. If the Kafka volume is lost, the topic initialization Job must be recreated. Network-policy enforcement and TLS remain outside this milestone. The optional Argo CD Application enables automatic sync and self-healing in the dedicated local namespace. A separate workflow installs the controller and tests reconciliation; see [GitOps guide](gitops.md). The image is locally loaded with pullPolicy Never. GitOps requires preloading a matching image and creating the secret before sync. Direct Helm deployment and execution passed in [run 45898190027](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/45898190027). Argo CD controller sync and replica self-healing passed in [run 45901613370](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/45901613370).

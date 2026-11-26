# Local Kubernetes and GitOps

The Helm chart deploys the six event-fabric processes, Kafka, Redis, PostgreSQL, OPA and a topic initialization job into a dedicated namespace. Application containers drop capabilities and do not mount service account tokens; HTTP services have liveness/readiness probes. No host-published services or cluster-admin permissions are needed by the applications.

```mermaid
flowchart TB
  Git[Git main] --> Argo[Optional Argo CD Application]
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

The chart is a disposable local lab: Kafka and PostgreSQL use emptyDir, so pod replacement loses state; no HA, PVCs, network-policy enforcement or TLS. Topic creation is a one-time Job, requiring rerun if the broker's data is lost. The optional Argo CD Application is a manual-sync manifest, not an installed controller or automated image promotion pipeline. The image is locally loaded with pullPolicy Never. GitOps requires preloading a matching image and creating the secret before sync. Kubernetes CI is pending.

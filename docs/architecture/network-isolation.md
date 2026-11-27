# Kubernetes ingress isolation

The Helm chart defaults to a namespace-wide deny-ingress policy. Ingest and workflow APIs permit inbound traffic to their HTTP ports. Internal backends permit same-namespace pods with explicitly allowed application labels, on their specific service ports:

| Backend | Allowed client labels |
| --- | --- |
| PostgreSQL | consumer, relay, workflow, executor |
| Kafka | ingest, consumer, relay, kafka-topics, kafka |
| Redis | ingest |
| OPA | workflow |
| Fixture | executor |

This requires a NetworkPolicy-capable CNI. The kind workflow uses the same pinned kind/Kubernetes versions as GitOps, and a separate test creates an unlabeled client pod, requires successful API requests, and checks that TCP connections to ready backend pod IPs are denied. The existing approved execution smoke tests exercise allowed internal connections.

Policies are ingress-only. Egress remains unrestricted, public API access is not authenticated at the network layer, and labels are not cryptographic workload identities. Node-originated traffic and kubectl port forwarding are not controlled by these policies. Restrict namespace write permissions in real deployments; a user who can create pods with trusted labels can bypass this model. Do not claim isolation if the CNI does not enforce NetworkPolicy. For an unsupported local CNI, explicitly set networkPolicy.enabled=false; this disables isolation. CI verification pending.

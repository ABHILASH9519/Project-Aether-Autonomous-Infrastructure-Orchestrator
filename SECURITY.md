# Security policy

AETHER is an integrated local engineering platform, not a production security product. Do not expose the development stack to the public internet. Never submit real credentials or sensitive incident data to demo endpoints.

## Scope and known limitations

- Shared role tokens rather than individual identities; optional executor-to-fixture mutual TLS, but no cluster-wide TLS or external identity provider.
- Development Kafka, Redis, OPA and fixture services are private to the container network but do not provide independent service authentication.
- The execution worker accepts only fixed synthetic fixture operations; it has no host mounts or Docker socket.
- Kubernetes has single-node persistent volumes and CI-verified ingress policies. No egress restrictions or cluster-wide workload identity are implemented.
- The optional S3 emulator and Grafana anonymous viewer are for local development only.

Report security issues privately to the repository owner via GitHub's private vulnerability reporting if enabled. Do not publish exploitable secrets or personal data in issues. Production deployments require a separate threat model, credential management, dependency audit, least-privilege service accounts and isolation testing.

The event-fabric image scan at revision d7bc25c reported zero HIGH/CRITICAL findings. CI now enforces that severity threshold for this image only; this is not a guarantee of security and does not cover infrastructure or automation images.

# Security policy

AETHER is a local engineering prototype, not a production security product. Do not expose the development stack to the public internet. Never submit real credentials or sensitive incident data to demo endpoints.

## Scope and known limitations

- Shared role tokens rather than individual identities; no TLS or external identity provider.
- Development Kafka, Redis, OPA and fixture services are private to the container network but do not provide independent service authentication.
- The execution worker accepts only fixed synthetic fixture operations; it has no host mounts or Docker socket.
- Kubernetes uses disposable storage and no tested network-policy enforcement.
- The optional S3 emulator and Grafana anonymous viewer are for local development only.

Report security issues privately to the repository owner via GitHub's private vulnerability reporting if enabled. Do not publish exploitable secrets or personal data in issues. Production deployments require a separate threat model, credential management, dependency audit, least-privilege service accounts and isolation testing.

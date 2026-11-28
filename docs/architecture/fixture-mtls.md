# Fixture mutual TLS

The optional Compose overlay protects the executor-to-fixture repair and verification path with TLS 1.3, a dedicated CA, certificate hostname validation and the exact executor URI SAN. Missing certificates, a wrong client identity, a partial TLS configuration or an HTTPS-to-HTTP downgrade fail closed. The base local lab remains plaintext unless this overlay is selected. This is not mesh-wide TLS.

```sh
export AETHER_TLS_DIR="$(mktemp -d)"
bash tests/generate-lab-certs.sh
docker compose -f deploy/compose/events.yaml -f deploy/compose/mtls.yaml --profile workflow up -d --build
bash tests/mtls-smoke.sh
```

Export both operator/approver tokens first. The disposable CA and its keys stay in a private directory; individual mounted key files are readable by the non-root container UID. Never publish them or reuse them outside this isolated lab. The test executes a real approved synthetic repair and verifies that an unauthenticated client and a CA-signed wrong identity cannot access the fixture. Clean up the Compose project and generated directory afterward. No production certificate rotation or service mesh is provided. CI pending.

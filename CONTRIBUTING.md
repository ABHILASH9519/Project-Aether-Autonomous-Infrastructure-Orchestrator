# Contributing

Thank you for contributing to AETHER.

1. Open an issue describing intended behavior, relevant threat model and acceptance tests.
2. Use a small feature branch; do not combine architecture changes with unrelated refactors.
3. For Go changes, run `gofmt -w .`, `go vet ./...` and `go test -race ./...` inside the service.
4. Update API/event contracts and ADRs whenever changing a public interface.
5. Prefer local fixtures and mocks over live cloud or public API dependencies.
6. Do not commit credentials, customer data, or unsafe unapproved remediation.
7. Mark unfinished capabilities **planned**; do not claim production readiness.

## Development

```bash
cd services/control-api
go run ./cmd/server
# In another terminal:
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/readyz
curl -fsS http://localhost:8080/api/v1/info
```

Alternatively: `docker compose -f deploy/compose/compose.yaml up --build`.

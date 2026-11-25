# Optional public API integrations

Reference catalog: https://github.com/public-apis/public-apis

The catalog is a curated collection, **not** a guarantee that every API is available, anonymous, stable, licensed for every use, or free at unlimited volume. Confirm each API's current terms, authentication, quotas, privacy constraints and expected response shape before integration.

## Integration policy

- Public API access is **optional**; core local scenarios use fixtures.
- External responses are untrusted inputs and never directly trigger privileged remediation.
- Use per-provider adapter interfaces with request timeouts, rate limiting and circuit breaking.
- Do not commit credentials; use environment variables or a local secrets store.
- Avoid ingesting personal or sensitive data; redact logs.
- Prefer deterministic mock servers for CI; record response-schema contract tests.
- Record source attribution and access timestamps for external-derived signals.

## Candidate demo scenarios

1. **Weather-status probe** (later): poll a permitted weather data service to simulate external dependency latency or outage; emit `ExternalDependency.Degraded` events.
2. **Public status/health probe** (later): collect permitted availability signals and visualize the difference between external API failure and internal failure.
3. **Reference data adapter** (later): demonstrate schema drift and retry policies, without coupling incident control to public data.

These are proposals, not implemented integrations. API choices remain subject to current access and licensing verification.

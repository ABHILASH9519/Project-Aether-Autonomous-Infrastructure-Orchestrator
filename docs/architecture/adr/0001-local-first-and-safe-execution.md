# ADR 0001 — Local-first implementation and safe execution boundaries

- **Status:** Accepted
- **Date:** 2026-11-26

## Context

AETHER must be easy to evaluate without cloud bills or privileged infrastructure access. Automated remediation has a high blast radius.

## Decision

Start with Go services, Docker Compose for the local environment, and a dry-run execution simulator. Introduce Kafka and PostgreSQL as independently testable services before adding infrastructure adapters. Run live remediation only through scoped, audited, allowlisted commands behind an explicit approval policy; production is never auto-approved by default.

## Consequences

- CI can run deterministic tests without cloud credentials.
- Rich infrastructure integrations arrive later.
- The project emphasizes correctness and reproducibility over an appearance of autonomy.

# AETHER operator demo

Use a fresh local lab, with Go, Docker Compose, Make, OpenSSL, curl and jq available.

```bash
export AETHER_OPERATOR_TOKEN="$(openssl rand -hex 32)"
export AETHER_APPROVER_TOKEN="$(openssl rand -hex 32)"
make up
make ready
make demo
```

The Go CLI sends one synthetic alert to the local ingestion API, waits for its incident, verifies premature execution is rejected, performs the four authorized transitions, and waits for independently verified fixture repair. It prints the six ordered audit entries. The explicit -approve-lab flag and both distinct credentials are required; this demonstration uses both roles from one local terminal and is not individual-person separation of duties. There are no arbitrary targets, commands or remote URLs.

For recording, show the terminal output, Grafana at http://localhost:3000 and Jaeger at http://localhost:16686. Do not record credentials. Stop the lab with make down. This does not delete named volumes. CI runs this CLI against the actual local stack and uploads its transcript. [Verified CLI and audit integration](https://github.com/ABHILASH9519/Project-Aether-Autonomous-Infrastructure-Orchestrator/actions/runs/46131801803).

SHELL := /bin/bash
COMPOSE := docker compose -f deploy/compose/events.yaml --profile workflow --profile observability --profile cloud
.PHONY: help up down logs test smoke automation helm-lint load demo
help:
	@printf '%s\n' 'make up         start local full-stack demo (export both tokens first)' 'make smoke      exercise ingestion, approvals, execution, monitoring and cloud' 'make test       run Go tests (database integrations require test service env)' 'make automation run isolated Ansible/Puppet drift tests' 'make helm-lint  lint local Kubernetes chart' 'make down       stop local stack, retaining named volumes' 'make load       run bounded local load and projection test' 'make demo       run explicitly approved synthetic repair demo'
up:
	@test -n "$$AETHER_OPERATOR_TOKEN" && test -n "$$AETHER_APPROVER_TOKEN" || { echo "Export distinct operator and approver tokens first"; exit 1; }
	AETHER_OTLP_ENDPOINT=http://jaeger:4318/v1/traces $(COMPOSE) up -d --build
down:
	$(COMPOSE) down
logs:
	$(COMPOSE) logs -f
test:
	cd services/control-api && go test -race ./... && go vet ./...
	cd services/event-fabric && go test -mod=readonly -race ./... && go vet -mod=readonly ./...
smoke:
	bash tests/event-fabric-smoke.sh
	bash tests/workflow-smoke.sh
	@export AETHER_EXECUTION_ID="make-evidence-$$(date +%s)-$$$$"; bash tests/execution-smoke.sh && bash tests/cloud-evidence-smoke.sh
	bash tests/observability-smoke.sh
	bash tests/tracing-smoke.sh
automation:
	docker build -t aether-automation:local automation
	docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges aether-automation:local
helm-lint:
	helm lint deploy/helm/aether
	cmp deploy/policy/workflow.rego deploy/helm/aether/files/workflow.rego

load:
	cd services/event-fabric && go run ./cmd/loadtest -requests 40 -concurrency 4

demo:
	cd services/event-fabric && go run -mod=readonly ./cmd/aetherctl demo -approve-lab

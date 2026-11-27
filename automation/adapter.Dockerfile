FROM golang:1.26.11-alpine AS build
WORKDIR /src
COPY services/event-fabric/ .
RUN go mod download && go mod verify && CGO_ENABLED=0 go build -mod=readonly -trimpath -o /fixture ./cmd/fixture
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ansible-core puppet-agent ca-certificates && rm -rf /var/lib/apt/lists/* && useradd --uid 65532 --create-home aether
COPY --from=build /fixture /fixture
COPY automation/ansible /lab/ansible
COPY automation/puppet /lab/puppet
COPY automation/reconcile.sh /lab/reconcile.sh
USER 65532:65532
ENV HOME=/tmp/aether-home ANSIBLE_LOCAL_TEMP=/tmp/ansible-local ANSIBLE_REMOTE_TEMP=/tmp/ansible-remote
ENTRYPOINT ["/fixture"]

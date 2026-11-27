FROM golang:1.26.11-alpine AS build
WORKDIR /src
COPY . .
RUN go mod download && go mod verify && CGO_ENABLED=0 go build -mod=readonly -o /ingest ./cmd/ingest && CGO_ENABLED=0 go build -mod=readonly -o /consumer ./cmd/consumer && CGO_ENABLED=0 go build -mod=readonly -o /relay ./cmd/relay && CGO_ENABLED=0 go build -mod=readonly -o /workflow ./cmd/workflow && CGO_ENABLED=0 go build -mod=readonly -o /executor ./cmd/executor && CGO_ENABLED=0 go build -mod=readonly -o /fixture ./cmd/fixture
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /ingest /ingest
COPY --from=build /consumer /consumer
COPY --from=build /relay /relay
COPY --from=build /workflow /workflow
COPY --from=build /executor /executor
COPY --from=build /fixture /fixture
USER nonroot:nonroot
ENTRYPOINT ["/ingest"]

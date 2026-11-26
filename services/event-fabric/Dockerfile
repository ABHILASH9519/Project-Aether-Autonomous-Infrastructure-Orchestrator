FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -mod=mod -o /ingest ./cmd/ingest && CGO_ENABLED=0 go build -mod=mod -o /consumer ./cmd/consumer
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /ingest /ingest
COPY --from=build /consumer /consumer
USER nonroot:nonroot
ENTRYPOINT ["/ingest"]

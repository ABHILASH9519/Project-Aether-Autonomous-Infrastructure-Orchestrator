# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /aether-control-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /aether-control-api /aether-control-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/aether-control-api"]

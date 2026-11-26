module aether/event-fabric

go 1.23

require github.com/segmentio/kafka-go v0.4.48

require github.com/lib/pq v1.10.9

require github.com/redis/go-redis/v9 v9.7.0

require (
 go.opentelemetry.io/otel v1.34.0
 go.opentelemetry.io/otel/trace v1.34.0
 go.opentelemetry.io/otel/sdk v1.34.0
 go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0
)

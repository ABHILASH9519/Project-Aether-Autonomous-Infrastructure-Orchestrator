package telemetry

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestKafkaTracePropagation(t *testing.T){
	p:=sdktrace.NewTracerProvider()
	defer p.Shutdown(context.Background())
	old:=otel.GetTracerProvider()
	otel.SetTracerProvider(p)
	defer otel.SetTracerProvider(old)
	ctx,root:=p.Tracer("test").Start(context.Background(),"root")
	defer root.End()
	var m kafka.Message
	Inject(ctx,&m)
	_,child:=Consumer(context.Background(),m)
	defer child.End()
	if child.SpanContext().TraceID()!=root.SpanContext().TraceID() || child.SpanContext().SpanID()==root.SpanContext().SpanID(){t.Fatal("trace propagation failed")}
}

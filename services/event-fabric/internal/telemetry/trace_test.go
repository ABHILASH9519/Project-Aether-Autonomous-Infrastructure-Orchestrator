package telemetry

import (
	"context"
	"testing"
	"go.opentelemetry.io/otel/trace"

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

func TestDurableParent(t *testing.T){
	p:=sdktrace.NewTracerProvider()
	defer p.Shutdown(context.Background())
	old:=otel.GetTracerProvider()
	otel.SetTracerProvider(p)
	defer otel.SetTracerProvider(old)
	ctx,root:=p.Tracer("test").Start(context.Background(),"queue producer")
	parent:=Parent(ctx)
	root.End()
	if len(parent)!=55{t.Fatalf("invalid serialized parent length: %d",len(parent))}
	_,child:=Resume(context.Background(),parent,"queue consumer",trace.SpanKindConsumer)
	defer child.End()
	if child.SpanContext().TraceID()!=root.SpanContext().TraceID() || child.SpanContext().SpanID()==root.SpanContext().SpanID(){t.Fatal("durable trace parent lost")}
	_,fresh:=Resume(context.Background(),"invalid","invalid parent",trace.SpanKindConsumer)
	defer fresh.End()
	if !fresh.SpanContext().IsValid() || fresh.SpanContext().TraceID()==root.SpanContext().TraceID(){t.Fatal("invalid parent did not create a fresh trace")}
}

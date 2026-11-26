// Package telemetry wires optional OTLP tracing without including event data or credentials.
package telemetry

import (
	"context"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var propagator = propagation.TraceContext{}

func Init(ctx context.Context,service string)(func(context.Context)error,error){
	endpoint:=os.Getenv("AETHER_OTLP_ENDPOINT")
	if endpoint=="" {return func(context.Context)error{return nil},nil}
	exporter,err:=otlptracehttp.New(ctx,otlptracehttp.WithEndpointURL(endpoint))
	if err!=nil{return nil,err}
	provider:=sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter),sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name",service))),sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())))
	otel.SetTracerProvider(provider)
	return provider.Shutdown,nil
}
func HTTP(next http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.URL.Path!="/api/v1/events" || r.Method!="POST"{next.ServeHTTP(w,r);return}
		ctx:=propagator.Extract(r.Context(),propagation.HeaderCarrier(r.Header))
		ctx,span:=otel.Tracer("aether").Start(ctx,"ingest HTTP",trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		next.ServeHTTP(w,r.WithContext(ctx))
	})
}

func Inject(ctx context.Context,m *kafka.Message){
	c:=propagation.MapCarrier{}
	propagator.Inject(ctx,c)
	for _,key:=range propagator.Fields(){
		if v:=c.Get(key);v!=""{m.Headers=append(m.Headers,kafka.Header{Key:key,Value:[]byte(v)})}
	}
}
func Consumer(ctx context.Context,m kafka.Message)(context.Context,trace.Span){
	c:=propagation.MapCarrier{}
	for _,h:=range m.Headers{
		if h.Key=="traceparent" || h.Key=="tracestate"{c.Set(h.Key,string(h.Value))}
	}
	return otel.Tracer("aether").Start(propagator.Extract(ctx,c),"incident consume",trace.WithSpanKind(trace.SpanKindConsumer))
}

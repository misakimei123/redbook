package ioc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitOTEL() func(ctx context.Context) error {
	r, err := newResource("redbook", "v0.0.1")
	if err != nil {
		panic(err)
	}
	//跨端传递trace，链路元数据
	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)
	traceProvider, err := newTraceProvider(r)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(traceProvider)
	return traceProvider.Shutdown
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion)),
	)
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://192.168.252.128:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)), trace.WithResource(res))
	return tracerProvider, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
}

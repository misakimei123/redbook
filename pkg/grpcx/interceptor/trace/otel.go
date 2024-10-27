package trace

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/misakimei123/redbook/pkg/grpcx/interceptor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type OTELInterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	interceptor.Builder
	serviceName string
}

func NewOTELInterceptorBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator, serviceName string) *OTELInterceptorBuilder {
	return &OTELInterceptorBuilder{tracer: tracer, propagator: propagator, serviceName: serviceName, Builder: *interceptor.NewBuilder()}
}

func (b *OTELInterceptorBuilder) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("ahyang/webook")
	}
	propagator := b.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	attrs := []attribute.KeyValue{semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"), attribute.Key("rpc.component").String("server")}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithAttributes(attrs...), trace.WithSpanKind(trace.SpanKindServer))
		defer func() {
			span.End()
		}()
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			semconv.NetPeerNameKey.String(b.PeerName(ctx)),
			attribute.Key("net.peer.ip").String(b.PeerIP(ctx)),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()
		return handler(ctx, req)
	}
}

func (b *OTELInterceptorBuilder) BuildUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer("ahyang/webook")
	}
	propagator := b.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	attrs := []attribute.KeyValue{semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"), attribute.Key("rpc.component").String("client")}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		newAttrs := append(attrs, semconv.RPCMethodKey.String(method), semconv.NetPeerNameKey.String(b.serviceName))
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(newAttrs...))
		ctx = inject(ctx, propagator)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func extract(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return propagators.Extract(ctx, GrpcHeaderCarrier(md))
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	propagators.Inject(ctx, GrpcHeaderCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

type GrpcHeaderCarrier metadata.MD

func (g GrpcHeaderCarrier) Get(key string) string {
	vals := metadata.MD(g).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (g GrpcHeaderCarrier) Set(key string, value string) {
	metadata.MD(g).Set(key, value)
}

func (g GrpcHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(g))
	for _, key := range keys {
		keys = append(keys, key)
	}
	return keys
}

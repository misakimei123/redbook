package ratelimit

import (
	"context"
	"strings"

	"github.com/misakimei123/redbook/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limit, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			return nil, err
		}
		if limit {
			return nil, status.Error(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !strings.HasPrefix(info.FullMethod, "/UserService") {
			return handler(ctx, req)
		}
		limit, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			return nil, err
		}
		if limit {
			return nil, status.Error(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorServiceV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !strings.HasPrefix(info.FullMethod, "/UserService") {
			return handler(ctx, req)
		}
		limit, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			return nil, err
		}
		if limit {
			ctx = context.WithValue(ctx, "downgrade", true)
		}
		return handler(ctx, req)
	}
}

package logger

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/misakimei123/redbook/pkg/grpcx/interceptor"
	"github.com/misakimei123/redbook/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	l logger.LoggerV1
	interceptor.Builder
}

func (b InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			cost := time.Since(start)
			if rec := recover(); rec != nil {
				switch re := rec.(type) {
				case error:
					err = re
				default:
					err = fmt.Errorf("%v", rec)
				}
				event = "recover"
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				err = status.New(codes.Internal, "panic, err"+err.Error()).Err()
			}
			fields := []logger.Field{
				logger.String("type", "unary"),
				logger.Int64("cost", cost.Microseconds()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_ip", b.PeerIP(ctx)),
			}
			st, _ := status.FromError(err)
			if st != nil {
				fields = append(fields, logger.String("code", st.Code().String()))
				fields = append(fields, logger.String("code_msg", st.Message()))
			}
			b.l.Info("RPC invoke", fields...)
		}()
		return handler(ctx, req)
	}
}

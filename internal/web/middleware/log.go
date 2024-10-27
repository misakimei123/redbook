package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(f func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: f,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}

		method := ctx.Request.Method

		accessLog := AccessLog{
			Path:   path,
			Method: method,
		}

		if l.allowReqBody {
			body, err := ctx.GetRawData()
			if err == nil {
				if len(body) > 2048 {
					accessLog.ReqBody = string(body[:2048])
				} else {
					accessLog.ReqBody = string(body)
				}

				ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			}
		}

		start := time.Now()

		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				al:             &accessLog,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			accessLog.Duration = time.Since(start)
			l.logFn(ctx, accessLog)
		}()

		ctx.Next()
	}
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	RespBody string        `json:"resp_body"`
	State    int           `json:"state"`
	Duration time.Duration `json:"duration"`
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.State = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

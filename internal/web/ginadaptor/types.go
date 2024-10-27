package ginadaptor

import (
	"context"
	"github.com/gin-gonic/gin"
)

type Context interface {
	ConvertCtx(f func(*LogContext)) gin.HandlerFunc
}

type CtxLog struct {
	LogMsg  string `json:"log_msg"`
	Handler string `json:"handler"`
}

type LogContext struct {
	ctxLog *CtxLog
	logFn  func(ctx context.Context, ctxLog *CtxLog)
	*gin.Context
}

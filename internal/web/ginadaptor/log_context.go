package ginadaptor

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/pkg/logger"
)

func getHandlerDetail() string {
	pc, file, no, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		return fmt.Sprintf("handler: %s, file: %s, line no: %d", details.Name(), file, no)
	}
	return ""
}

func (c *LogContext) String(code int, format string, values ...any) {
	c.ctxLog.Handler = getHandlerDetail()
	c.ctxLog.LogMsg = fmt.Sprintf(format, values...)
	c.logFn(c, c.ctxLog)
	c.Context.String(code, format, values)
}

func (c *LogContext) JSON(code int, obj any) {
	c.ctxLog.Handler = getHandlerDetail()
	jsonStr, _ := json.Marshal(obj)
	c.ctxLog.LogMsg = string(jsonStr)
	c.logFn(c, c.ctxLog)
	c.Context.JSON(code, obj)
}

type LogContextBuilder struct {
	l logger.LoggerV1
}

func NewLogContextBuilder(l logger.LoggerV1) Context {
	return &LogContextBuilder{l: l}
}

func (b *LogContextBuilder) ConvertCtx(f func(*LogContext)) gin.HandlerFunc {
	logFn := func(ctx context.Context, ctxLog *CtxLog) {
		jsonStr, _ := json.Marshal(*ctxLog)
		b.l.Info("handler log:", logger.Field{
			Key: "result",
			Val: string(jsonStr),
		})
	}
	return func(c *gin.Context) {
		ctxLog := CtxLog{}
		ctx := &LogContext{
			ctxLog:  &ctxLog,
			logFn:   logFn,
			Context: c,
		}
		f(ctx)
	}
}

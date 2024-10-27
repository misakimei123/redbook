package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/pkg/gormx/connpool"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/migrator"
	"github.com/misakimei123/redbook/pkg/migrator/events"
	"github.com/misakimei123/redbook/pkg/migrator/validator"
	"gorm.io/gorm"
)

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer
	fulls      map[string]func()
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB,
	pool *connpool.DoubleWritePool, l logger.LoggerV1, producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{src: src, dst: dst, pool: pool, l: l, producer: producer,
		cancelFull: func() {
			// 初始的时候，啥也不用做
		},
		cancelIncr: func() {
			// 初始的时候，啥也不用做
		}}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", s.SrcOnly)
	server.POST("/src_first", s.SrcFirst)
	server.POST("/dst_only", s.DstOnly)
	server.POST("/dst_first", s.DstFirst)
	server.POST("/full/start", s.StartFullValidation)
	server.POST("/full/stop", s.StopFullValidation)
	server.POST("/incr/stop", s.StopIncrementValidation)
	server.POST("/incr/start", s.StartIncrementValidation)
}

func (s *Scheduler[T]) SrcOnly(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	err := s.pool.UpdatePattern(connpool.PatternSrcOnly)
	if err != nil {
		s.l.Error("UpdatePattern fail ", logger.Error(err))
		c.JSON(http.StatusOK, Result{Msg: "NOK"})
		return
	}
	c.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, s.l, s.producer, "SRC"), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, s.l, s.producer, "DST"), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Msg: "system fail",
		})
		s.l.Error("newValidator fail", logger.Error(err))
		return
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		cancel()
		err := v.Full().Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	c.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	type Req struct {
		Utime int64 `json:"utime"`
		// 毫秒数
		// json 不能正确处理 time.Duration 类型
		Interval int64 `json:"interval"`
	}
	var req Req
	err := c.Bind(&req)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 501001,
			Msg:  "system error",
		})
		return
	}
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Msg: "system fail",
		})
		return
	}
	v.Incr().Utime(req.Utime).SleepInterval(time.Duration(req.Interval) * time.Millisecond)
	go func() {
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
}

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/pkg/logger"
	"golang.org/x/sync/semaphore"
)

type Executor interface {
	Name() string
	Execute(ctx context.Context, job domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) RegisterLocalFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Execute(ctx context.Context, job domain.Job) error {
	fn, ok := l.funcs[job.Name]
	if !ok {
		return fmt.Errorf("not register local job func  %s ", job.Name)
	}
	return fn(ctx, job)
}

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.JobService
	l         logger.LoggerV1
	executors map[string]Executor
	limiter   *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{svc: svc, l: l, dbTimeout: time.Second,
		limiter: semaphore.NewWeighted(100), executors: make(map[string]Executor)}
}

func (s *Scheduler) RegisterExecutor(executor Executor) {
	s.executors[executor.Name()] = executor
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancelFunc := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.svc.Preempt(dbCtx)
		cancelFunc()
		if err != nil {
			continue
		}

		executor, ok := s.executors[job.Executor]
		if !ok {
			s.l.Error("could not find executor",
				logger.Int64("job id", job.Id),
				logger.String("executor", job.Executor),
			)
			continue
		}
		go func() {
			defer func() {
				s.limiter.Release(1)
				job.CancelFunc()
			}()
			err1 := executor.Execute(ctx, job)
			if err1 != nil {
				s.l.Error("executor fail",
					logger.Error(err),
					logger.Int64("job id", job.Id),
					logger.String("executor", job.Executor),
				)
				return
			}
			err1 = s.svc.ResetNextTime(ctx, job)
			if err1 != nil {
				s.l.Error("reset next time fail",
					logger.Error(err),
					logger.Int64("job id", job.Id),
					logger.String("executor", job.Executor),
				)
			}
		}()
	}
}

package service

import (
	"context"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/pkg/logger"
)

type JobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type CronJobService struct {
	repo            repository.JobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func (c *CronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		c.l.Error("Preempt error", logger.Error(err))
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		rCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		releaseErr := c.repo.Release(rCtx, j.Id)
		if releaseErr != nil {
			c.l.Error("release job fail",
				logger.Error(releaseErr),
				logger.Int64("job id", j.Id))
		}
	}
	return j, nil
}

func NewCronJobService(repo repository.JobRepository, l logger.LoggerV1) JobService {
	return &CronJobService{repo: repo, l: l, refreshInterval: time.Minute}
}

func (c *CronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	return c.repo.UpdateNextTime(ctx, j.Id, j.NextTime())
}

func (c *CronJobService) refresh(jid int64) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	err := c.repo.UpdateUtime(ctx, jid)
	if err != nil {
		c.l.Error("refresh fail", logger.Error(err), logger.Int64("job id", jid))
	}
}

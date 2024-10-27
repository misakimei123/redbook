package repository

import (
	"context"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/dao"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

func NewPreemptJobRepository(dao dao.JobDao) JobRepository {
	return &PreemptJobRepository{dao: dao}
}

type PreemptJobRepository struct {
	dao dao.JobDao
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	return p.dao.UpdateNextTime(ctx, jid, nextTime)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, jid int64) error {
	return p.dao.UpdateUtime(ctx, jid)
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return p.dao.Release(ctx, jid)
}

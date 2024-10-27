package job

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/pkg/distribute/balance"
	"github.com/misakimei123/redbook/pkg/distribute/lock"
	"github.com/misakimei123/redbook/pkg/logger"
)

type RankingJob struct {
	svc       service.RankingService
	name      string
	l         logger.LoggerV1
	client    lock.Lock
	key       string
	ttl       time.Duration
	lock      *bool
	localLock *sync.Mutex
	balancer  balance.LoadBalance[RunNode]
}

func NewRankingJob(svc service.RankingService, l logger.LoggerV1, client lock.Lock, balancer balance.LoadBalance[RunNode]) *RankingJob {
	return &RankingJob{svc: svc, name: "job:ranking", l: l, client: client,
		key: "Lock:ranking", ttl: time.Minute,
		localLock: &sync.Mutex{}, balancer: balancer}
}

func (r *RankingJob) Name() string {
	return r.name
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	dLock := r.lock
	if dLock == nil {
		locked, err := r.client.AcquireLock(r.key, r.ttl)
		if err != nil {
			r.l.Error("acquire lock fail ", logger.Error(err))
			return err
		}
		if !locked {
			r.l.Info("ranking job is locked")
			return errors.New("ranking job is locked")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_ = r.balancer.UpdateLoad(ctx, float64(rand.Intn(100)))
		suited, _ := r.balancer.CurNodeIsSuitable(ctx)
		cancel()
		if !suited {
			return errors.New("cur node is not suited")
		}
		r.lock = &locked
		r.l.Info("got lock")
		go func() {
			er := r.client.AutoRefresh(r.key, r.ttl, r.ttl/2)
			if er != nil {
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), r.ttl)
	defer cancelFunc()
	err := r.svc.Rank(ctx)
	if err != nil {
		r.l.Error("do ranking fail", logger.Error(err))
	}
	return err
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	r.lock = nil
	r.localLock.Unlock()
	er := r.client.ReleaseLock(r.key)
	if er != nil {
		r.l.Error("release lock fail",
			logger.String("key", r.key),
			logger.Error(er))
	}
	return er
}

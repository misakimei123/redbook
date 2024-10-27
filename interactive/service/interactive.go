package service

import (
	"context"

	"github.com/misakimei123/redbook/interactive/domain"
	"github.com/misakimei123/redbook/interactive/repository"

	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=./mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error
	Like(ctx context.Context, like bool, bizStr string, bizId int64, uid int64) error
	Collect(ctx context.Context, bizStr string, bizId, cid, uid int64) error
	Get(ctx context.Context, bizStr string, bizId int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, bizStr string, ids []int64) (map[int64]domain.Interactive, error)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) GetByIds(ctx context.Context, bizStr string, ids []int64) (map[int64]domain.Interactive, error) {
	intras, err := i.repo.GetByIds(ctx, bizStr, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intras))
	for _, intra := range intras {
		res[intra.BizId] = intra
	}
	return res, nil
}

func (i *interactiveService) Get(ctx context.Context, bizStr string, bizId int64, uid int64) (domain.Interactive, error) {
	intra, err := i.repo.Get(ctx, bizStr, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intra.Collected, er = i.repo.Collectd(ctx, bizStr, bizId, uid)
		return er
	})
	eg.Go(func() error {
		var er error
		intra.Liked, er = i.repo.Liked(ctx, bizStr, bizId, uid)
		return er
	})
	err = eg.Wait()
	if err != nil {
		return domain.Interactive{}, err
	}
	intra.Biz = bizStr
	return intra, nil
}

func (i *interactiveService) Collect(ctx context.Context, bizStr string, bizId, cid, uid int64) error {
	return i.repo.AddCollectItem(ctx, bizStr, bizId, cid, uid)
}

func (i *interactiveService) Like(ctx context.Context, like bool, bizStr string, bizId int64, uid int64) error {
	var err error
	if like {
		err = i.repo.IncrLike(ctx, bizStr, bizId, uid)
	} else {
		err = i.repo.DecrLike(ctx, bizStr, bizId, uid)
	}
	return err
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, bizStr, bizId)
}

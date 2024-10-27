package repository

import (
	"context"

	"github.com/misakimei123/redbook/interactive/domain"
	"github.com/misakimei123/redbook/interactive/repository/cache"
	"github.com/misakimei123/redbook/interactive/repository/dao"

	"github.com/ecodeclub/ekit/slice"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error
	IncrLike(ctx context.Context, bizStr string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, bizStr string, bizId int64, uid int64) error
	AddCollectItem(ctx context.Context, bizStr string, bizId, cid, uid int64) error
	Get(ctx context.Context, bizStr string, bizId int64) (domain.Interactive, error)
	Collectd(ctx context.Context, bizStr string, bizId int64, uid int64) (bool, error)
	Liked(ctx context.Context, bizStr string, bizId int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
	GetByIds(ctx context.Context, bizStr string, ids []int64) ([]domain.Interactive, error)
}

func NewCachedInteractiveRepository(dao dao.InteractiveDao, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache}
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, bizStr string, ids []int64) ([]domain.Interactive, error) {
	intras, err := c.dao.GetByIds(ctx, bizStr, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intras, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) Collectd(ctx context.Context, bizStr string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, bizStr, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, bizStr string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, bizStr, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, bizStr string, bizId int64) (domain.Interactive, error) {
	intra, err := c.cache.Get(ctx, bizStr, bizId)
	if err == nil {
		return intra, nil
	} else {
		//TODO:	record log
	}
	interactive, err := c.dao.Get(ctx, bizStr, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intra = c.toDomain(interactive)
	err = c.cache.Set(ctx, bizStr, bizId, intra)
	if err != nil {
		//TODO:	record log
	}
	return intra, nil
}

func (c *CachedInteractiveRepository) AddCollectItem(ctx context.Context, bizStr string, bizId, cid, uid int64) error {
	err := c.dao.InsertCollectBiz(ctx, dao.UserCollectBiz{
		Uid:    uid,
		BizId:  bizId,
		BizStr: bizStr,
		Cid:    cid,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, bizStr, bizId)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, bizStr, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, bizStr, bizId)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, bizStr, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, bizStr, bizId)
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, bizStr, bizId)
	if err != nil {
		return err
	}
	err = c.cache.IncrReadCntIfPresent(ctx, bizStr, bizId)
	if err != nil {
		//	TODO: record log
	}
	return nil
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		return err
	}
	for i, id := range ids {
		er := c.cache.IncrReadCntIfPresent(ctx, bizs[i], id)
		if er != nil {
			//	TODO: record log
		}
	}
	return nil
}

func (c *CachedInteractiveRepository) toDomain(intra dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      intra.BizId,
		ReadCnt:    intra.ReadCnt,
		LikeCnt:    intra.LikeCnt,
		CollectCnt: intra.CollectCnt,
	}
}

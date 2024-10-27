package dao

import (
	"context"
	"errors"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/misakimei123/redbook/pkg/logger"
)

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

var errUnknownPattern = errors.New("unknown double write pattern")

type DoubleWriteDAO struct {
	src     InteractiveDao
	dst     InteractiveDao
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, bizStr, bizId)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, bizStr, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, bizStr, bizId)
		if err != nil {
			d.l.Error("double write dst fail",
				logger.Error(err), logger.Int64("id", bizId), logger.String("biz", bizStr))
		}
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, bizStr, bizId)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, bizStr, bizId)
		if err != nil {
			d.l.Error("double write src fail",
				logger.Error(err), logger.Int64("id", bizId), logger.String("biz", bizStr))
		}
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, bizStr, bizId)
	default:
		return errUnknownPattern
	}
	return nil
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertCollectBiz(ctx context.Context, userCollectBiz UserCollectBiz) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserLikeBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserCollectBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) Get(ctx context.Context, bizStr string, bizId int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.Get(ctx, bizStr, bizId)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, bizStr, bizId)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, bizStr string, ids []int64) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}

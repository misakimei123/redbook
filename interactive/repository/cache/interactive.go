package cache

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
	"time"

	"fmt"

	"github.com/misakimei123/redbook/interactive/domain"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt     string
	ErrKeyNotExist = errors.New("cache key not exist")
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, bizStr string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, bizStr string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, bizStr string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, bizStr string, bizId int64) error
	Get(ctx context.Context, bizStr string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, bizStr string, bizId int64, intra domain.Interactive) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, bizStr string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(bizStr, bizId)}, fieldReadCnt, 1).Err()
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, bizStr string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(bizStr, bizId)}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, bizStr string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(bizStr, bizId)}, fieldLikeCnt, -1).Err()
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, bizStr string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(bizStr, bizId)}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, bizStr string, bizId int64) (domain.Interactive, error) {
	result, err := i.client.HGetAll(ctx, i.key(bizStr, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(result) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	var interactive domain.Interactive
	interactive.BizId = bizId
	interactive.ReadCnt, err = strconv.ParseInt(result[fieldReadCnt], 10, 64)
	if err != nil {
		return domain.Interactive{}, err
	}
	interactive.LikeCnt, err = strconv.ParseInt(result[fieldLikeCnt], 10, 64)
	if err != nil {
		return domain.Interactive{}, err
	}
	interactive.CollectCnt, err = strconv.ParseInt(result[fieldCollectCnt], 10, 64)
	if err != nil {
		return domain.Interactive{}, err
	}
	return interactive, nil

}

func (i *InteractiveRedisCache) Set(ctx context.Context, bizStr string, bizId int64, intra domain.Interactive) error {
	key := i.key(bizStr, bizId)
	err := i.client.HSet(ctx, key, fieldReadCnt, intra.ReadCnt, fieldLikeCnt, intra.LikeCnt, fieldCollectCnt, intra.CollectCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()

}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{client: client}
}

func (i *InteractiveRedisCache) key(bizStr string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", bizStr, bizId)
}

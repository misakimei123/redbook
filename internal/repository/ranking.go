package repository

import (
	"context"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/cache"
	"github.com/misakimei123/redbook/pkg/logger"
)

//go:generate mockgen -source=./ranking.go -package=repomocks -destination=./mocks/ranking.mock.go RankingRepository
type RankingRepository interface {
	GetTopN(ctx context.Context) ([]domain.Article, error)
	SetTopN(ctx context.Context, articles []domain.Article) error
}

type CachedRankingRepository struct {
	cache      cache.RankingCache
	l          logger.LoggerV1
	redisCache cache.RedisRankingCache
	localCache cache.RankingLocalCache
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}

func (c *CachedRankingRepository) SetTopN(ctx context.Context, articles []domain.Article) error {
	return c.cache.Set(ctx, articles)
}

func NewCachedRankingRepository(cache cache.RankingCache, l logger.LoggerV1) RankingRepository {
	return &CachedRankingRepository{cache: cache, l: l}
}

func NewCachedRankingRepositoryV1(rCache cache.RedisRankingCache, lCache cache.RankingLocalCache, l logger.LoggerV1) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: rCache, localCache: lCache, l: l}
}

func (c *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	articles, err := c.localCache.Get(ctx)
	if err == nil {
		return articles, nil
	}
	articles, err = c.redisCache.Get(ctx)
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}
	_ = c.localCache.Set(ctx, articles)
	return articles, nil
}

func (c *CachedRankingRepository) SetTopNV1(ctx context.Context, articles []domain.Article) error {
	_ = c.localCache.Set(ctx, articles)
	return c.redisCache.Set(ctx, articles)
}

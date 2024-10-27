package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RankingCache interface {
	Get(ctx context.Context) ([]domain.Article, error)
	Set(ctx context.Context, arts []domain.Article) error
}

type RedisRankingCache struct {
	client     redis.Cmdable
	expiration time.Duration
	key        string
}

func (r *RedisRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func NewRedisRankingCache(client redis.Cmdable) RankingCache {
	return &RedisRankingCache{client: client, expiration: 3 * time.Minute, key: "Ranking:Article"}
}

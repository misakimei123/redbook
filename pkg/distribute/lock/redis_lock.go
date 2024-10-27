package lock

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisLock(client redis.Cmdable) Lock {
	return &RedisLock{client: client}
}

type RedisLock struct {
	client redis.Cmdable
}

func (r *RedisLock) AcquireLock(key string, ttl time.Duration) (bool, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	ok, err := r.client.SetNX(ctx, key, 1, ttl).Result()
	return ok, err
}

func (r *RedisLock) ReleaseLock(key string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	return r.client.Del(ctx, key).Err()
}

func (r *RedisLock) AutoRefresh(key string, ttl time.Duration, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	ctx := context.Background()
	for {
		select {
		case <-ticker.C:
			err := r.client.Expire(ctx, key, ttl).Err()
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return errors.New("time out")
		}
	}
}

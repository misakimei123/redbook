package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
)

type UserCache interface {
	Get(ctx context.Context, userId int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *RedisUserCache) Get(ctx context.Context, userId int64) (domain.User, error) {
	key := c.key(userId)
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}

	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	duJson, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, c.key(du.Id), duJson, c.expiration).Err()
}

func (c *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

package limiter

import (
	_ "embed"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaScript string

type RedisSlidingWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	// 阈值
	rate int64
}

func (l *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return l.cmd.Eval(ctx, luaScript, []string{key}, l.interval.Milliseconds(), l.rate, time.Now().UnixMilli()).Bool()
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int64) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

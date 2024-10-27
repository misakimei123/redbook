package ioc

import (
	"github.com/misakimei123/redbook/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{Addr:
	//viper.GetString("redis.addr")})
	config.Config.Redis.Addr})
}
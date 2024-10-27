package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RedisLoadBalancer[T any] struct {
	client     redis.Cmdable
	expire     time.Duration
	zKey       string
	curNodeKey string
	loadFunc   func() float64
	l          logger.LoggerV1
}

func NewRedisLoadBalancer[T any](client redis.Cmdable, nodeKey string,
	loadFunc func() float64, l logger.LoggerV1) LoadBalance[T] {
	return &RedisLoadBalancer[T]{client: client, expire: time.Minute,
		zKey: "JOB:NODESET", curNodeKey: nodeKey, loadFunc: loadFunc, l: l}
}

func (r *RedisLoadBalancer[T]) key(eKey string) string {
	return fmt.Sprintf("%s:%s", r.zKey, eKey)
}

func (r *RedisLoadBalancer[T]) Register(ctx context.Context, ele Ele[T], load float64) error {
	val, err := json.Marshal(ele)
	if err != nil {
		return err
	}
	err = r.client.Set(ctx, r.key(ele.Key), val, r.expire).Err()
	if err != nil {
		return err
	}
	err = r.client.ZAdd(ctx, r.zKey, redis.Z{Score: load, Member: val}).Err()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for range ticker.C {
			cx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			_ = r.UpdateLoad(cx, r.loadFunc())
			cancelFunc()
		}
	}()
	return err
}

func (r *RedisLoadBalancer[T]) UpdateLoad(ctx context.Context, load float64) error {
	err := r.client.Expire(ctx, r.key(r.curNodeKey), r.expire).Err()
	if err != nil {
		return err
	}
	val, err := r.client.Get(ctx, r.key(r.curNodeKey)).Result()
	if err != nil {
		return err
	}
	return r.client.ZAdd(ctx, r.zKey, redis.Z{Score: load, Member: string(val)}).Err()
}

func (r *RedisLoadBalancer[T]) CurNodeIsSuitable(ctx context.Context) (bool, error) {
	result, err := r.client.ZRangeWithScores(ctx, r.zKey, 0, -1).Result()
	if err != nil {
		panic(err)
	}
	// 删除超时的节点
	for _, item := range result {
		val, ok := item.Member.(string)
		if !ok {
			continue
		}
		var t Ele[T]
		er := json.Unmarshal([]byte(val), &t)
		if er != nil {
			continue
		}
		exists, er := r.client.Exists(ctx, r.key(t.Key)).Result()
		if er != nil {
			continue
		}
		if exists == 1 {
			continue
		}
		er = r.client.ZRem(ctx, r.zKey, val).Err()
		if er != nil {
			continue
		}
	}

	result, err = r.client.ZRevRangeWithScores(ctx, r.zKey, 0, 0).Result()
	if err != nil {
		return true, nil
	}

	if len(result) > 0 {
		value := result[0].Member.(string)
		var ele Ele[T]
		er := json.Unmarshal([]byte(value), &ele)
		if er != nil {
			return false, er
		}
		r.l.Info("node performance:",
			logger.String("cur node", r.curNodeKey),
			logger.String("best one", value))
		return r.curNodeKey == ele.Key, nil
	} else {
		return true, nil
	}
}

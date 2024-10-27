package ioc

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/misakimei123/redbook/internal/job"
	"github.com/misakimei123/redbook/pkg/distribute/balance"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func InitBalancer(client redis.Cmdable, l logger.LoggerV1) balance.LoadBalance[job.RunNode] {
	key := fmt.Sprintf("NODE:%s", job.NodeId)
	balancer := balance.NewRedisLoadBalancer[job.RunNode](client, key, func() float64 {
		return float64(rand.Intn(100))
	}, l)
	node := job.RunNode{
		Id:          job.NodeId,
		Performance: float64(rand.Intn(100)),
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	err := balancer.Register(ctx, balance.Ele[job.RunNode]{
		Key: key,
		Val: node,
	}, node.Performance)
	if err != nil {
		panic(err)
	}
	cancelFunc()
	return balancer
}

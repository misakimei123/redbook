package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/saramax"
)

type InteractiveReadEventBatchConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	bizStr string
	l      logger.LoggerV1
	vector *prometheus.GaugeVec
}

func NewInteractiveReadEventBatchConsumer(repo repository.InteractiveRepository, client sarama.Client,
	l logger.LoggerV1,
	opts prometheus.GaugeOpts,

) *InteractiveReadEventBatchConsumer {
	vec := prometheus.NewGaugeVec(opts, []string{"topic", "type"})
	prometheus.MustRegister(vec)
	return &InteractiveReadEventBatchConsumer{repo: repo, client: client, l: l, bizStr: "article", vector: vec}
}

func (c *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewBatchHandler[ReadEvent](c.l, c.Consume))
		if er != nil {
			c.l.Error("consume fail", logger.Error(er))
		}
	}()
	return nil
}

func (c *InteractiveReadEventBatchConsumer) Consume(messages []*sarama.ConsumerMessage, events []ReadEvent) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()

	ids := make([]int64, 0, len(events))
	bizs := make([]string, 0, len(events))

	for _, event := range events {
		ids = append(ids, event.Aid)
		bizs = append(bizs, c.bizStr)
	}

	defer func() {
		c.vector.WithLabelValues(TopicReadEvent, "consumer").Add(float64(len(events)))
	}()
	return c.repo.BatchIncrReadCnt(ctx, bizs, ids)
}

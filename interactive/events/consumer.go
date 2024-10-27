package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/saramax"
)

const TopicReadEvent = "article_read"

type ReadEvent struct {
	Aid int64
	Uid int64
}

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	bizStr string
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client,
	l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l, bizStr: "article"}
}

func (c *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewHandler[ReadEvent](c.l, c.Consume))
		if er != nil {
			c.l.Error("consume fail", logger.Error(er))
		}
	}()
	return nil
}

func (c *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	return c.repo.IncrReadCnt(ctx, c.bizStr, event.Aid)
}

package fixer

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/migrator"
	"github.com/misakimei123/redbook/pkg/migrator/events"
	"github.com/misakimei123/redbook/pkg/migrator/fixer"
	"github.com/misakimei123/redbook/pkg/saramax"
	"gorm.io/gorm"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](client sarama.Client, l logger.LoggerV1,
	src *gorm.DB, dst *gorm.DB, topic string) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{client: client, l: l, srcFirst: srcFirst, dstFirst: dstFirst, topic: topic}, nil
}

func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{c.topic},
			saramax.NewHandler[events.InconsistentEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("consume err", logger.Error(err))
		}
	}()
	return err
}

func (c *Consumer[T]) Consume(message *sarama.ConsumerMessage, event events.InconsistentEvent) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	switch event.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, event.ID)
	case "DST":
		return c.dstFirst.Fix(ctx, event.ID)
	}
	return errors.New("unknown direction")
}

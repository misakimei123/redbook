package saramax

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/pkg/logger"
)

type BatchHandler[T any] struct {
	l  logger.LoggerV1
	fn func(messages []*sarama.ConsumerMessage, event []T) error
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := claim.Messages()
	const batchSize = 10

	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		var done = false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时了
				done = true
			case msg, ok := <-messages:
				if !ok {
					cancel()
					return errors.New("chanel closed")
				}
				batch = append(batch, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("unmarshal fail ", logger.Error(err))
					continue
				}
				ts = append(ts, t)
			}
		}
		cancel()
		// 凑够了一批，然后处理
		err := b.fn(batch, ts)
		if err != nil {
			b.l.Error("handle event fail ", logger.Error(err))
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(messages []*sarama.ConsumerMessage, event []T) error) sarama.ConsumerGroupHandler {
	return &BatchHandler[T]{l: l, fn: fn}
}

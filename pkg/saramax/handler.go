package saramax

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/pkg/logger"
)

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, event T) error
	l  logger.LoggerV1
}

// Cleanup implements sarama.ConsumerGroupHandler.
func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := claim.Messages()
	for message := range messages {
		var t T
		err := json.Unmarshal(message.Value, &t)
		if err != nil {
			h.l.Error("unmarshal fail ",
				logger.String("topic", message.Topic),
				logger.Int32("partition", message.Partition),
				logger.Int64("offset", message.Offset),
				logger.Error(err))
		}
		err = h.fn(message, t)
		if err != nil {
			h.l.Error("consume fail ",
				logger.String("topic", message.Topic),
				logger.Int32("partition", message.Partition),
				logger.Int64("offset", message.Offset),
				logger.Error(err))
		}
		session.MarkMessage(message, "")
	}
	return nil
}

// Setup implements sarama.ConsumerGroupHandler.
func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func NewHandler[T any](l logger.LoggerV1, fn func(msg *sarama.ConsumerMessage, event T) error) sarama.ConsumerGroupHandler {
	return &Handler[T]{fn: fn, l: l}
}

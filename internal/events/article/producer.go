package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

const TopicReadEvent = "article_read"

type Producer interface {
	ProduceReadEvent(event ReadEvent) error
}

type ReadEvent struct {
	Aid int64
	Uid int64
}

func NewSaramaSyncProducer(producer sarama.SyncProducer,
	opts prometheus.GaugeOpts,
) Producer {
	vec := prometheus.NewGaugeVec(
		opts, []string{"topic", "result"},
	)
	prometheus.MustRegister(vec)
	return &SaramaSyncProducer{producer: producer, vector: vec}
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
	vector   *prometheus.GaugeVec
}

func (s *SaramaSyncProducer) ProduceReadEvent(event ReadEvent) error {
	val, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})
	if err != nil {
		s.vector.WithLabelValues(TopicReadEvent, err.Error()).Inc()
	} else {
		s.vector.WithLabelValues(TopicReadEvent, "success").Inc()
	}
	return err
}

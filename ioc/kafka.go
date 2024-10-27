package ioc

import (
	"github.com/IBM/sarama"
	events2 "github.com/misakimei123/redbook/interactive/events"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/misakimei123/redbook/internal/events/article"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true
	client, err := sarama.NewClient(cfg.Addr, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitialProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

func InitialSaramaSyncProducer(producer sarama.SyncProducer) article.Producer {
	return article.NewSaramaSyncProducer(producer, prometheus.GaugeOpts{
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "kafka_producer",
	})
}

func InitConsumers(
// consumer *events2.InteractiveReadEventBatchConsumer,
) []events2.Consumer {
	return []events2.Consumer{
		// consumer,
	}
}

func InitialInteractiveReadEventBatchConsumer(repo repository.InteractiveRepository, client sarama.Client,
	l logger.LoggerV1) *events2.InteractiveReadEventBatchConsumer {
	return events2.NewInteractiveReadEventBatchConsumer(repo, client, l, prometheus.GaugeOpts{
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "kafka_consumer",
	})
}

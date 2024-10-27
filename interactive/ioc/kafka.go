package ioc

import (
	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/interactive/events"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/misakimei123/redbook/pkg/logger"

	"github.com/misakimei123/redbook/pkg/migrator/events/fixer"
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

func InitConsumers(consumer *events.InteractiveReadEventBatchConsumer, fixConsumer *fixer.Consumer[dao.Interactive]) []events.Consumer {
	return []events.Consumer{consumer, fixConsumer}
}

func InitialInteractiveReadEventBatchConsumer(repo repository.InteractiveRepository, client sarama.Client,
	l logger.LoggerV1) *events.InteractiveReadEventBatchConsumer {
	return events.NewInteractiveReadEventBatchConsumer(repo, client, l, prometheus.GaugeOpts{
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "kafka_consumer",
	})
}

func InitSaramaSyncProducer(client sarama.Client) sarama.SyncProducer {
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return syncProducer
}

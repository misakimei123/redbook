package startup

import (
	"github.com/IBM/sarama"
	"github.com/misakimei123/redbook/internal/events/article"
	"github.com/prometheus/client_golang/prometheus"
)

func InitSaramaClient() sarama.Client {
	client, err := sarama.NewClient([]string{"127.0.0.1:9092"}, sarama.NewConfig())
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
		Namespace: "ahyang",
		Subsystem: "webook",
		Name:      "kafka_producer",
	})
}

// func InitConsumers(consumer *article.InteractiveReadEventConsumer) []events.Consumer {
// 	return []events.Consumer{consumer}
// }

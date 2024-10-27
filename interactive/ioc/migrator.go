package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/misakimei123/redbook/pkg/gormx/connpool"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/migrator/events"
	"github.com/misakimei123/redbook/pkg/migrator/events/fixer"
	"github.com/misakimei123/redbook/pkg/migrator/scheduler"
)

func InitGinServer(l logger.LoggerV1, src SrcDB, dst DstDB,
	pool *connpool.DoubleWritePool, producer events.Producer) *gin.Engine {
	server := gin.Default()
	group := server.Group("/migrator")
	sch := scheduler.NewScheduler[dao.Interactive](src, dst, pool, l, producer)
	sch.RegisterRoutes(group)
	return server
}

func InitInteractiveProducer(producer sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(producer, "inconsistent_interactive")
}

func InitFixerConsumer(client sarama.Client, l logger.LoggerV1, src SrcDB, dst DstDB) *fixer.Consumer[dao.Interactive] {
	consumer, err := fixer.NewConsumer[dao.Interactive](client, l, src, dst, "inconsistent_interactive")
	if err != nil {
		panic(err)
	}
	return consumer
}

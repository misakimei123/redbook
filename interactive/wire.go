//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/misakimei123/redbook/interactive/grpc"
	"github.com/misakimei123/redbook/interactive/ioc"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/misakimei123/redbook/interactive/repository/cache"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/misakimei123/redbook/interactive/service"
)

var (
	thirdPartySet = wire.NewSet(
		ioc.InitRedis, ioc.InitialLogger,
		ioc.InitSaramaClient, ioc.InitSaramaSyncProducer,
		ioc.InitSrcDB, ioc.InitDstDB, ioc.InitDoubleWritePool, ioc.InitBizDB,
	)
	interactiveSvcSet = wire.NewSet(
		dao.NewInteractiveGormDao,
		cache.NewInteractiveRedisCache,
		repository.NewCachedInteractiveRepository,
		service.NewInteractiveService,
	)
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		ioc.InitConsumers,
		ioc.InitialInteractiveReadEventBatchConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewGrpcxServer,
		ioc.InitGinServer,
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

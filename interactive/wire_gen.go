// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitApp() *App {
	loggerV1 := ioc.InitialLogger()
	srcDB := ioc.InitSrcDB(loggerV1)
	dstDB := ioc.InitDstDB(loggerV1)
	doubleWritePool := ioc.InitDoubleWritePool(srcDB, dstDB, loggerV1)
	db := ioc.InitBizDB(doubleWritePool)
	interactiveDao := dao.NewInteractiveGormDao(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache)
	client := ioc.InitSaramaClient()
	interactiveReadEventBatchConsumer := ioc.InitialInteractiveReadEventBatchConsumer(interactiveRepository, client, loggerV1)
	consumer := ioc.InitFixerConsumer(client, loggerV1, srcDB, dstDB)
	v := ioc.InitConsumers(interactiveReadEventBatchConsumer, consumer)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	server := ioc.NewGrpcxServer(interactiveServiceServer, loggerV1)
	syncProducer := ioc.InitSaramaSyncProducer(client)
	producer := ioc.InitInteractiveProducer(syncProducer)
	engine := ioc.InitGinServer(loggerV1, srcDB, dstDB, doubleWritePool, producer)
	app := &App{
		consumers: v,
		server:    server,
		web:       engine,
	}
	return app
}

// wire.go:

var (
	thirdPartySet     = wire.NewSet(ioc.InitRedis, ioc.InitialLogger, ioc.InitSaramaClient, ioc.InitSaramaSyncProducer, ioc.InitSrcDB, ioc.InitDstDB, ioc.InitDoubleWritePool, ioc.InitBizDB)
	interactiveSvcSet = wire.NewSet(dao.NewInteractiveGormDao, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository, service.NewInteractiveService)
)

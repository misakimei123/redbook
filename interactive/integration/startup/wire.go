//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/misakimei123/redbook/interactive/grpc"
	"github.com/misakimei123/redbook/interactive/repository"
	"github.com/misakimei123/redbook/interactive/repository/cache"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/misakimei123/redbook/interactive/service"
)

var (
	thirdPartySet = wire.NewSet(
		InitRedis, InitDB, InitLogger)
	interactiveSvcSet = wire.NewSet(
		dao.NewInteractiveGormDao,
		cache.NewInteractiveRedisCache,
		repository.NewCachedInteractiveRepository,
		service.NewInteractiveService,
	)
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}

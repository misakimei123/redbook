//go:build wireinject

package main

import (
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/internal/repository/cache"
	"github.com/misakimei123/redbook/internal/repository/cache/code"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/internal/web"
	ginCtx "github.com/misakimei123/redbook/internal/web/ginadaptor"
	ijwt "github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/ioc"

	"github.com/misakimei123/redbook/pkg/distribute/lock"

	"github.com/google/wire"
)

var (
	rankingSvcSet = wire.NewSet(
		cache.NewRedisRankingCache,
		repository.NewCachedRankingRepository,
		service.NewArticleRankingService,
	)
	// interactiveSvcSet = wire.NewSet(
	// 	dao2.NewInteractiveGormDao,
	// 	cache2.NewInteractiveRedisCache,
	// 	repository2.NewCachedInteractiveRepository,
	// 	service2.NewInteractiveService,
	// )
)

func InitWebServer() *App {
	wire.Build(
		ioc.InitialLogger, ioc.InitRedis, ioc.InitDB, ioc.InitSMSService,
		ioc.InitSaramaClient,
		ioc.InitialProducer,
		lock.NewRedisLock,
		ioc.InitEtcdClient,
		dao.NewGormUserDao, dao.NewGormProfileDao,
		dao.NewArticleGormDao,
		cache.NewRedisUserCache, code.NewRedisCodeCache,
		cache.NewArticleRedisCache,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository,
		repository.NewCachedArticleRepository,
		service.NewUserService, service.NewCodeService,
		service.NewArticleService,
		rankingSvcSet,
		ioc.InitBalancer,
		ioc.InitRankingJob,
		ioc.InitJobs,
		// interactiveSvcSet,
		ioc.InitIntrClientV1,
		// ioc.InitialInteractiveReadEventBatchConsumer,
		ioc.InitConsumers,
		ioc.InitialSaramaSyncProducer,
		ijwt.NewRedisJWTHandler,
		ginCtx.NewLogContextBuilder,
		web.NewUserHandler,
		ioc.InitWechatService,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

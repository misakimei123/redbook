//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "github.com/misakimei123/redbook/interactive/repository"
	cache2 "github.com/misakimei123/redbook/interactive/repository/cache"
	dao2 "github.com/misakimei123/redbook/interactive/repository/dao"
	service2 "github.com/misakimei123/redbook/interactive/service"
	"github.com/misakimei123/redbook/internal/job"
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/internal/repository/cache"
	"github.com/misakimei123/redbook/internal/repository/cache/code"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/internal/web"
	"github.com/misakimei123/redbook/internal/web/ginadaptor"
	"github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/ioc"
)

var (
	thirdPartySet = wire.NewSet(
		InitRedis, InitDB, InitLogger, InitSaramaClient, InitialProducer)
	interactiveSvcSet = wire.NewSet(
		dao2.NewInteractiveGormDao,
		cache2.NewInteractiveRedisCache,
		repository2.NewCachedInteractiveRepository,
		service2.NewInteractiveService,
	)

	jobSet = wire.NewSet(
		dao.NewGormJobDao,
		repository.NewPreemptJobRepository,
		service.NewCronJobService,
	)
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		InitialSaramaSyncProducer,
		InitWechatService,
		dao.NewGormUserDao, dao.NewGormProfileDao, dao.NewArticleGormDao,
		cache.NewRedisUserCache, code.NewRedisCodeCache, cache.NewArticleRedisCache, cache.NewRedisRankingCache,
		repository.NewCacheUserRepository, repository.NewCacheCodeRepository, repository.NewCachedArticleRepository,
		repository.NewCachedRankingRepository,
		ioc.InitSMSService,
		service.NewUserService, service.NewCodeService, service.NewArticleService, service.NewArticleRankingService,
		interactiveSvcSet,
		jwt.NewRedisJWTHandler, ginadaptor.NewLogContextBuilder,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		web.NewUserHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		ioc.InitIntrClientV1,
		ioc.InitEtcdClient,
	)
	return gin.Default()
}

func InitArticleHandler(articleDao dao.ArticleDao) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		InitialSaramaSyncProducer,
		dao.NewGormUserDao, dao.NewGormProfileDao,
		cache.NewRedisUserCache, cache.NewArticleRedisCache, cache.NewRedisRankingCache,
		repository.NewCacheUserRepository, repository.NewCachedArticleRepository, repository.NewCachedRankingRepository,
		service.NewArticleService, service.NewArticleRankingService,
		interactiveSvcSet,
		ioc.InitIntrClientV1,
		ioc.InitEtcdClient,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}

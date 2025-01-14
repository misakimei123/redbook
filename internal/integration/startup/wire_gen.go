// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := InitDB()
	userDao := dao.NewGormUserDao(db)
	profileDao := dao.NewGormProfileDao(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCacheUserRepository(userDao, profileDao, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := code.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCacheCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	context := ginadaptor.NewLogContextBuilder(loggerV1)
	userHandler := web.NewUserHandler(userService, codeService, handler, context)
	articleDao := dao.NewArticleGormDao(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDao, articleCache, userRepository)
	client := InitSaramaClient()
	syncProducer := InitialProducer(client)
	producer := InitialSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	clientv3Client := ioc.InitEtcdClient()
	interactiveServiceClient := ioc.InitIntrClientV1(clientv3Client)
	rankingCache := cache.NewRedisRankingCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache, loggerV1)
	rankingService := service.NewArticleRankingService(rankingRepository, articleService, interactiveServiceClient, loggerV1)
	articleHandler := web.NewArticleHandler(articleService, interactiveServiceClient, loggerV1, rankingService)
	wechatService := InitWechatService()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2WechatHandler)
	return engine
}

func InitArticleHandler(articleDao dao.ArticleDao) *web.ArticleHandler {
	cmdable := InitRedis()
	articleCache := cache.NewArticleRedisCache(cmdable)
	db := InitDB()
	userDao := dao.NewGormUserDao(db)
	profileDao := dao.NewGormProfileDao(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCacheUserRepository(userDao, profileDao, userCache)
	articleRepository := repository.NewCachedArticleRepository(articleDao, articleCache, userRepository)
	client := InitSaramaClient()
	syncProducer := InitialProducer(client)
	producer := InitialSaramaSyncProducer(syncProducer)
	loggerV1 := InitLogger()
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	clientv3Client := ioc.InitEtcdClient()
	interactiveServiceClient := ioc.InitIntrClientV1(clientv3Client)
	rankingCache := cache.NewRedisRankingCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache, loggerV1)
	rankingService := service.NewArticleRankingService(rankingRepository, articleService, interactiveServiceClient, loggerV1)
	articleHandler := web.NewArticleHandler(articleService, interactiveServiceClient, loggerV1, rankingService)
	return articleHandler
}

func InitJobScheduler() *job.Scheduler {
	db := InitDB()
	jobDao := dao.NewGormJobDao(db)
	jobRepository := repository.NewPreemptJobRepository(jobDao)
	loggerV1 := InitLogger()
	jobService := service.NewCronJobService(jobRepository, loggerV1)
	scheduler := job.NewScheduler(jobService, loggerV1)
	return scheduler
}

// wire.go:

var (
	thirdPartySet = wire.NewSet(
		InitRedis, InitDB, InitLogger, InitSaramaClient, InitialProducer)
	interactiveSvcSet = wire.NewSet(dao2.NewInteractiveGormDao, cache2.NewInteractiveRedisCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)

	jobSet = wire.NewSet(dao.NewGormJobDao, repository.NewPreemptJobRepository, service.NewCronJobService)
)

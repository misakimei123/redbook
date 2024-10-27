package ioc

import (
	"context"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/web"
	ijwt "github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/internal/web/middleware"
	"github.com/misakimei123/redbook/pkg/ginx/middleware/prometheus"
	"github.com/misakimei123/redbook/pkg/ginx/middleware/ratelimit"
	"github.com/misakimei123/redbook/pkg/limiter"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, articleHdl *web.ArticleHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	wechatHdl.RegisterRotes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, ijwtHdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	prometheusBuilder := &prometheus.Builder{
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "gin_http",
	}
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{ijwt.JWTHttpHeaderKey, ijwt.RefreshTokenHeaderKey},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return false
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 10000)).Build(),
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("request message: ", logger.Field{
				Key: "req",
				Val: al,
			})
		}).AllowReqBody().AllowRespBody().Build(),
		prometheusBuilder.BuildResponseTime(),
		prometheusBuilder.BuildActiveRequest(),
		otelgin.Middleware("redbook"),
		middleware.NewLoginJWTMiddlewareBuilder(redisClient, ijwtHdl).CheckLogin(),
	}
}

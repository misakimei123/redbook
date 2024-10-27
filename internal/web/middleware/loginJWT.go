package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ijwt "github.com/misakimei123/redbook/internal/web/jwt"

	"github.com/redis/go-redis/v9"
)

type LoginJWTMiddlewareBuilder struct {
	cmd    redis.Cmdable
	jwtHdl ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(cmd redis.Cmdable, handler ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{cmd: cmd, jwtHdl: handler}
}

func (l *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/refresh_token" ||
			path == "/users/hello" ||
			path == "/users/loginsms/code/send" ||
			path == "/users/loginsms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}

		uc, err := l.jwtHdl.GetUserClaim(ctx)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = l.jwtHdl.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", uc)
	}
}

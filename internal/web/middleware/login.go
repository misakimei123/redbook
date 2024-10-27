package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	UserId        = "userId"
	UpdateTimeKey = "updateTime"
)

var (
	SessionOption = sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
)

type LoginMiddlewareBuilder struct {
}

func (b LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" || path == "/users/hello" {
			return
		}
		session := sessions.Default(ctx)
		userId := session.Get(UserId)
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		val := session.Get(UpdateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			session.Set(UpdateTimeKey, now)
			session.Set(UserId, userId)
			session.Options(SessionOption)
			err := session.Save()
			if err != nil {
				fmt.Println("check login session save error: ", err)
			}
		}
	}
}

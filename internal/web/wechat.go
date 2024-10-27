package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/internal/service/oauth2/wechat"
	ijwt "github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/internal/web/result"
)

type OAuth2WechatHandler struct {
	svc             wechat.Service
	userSvc         service.UserService
	jwtHandler      ijwt.Handler
	key             []byte
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service, userService service.UserService, jwtHandler ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userService,
		key:             []byte(`sbUOPKLeSMJIwJ4piu5f2kkpHCFPUJPK`),
		stateCookieName: "jwt-state",
		jwtHandler:      jwtHandler,
	}
}

func (o *OAuth2WechatHandler) RegisterRotes(server *gin.Engine) {
	group := server.Group("/oauth2/wechat")
	group.GET("/authurl", o.Auth2URL)
	group.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New()
	url := o.svc.AuthURL(ctx, state)
	err := o.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.Result{
		Data: url,
	})
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetIllegalRequest)
		return
	}
	code := ctx.Query("code")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetAuthCodeError)
		return
	}
	user, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	err = o.jwtHandler.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("can not get cookie %s", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("parse jwt token fail %s", err)
	}

	if state != sc.State {
		return fmt.Errorf("state is not match")
	}
	return nil
}

func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback", "", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

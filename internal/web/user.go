package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/service"
	ginCtx "github.com/misakimei123/redbook/internal/web/ginadaptor"
	ijwt "github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/internal/web/middleware"
	"github.com/misakimei123/redbook/internal/web/result"
)

const (
	emailPattern    = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	passwordPattern = `^(?=.*\d)(?=.*[a-zA-Z])(?=.*[^\da-zA-Z\s]).{8,20}$`
	datePattern     = `^(?:(?!0000)[0-9]{4}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1[0-9]|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[0-9]{2}(?:0[48]|[2468][048]|[13579][26])|(?:0[48]|[2468][048]|[13579][26])00)-02-29)$`
	UserId          = "userId"
	bizLogin        = "login"
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	dateRexExp     *regexp.Regexp
	svc            service.UserService
	code           service.CodeService
	jwtHandler     ijwt.Handler
	logContext     ginCtx.Context
}

func NewUserHandler(svc service.UserService, code service.CodeService, handler ijwt.Handler, logContext ginCtx.Context) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordPattern, regexp.None),
		dateRexExp:     regexp.MustCompile(datePattern, regexp.None),
		svc:            svc,
		code:           code,
		jwtHandler:     handler,
		logContext:     logContext,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/users")
	group.POST("/signup", h.logContext.ConvertCtx(h.SignUp))
	group.POST("/login", h.logContext.ConvertCtx(h.LoginJWT))
	group.POST("/logout", h.logContext.ConvertCtx(h.LogoutJWT))
	group.POST("/edit", h.logContext.ConvertCtx(h.Edit))
	group.GET("/profile", h.logContext.ConvertCtx(h.Profile))
	group.POST("/loginsms/code/send", h.logContext.ConvertCtx(h.SendSMSLoginCode))
	group.POST("/loginsms", h.logContext.ConvertCtx(h.LoginSMS))
	group.GET("/hello", h.hello)
	group.GET("/refresh_token", h.logContext.ConvertCtx(h.RefreshToken))
}

func (h *UserHandler) SignUp(ctx *ginCtx.LogContext) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	if !isEmail {
		ctx.String(http.StatusOK, "email format is not correct")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "password format is not correct")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "password is not same")
		return
	}

	err = h.svc.SignUp(ctx, &domain.User{Password: req.Password, Email: req.Email})
	switch {
	case err == nil:
		ctx.String(http.StatusOK, "%s signup success.", req.Email)
	case errors.Is(err, service.ErrDuplicateEmail):
		ctx.String(http.StatusOK, "duplicate email.")
	default:
		ctx.String(http.StatusOK, "system error")
	}

}

func (h *UserHandler) LoginJWT(ctx *ginCtx.LogContext) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "system error %s", err)
		return
	}
	user, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "invalid user or password")
	case err == nil:
		err := h.jwtHandler.SetLoginToken(ctx.Context, user.Id)
		if err != nil {
			ctx.JSON(http.StatusOK, result.RetSystemError)
			return
		}
		ctx.JSON(http.StatusOK, result.RetLoginSuccess)
	default:
		ctx.JSON(http.StatusOK, result.RetSystemError)
	}
}

func (h *UserHandler) Login(ctx *ginCtx.LogContext) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "system error %s", err)
		return
	}
	user, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "invalid user or password")
	case err == nil:
		session := sessions.Default(ctx.Context)
		session.Set(UserId, user.Id)
		session.Options(middleware.SessionOption)
		err := session.Save()
		if err != nil {
			fmt.Println("session save error: ", err)
			ctx.String(http.StatusOK, "system error")
			return
		}
		ctx.String(http.StatusOK, "Login Success")
	default:
		ctx.String(http.StatusOK, "system error")
	}
}

func (h *UserHandler) Profile(ctx *ginCtx.LogContext) {
	user, err := h.svc.Profile(ctx.Request.Context(),
		GetUserId(ctx),
	)
	if err != nil {
		ctx.String(http.StatusOK, "system error %s", err)
	}
	ctx.JSON(http.StatusOK, gin.H{"Nick": user.Nick, "AboutMe": user.AboutMe, "Birthday": user.Birthday.Format("2006-01-02")})
}

func (h *UserHandler) Edit(ctx *ginCtx.LogContext) {
	type Req struct {
		Nick     string `json:"nick"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "system error %s", err)
		return
	}
	var birthday time.Time
	if len(req.Birthday) != 0 {
		birthday, err = time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			ctx.String(http.StatusOK, "birthday format must be yyyy-mm-dd")
			return
		}
	}

	if len(req.Nick) > 50 {
		ctx.String(http.StatusOK, "nick length should not exceed 50")
		return
	}

	if len(req.AboutMe) > 1024 {
		ctx.String(http.StatusOK, "about me length should not exceed 50")
		return
	}

	err = h.svc.Edit(ctx, domain.User{
		Id:       GetUserId(ctx),
		Nick:     req.Nick,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
	})

	if err != nil {
		ctx.String(http.StatusOK, "system error %s", err)
		return
	}

	ctx.String(http.StatusOK, "edit profile success")

}

func (h *UserHandler) hello(ctx *gin.Context) {
	ctx.String(http.StatusOK, "hello webook")
}

func (h *UserHandler) SendSMSLoginCode(ctx *ginCtx.LogContext) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}

	if req.Phone == "" {
		ctx.JSON(http.StatusOK, result.RetNeedPhoneNumber)
		return
	}

	err = h.code.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, result.RetSuccess)
	case service.ErrCodeVerifyTooMany:
		ctx.JSON(http.StatusOK, result.RetTooFrequent)
	default:
		ctx.JSON(http.StatusOK, result.RetSystemError)
	}
}

func (h *UserHandler) LoginSMS(ctx *ginCtx.LogContext) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ret, err := h.code.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	if !ret {
		ctx.JSON(http.StatusOK, result.RetVerifyFail)
		return
	}
	//ctx.JSON(http.StatusOK, RetSuccess)
	user, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	err = h.jwtHandler.SetLoginToken(ctx.Context, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetLoginSuccess)
}

func (h *UserHandler) RefreshToken(ctx *ginCtx.LogContext) {
	rc, err := h.jwtHandler.GetRefreshClaim(ctx.Context)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.jwtHandler.CheckSession(ctx.Context, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.jwtHandler.SetJWTToken(ctx.Context, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
}

func (h *UserHandler) LogoutJWT(ctx *ginCtx.LogContext) {
	err := h.jwtHandler.ClearToken(ctx.Context)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
	return
}

func GetUserId(ctx *ginCtx.LogContext) int64 {
	//return sessions.Default(ctx).Get(UserId).(int64)
	uc, _ := ctx.MustGet("user").(ijwt.UserClaims)
	return uc.Uid
}

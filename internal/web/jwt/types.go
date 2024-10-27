package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	ClearToken(ctx *gin.Context) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	GetUserClaim(ctx *gin.Context) (UserClaims, error)
	GetRefreshClaim(ctx *gin.Context) (RefreshClaims, error)
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

const (
	JWTHttpHeaderKey      = "x-jwt-token"
	RefreshTokenHeaderKey = "x-refresh-token"
)

package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisJWTHandler struct {
	cmd           redis.Cmdable
	signingMethod jwt.SigningMethod
	refreshKey    []byte
	JWTKey        []byte
	rcExpiration  time.Duration
	tkExpiration  time.Duration
	SsidKeyFmt    string
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header(JWTHttpHeaderKey, "")
	ctx.Header(RefreshTokenHeaderKey, "")
	userClaims := ctx.MustGet("user").(UserClaims)
	return r.cmd.Set(ctx, fmt.Sprintf(r.SsidKeyFmt, userClaims.Ssid), "", r.rcExpiration).Err()
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return r.SetJWTToken(ctx, uid, ssid)
}

func (r *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(r.tkExpiration)),
		},
		UserAgent: ctx.GetHeader("User-Agent"),
	}
	token := jwt.NewWithClaims(r.signingMethod, claims)
	tokenStr, err := token.SignedString(r.JWTKey)
	if err != nil {
		return err
	}
	ctx.Header(JWTHttpHeaderKey, tokenStr)
	return nil
}

func (r *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(r.rcExpiration)),
		}}
	token := jwt.NewWithClaims(r.signingMethod, claims)
	tokenStr, err := token.SignedString(r.refreshKey)
	if err != nil {
		return err
	}
	ctx.Header(RefreshTokenHeaderKey, tokenStr)
	return nil
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := r.cmd.Exists(ctx, fmt.Sprintf(r.SsidKeyFmt, ssid)).Result()
	if err != nil || cnt > 0 {
		return fmt.Errorf("invalid ssid")
	}
	return nil
}

func (r *RedisJWTHandler) GetUserClaim(ctx *gin.Context) (UserClaims, error) {
	tokenStr := r.ExtractToken(ctx)
	var uc UserClaims
	token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
		return r.JWTKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		return uc, fmt.Errorf("invalid token")
	}
	return uc, nil
}

func (r *RedisJWTHandler) GetRefreshClaim(ctx *gin.Context) (RefreshClaims, error) {
	tokenStr := r.ExtractToken(ctx)
	var rc RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return r.JWTKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		return rc, fmt.Errorf("invalid token")
	}
	return rc, nil
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd:           cmd,
		signingMethod: jwt.SigningMethodHS512,
		refreshKey:    []byte(`sbUZPISeSMJIwJ4pfc1AdkkpUWTVHGFT`),
		JWTKey:        []byte(`sbUZPISeSMJIwJ4pfc1AdkkpUWTVHGFT`),
		SsidKeyFmt:    "users:ssid:%s",
		rcExpiration:  time.Hour * 24 * 7,
		tkExpiration:  time.Minute * 30,
	}
}

func (r *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	segs := strings.Split(auth, " ")
	if len(segs) != 2 {
		return ""
	}
	tokenStr := segs[1]
	return tokenStr
}

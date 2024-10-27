package auth

import (
	"context"

	"github.com/misakimei123/redbook/internal/service/sms"

	"github.com/golang-jwt/jwt/v5"
)

type AuthSMSService struct {
	svc    sms.SMSService
	JWTKey string
}

func (s *AuthSMSService) Send(ctx context.Context, tplToken string, args []string, number ...string) error {
	var claims SMSClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.JWTKey, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, claims.Tpl, args, number...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
}

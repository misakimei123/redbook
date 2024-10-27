package ratelimit

import (
	"errors"

	"context"

	"github.com/misakimei123/redbook/internal/service/sms"
	"github.com/misakimei123/redbook/pkg/limiter"
)

var ErrorLimited = errors.New("limited")

type RateLimitSMSService struct {
	svc     sms.SMSService
	limiter limiter.Limiter
	key     string
}

func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	limit, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limit {
		return ErrorLimited
	}
	return r.svc.Send(ctx, tplId, args, number...)
}

func NewRateLimitSMSService(svc sms.SMSService, l limiter.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

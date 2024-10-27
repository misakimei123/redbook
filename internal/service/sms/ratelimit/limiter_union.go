package ratelimit

import (
	"context"

	"github.com/misakimei123/redbook/internal/service/sms"
	"github.com/misakimei123/redbook/pkg/limiter"
)

// RateLimiterUnionSMSService 对外暴露了SMSService和Limiter
type RateLimiterUnionSMSService struct {
	sms.SMSService
	limiter.Limiter
	key string
}

func (r *RateLimiterUnionSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	limit, err := r.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limit {
		return ErrorLimited
	}
	return r.SMSService.Send(ctx, tplId, args, number...)
}

func NewRateLimiterUnionSMSService(svc sms.SMSService, l limiter.Limiter) *RateLimiterUnionSMSService {
	return &RateLimiterUnionSMSService{
		SMSService: svc,
		Limiter:    l,
		key:        "sms-limiter",
	}
}
